package gen

import (
	"context"
	"database/sql"
	"fmt"
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

var SemicolonEndRE = regexp.MustCompile(`;?\s*$`)

func FindNew(current []string, original []string) []string {
	return slices.DeleteFunc(current, func(s string) bool {
		return slices.Contains(original, s)
	})
}

// RegisterNewDrivers registers in xo/dburl all database/sql.Drivers()
// that are not present in provided existing list.
func RegisterNewDrivers(existing []string) []string {
	newDrivers := FindNew(sql.Drivers(), existing)
	for _, driver := range newDrivers {
		dburl.Unregister(driver)

		// xo/dburl registers a 2 char alias of all driver names longer than 2 chars
		if len(driver) > 2 {
			dburl.Unregister(driver[:2])
		}

		dburl.Register(GetScheme(driver))
	}
	return newDrivers
}

func GetScheme(driver string) dburl.Scheme {
	return dburl.Scheme{
		Driver: driver,
		Generator: func(u *dburl.URL) (string, string, error) {
			// same as dburl.GenOpaque but accepts empty opaque part.
			// Empty DSN is accepted by some DB drivers like go-duckdb.
			return u.Opaque + genQueryOptions(u.Query()), "", nil
		},
		Opaque: true,
	}
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
// It is similar to usql defaults implementation, but it tries to adjust to some runtime errors
// by trying alternative database features. usql never does that because the usql copy implementation
// is always adapted to the specific database.
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
			// Can't use sclerr since dbmgr is standalone.
			defer colRows.Close()
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
	defer stmt.Close()
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
		defer finStmt.Close()
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
