package gen

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ansel1/merry/v2"
	"github.com/samber/lo"
	"github.com/xo/dburl"
	"reflect"
	"strings"
)

// This file is copied as in the generated usql wrapper.
// It must be self-contained. TODO move to dedicated package so self-containment can be enforced.

// For now, we avoid depending on xo/usql, only on xo/dburl, to keep our dep tree small.
// This may change in the future.

func FindNew(current []string, original []string) []string {
	diff, _ := lo.Difference(current, original)
	return diff
}

// RegisterNewDrivers registers in xo/dburl all database/sql.Drivers()
// that are not present in provided existing list.
func RegisterNewDrivers(existing []string) []string {
	newDrivers := FindNew(sql.Drivers(), existing)
	for _, driver := range newDrivers {
		dburl.Unregister(driver)
		if len(driver) > 2 {
			dburl.Unregister(driver[:2])
		}
		dburl.Register(GetScheme(driver))
	}
	return newDrivers
}

func GetScheme(driver string) dburl.Scheme {
	return dburl.Scheme{
		Driver:    driver,
		Generator: dburl.GenOpaque,
		Opaque:    true,
	}
}

func FixedPlaceholder(placeholder string) func(int) string {
	return func(n int) string {
		return placeholder
	}
}

// SimpleCopyWithInsert builds a copy handler based on insert.
func SimpleCopyWithInsert(placeholder func(n int) string) func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
	return func(ctx context.Context, db *sql.DB, rows *sql.Rows, table string) (int64, error) {
		columns, err := rows.Columns()
		if err != nil {
			return 0, merry.Errorf("failed to fetch source rows columns: %w", err)
		}
		clen := len(columns)
		query := table
		if !strings.HasPrefix(strings.ToLower(query), "insert into") {
			leftParen := strings.IndexRune(table, '(')
			if leftParen == -1 {
				colRows, err := db.QueryContext(ctx, "SELECT * FROM "+table+" WHERE 1=0")
				if err != nil {
					return 0, merry.Errorf("failed to execute query to determine target table columns: %w", err)
				}
				defer colRows.Close()
				columns, err := colRows.Columns()
				if err != nil {
					return 0, merry.Errorf("failed to fetch target table columns: %w", err)
				}
				table += "(" + strings.Join(columns, ", ") + ")"
			}
			// TODO if the db supports multiple rows per insert, create batches of 100 rows
			placeholders := make([]string, clen)
			for i := 0; i < clen; i++ {
				placeholders[i] = placeholder(i + 1)
			}
			query = "INSERT INTO " + table + " VALUES (" + strings.Join(placeholders, ", ") + ")"
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return 0, merry.Errorf("failed to begin transaction: %w", err)
		}
		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return 0, merry.Errorf("failed to prepare insert query: %w", err)
		}
		defer stmt.Close()
		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			return 0, merry.Errorf("failed to fetch source column types: %w", err)
		}
		values := make([]interface{}, clen)
		valueRefs := make([]reflect.Value, clen)
		actuals := make([]interface{}, clen)

		for i := 0; i < len(columnTypes); i++ {
			valueRefs[i] = reflect.New(columnTypes[i].ScanType())
			values[i] = valueRefs[i].Interface()
		}

		var n int64
		for rows.Next() {
			err = rows.Scan(values...)
			if err != nil {
				return n, merry.Wrap(fmt.Errorf("failed to scan row: %w", err))
			}

			for i := range values {
				actuals[i] = valueRefs[i].Elem().Interface()
			}
			res, err := stmt.ExecContext(ctx, actuals...)
			if err != nil {
				return n, merry.Wrap(fmt.Errorf("failed to exec insert: %w", err))
			}
			rn, err := res.RowsAffected()
			if err != nil {
				return n, merry.Wrap(fmt.Errorf("failed to check rows affected: %w", err))
			}
			n += rn
		}
		// TODO if using batches, flush the last batch,
		// TODO prepare another statement and count remaining rows
		err = tx.Commit()
		if err != nil {
			return n, merry.Wrap(fmt.Errorf("failed to commit transaction: %w", err))
		}
		return n, merry.Wrap(rows.Err())
	}
}
