package gen

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/xo/dburl"
)

// This file is copied as is in the generated usql wrapper.
// It should not depend on other usqlgen packages, public or internal.
// TODO Enforce this possibly with golangci-lint.

// For now, we also avoid depending on xo/usql, only on xo/dburl, to keep our dep tree small.
// This may change in the future.
// Avoid depending on libraries, not already used in usql.

// SemicolonEndRE is used in driver.Process in main.go.tpl
var SemicolonEndRE = regexp.MustCompile(`;?\s*$`)

type set map[string]bool

func findNew(current []string, original set) []string {
	return slices.DeleteFunc(current, func(s string) bool {
		_, ok := original[s]
		return ok
	})
}

// expands expands the set of the driver names to the set of all names and their aliases
func expand(drivers []string) set {
	m := make(map[string]bool, len(drivers))
	for _, drv := range drivers {
		for _, item := range dburl.Protocols(drv) {
			m[item] = true
		}
	}
	return m
}

// RegisterNewDrivers registers in xo/dburl all database/sql.Drivers()
// that are not present in provided existing list.
func RegisterNewDrivers(existing []string) []string {
	existingAll := expand(existing)
	newDrivers := findNew(sql.Drivers(), existingAll)
	if newDrivers == nil {
		log.Printf("Did not find new drivers despite new imports. " +
			"The import path may be incomplete or the imported driver name clashes with existing drivers or their aliases. " +
			"If a clash is suspected, try adding '-- -tags no_xxx' to the usqlgen command-line, where xxx is a DB tag from usql docs.")
	}

	for _, driver := range newDrivers {
		// We have validated that the schemes we are unregistering are not from linked drivers.
		dburl.Unregister(driver)
		scheme := getScheme(driver, existingAll)
		dburl.Unregister(scheme.Aliases[0])

		dburl.Register(scheme)
	}
	return newDrivers
}

func getScheme(driver string, existing set) dburl.Scheme {
	return dburl.Scheme{
		Driver: driver,
		Generator: func(u *dburl.URL) (string, string, error) {
			// same as dburl.GenOpaque but accepts empty opaque part.
			// Empty DSN is accepted by some DB drivers like go-duckdb.
			return u.Opaque + genQueryOptions(u.Query()), "", nil
		},
		Opaque: true,
		// If we don't generate a unique short (2 char) alias, xo/dburl creates
		// it from the first 2 chars of the name which might overwrite a built-in one.
		Aliases: getUniqueShortAlias(driver, existing),
	}
}

// getUniqueShortAlias generates a short (2 char) alias which doesn't repeat any alias in the given set
// The given set is expected to be the result of expand(usql/drivers.Available() -> keys)
// The result may still overwrite an alias of a schemes which is in dburl.BaseSchemes but this is
// fine since this scheme is for a driver which is not in usql/drivers.Available().
func getUniqueShortAlias(driver string, existing set) []string {
	if len(driver) <= 2 {
		return nil
	}
	for _, a := range driver {
		for _, b := range driver[1:] + "0123456789" {
			shortAlias := string(a) + string(b)
			if _, ok := existing[shortAlias]; !ok {
				return []string{shortAlias}
			}
		}
	}
	return nil // should be impossible, because digits are not used in dburl aliases.
}

// genQueryOptions generates standard query options.
func genQueryOptions(q url.Values) string {
	if s := q.Encode(); s != "" {
		return "?" + s
	}
	return ""
}

func FixedPlaceholder(placeholder string) func(int) string {
	return func(n int) string {
		return placeholder
	}
}

// DbWriter is the common subset between *sql.DB and *sql.Tx used by the main loop of SimpleCopyWithInsert
type DbWriter interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// DB is the subset of *sql.DB used by SimpleCopyWithInsert
type DB interface {
	DbWriter
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

// BuildSimpleCopy builds a copy handler based on insert.
// The result func matches the signature required by github.com/xo/usql/drivers.Driver.Copy
// placeholder parameter is currently hardcoded on main.go.tpl but may be configurable in the future.
func BuildSimpleCopy(placeholder func(n int) string) func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
	return func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
		return SimpleCopyWithInsert(ctx, db, rows, table, 10, placeholder)
	}
}

// SimpleCopyWithInsert implements usql \copy
// It is similar to the usql default implementation, but it tries to handle runtime errors
// by trying fallback options (e.g. no transaction if BeginTx fails).
// usql never does that because its copy implementation is always adapted to the specific database.
func SimpleCopyWithInsert(ctx context.Context, db DB, rows *sql.Rows, table string, batchSize int, placeholder func(n int) string) (int64, error) {
	columns, err := rows.Columns()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch source rows columns: %w", err)
	}
	clen := len(columns)
	query := table
	if !strings.HasPrefix(strings.ToLower(query), "insert into") {
		leftParen := strings.IndexRune(table, '(')
		if leftParen == -1 {
			colRows, err := db.QueryContext(ctx, "SELECT * FROM "+table+" WHERE 1=0")
			if err != nil {
				return 0, fmt.Errorf("failed to execute query to determine target table columns: %w", err)
			}
			defer closeQuietly(colRows)
			columns, err := colRows.Columns()
			if err != nil {
				return 0, fmt.Errorf("failed to fetch target table columns: %w", err)
			}
			table += "(" + strings.Join(columns, ", ") + ")"
		}
		query = makeQuery(clen, batchSize, table, placeholder)
	} else {
		batchSize = 1 // no batching
	}
	var wrt DbWriter
	wrt, err = db.BeginTx(ctx, nil)
	if err != nil {
		fmt.Printf("Failed to begin transaction. Falling back to non-transactional copy: %s\n", err)
		wrt = db
	}
	stmt, err := wrt.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert query: %w", err)
	}
	defer closeQuietly(stmt)
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return 0, fmt.Errorf("failed to fetch source column types: %w", err)
	}
	values := make([]interface{}, clen)
	valueRefs := make([]reflect.Value, clen)
	actuals := make([]interface{}, 0, clen*batchSize)

	for i := 0; i < len(columnTypes); i++ {
		valueRefs[i] = reflect.New(columnTypes[i].ScanType())
		values[i] = valueRefs[i].Interface()
	}

	rowsAffectedSupported := true
	var n int64

	for rows.Next() {
		err = rows.Scan(values...)
		if err != nil {
			return n, fmt.Errorf("failed to scan row: %w", err)
		}

		for i := range values {
			actuals = append(actuals, valueRefs[i].Elem().Interface())
		}

		if len(actuals) < batchSize*clen {
			continue
		}

		rn, err := writeActuals(ctx, stmt, actuals, &rowsAffectedSupported)
		if err != nil {
			return n, err
		}
		n += rn
		actuals = actuals[:0] // truncate but keep underlying array size
	}

	if len(actuals) > 0 {
		finStmt, err := wrt.PrepareContext(ctx, makeQuery(clen, len(actuals)/clen, table, placeholder))
		if err != nil {
			return 0, fmt.Errorf("failed to prepare insert query: %w", err)
		}
		defer closeQuietly(finStmt)
		rn, err := writeActuals(ctx, finStmt, actuals, &rowsAffectedSupported)
		if err != nil {
			return n, err
		}
		n += rn
	}

	if tx, ok := wrt.(*sql.Tx); ok {
		err = tx.Commit()
	}

	if err != nil {
		return n, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return n, rows.Err()
}

func closeQuietly(c io.Closer) {
	// Avoid gorich/helperr dependency
	_ = c.Close()
}

func writeActuals(ctx context.Context, stmt *sql.Stmt, actuals []interface{}, rowsAffectedSupported *bool) (int64, error) {
	res, err := stmt.ExecContext(ctx, actuals...)
	if err != nil {
		return 0, fmt.Errorf("failed to exec insert: %w", err)
	}

	var rn int64
	if *rowsAffectedSupported {
		rn, err = res.RowsAffected()
	}
	if err != nil {
		if *rowsAffectedSupported {
			fmt.Printf("Failed to retrieve rowsAffected. Assuming not supported by driver: %s\n", err)
			*rowsAffectedSupported = false
			rn = 0
		}
	}
	return rn, nil
}

func makeQuery(clen int, rows int, tableSpec string, placeholder func(n int) string) string {
	query := "INSERT INTO " + tableSpec + " VALUES "
	placeholders := make([]string, clen)
	for i := 0; i < rows; i++ {
		for j := 0; j < clen; j++ {
			placeholders[j] = placeholder(i*clen + j + 1)
		}
		query += "(" + strings.Join(placeholders, ", ") + ")"
		if i < rows-1 {
			query += ", "
		}
	}
	return query
}
