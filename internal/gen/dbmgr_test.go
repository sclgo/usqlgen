package gen_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/sclgo/usqlgen/internal/gen"
	"github.com/sclgo/usqlgen/pkg/sclerr"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"

	// drivers
	_ "github.com/mithrandie/csvq-driver"
	_ "modernc.org/sqlite"
)

const sqliteNumInputRows = 100

var sqliteRandomDataQuery = fmt.Sprintf(`
  WITH RECURSIVE
    cte(x, y) AS (
       SELECT 1, "hello"
       UNION ALL
       SELECT x,y 
         FROM cte
        LIMIT %d
  )
SELECT x,y FROM cte;
`, sqliteNumInputRows)

func TestRegisterNewDrivers(t *testing.T) {
	newDrivers := gen.RegisterNewDrivers(sql.Drivers())
	require.Empty(t, newDrivers)
}

func TestSimpleCopyWithInsert_SqliteDest(t *testing.T) {
	sqliteSourceSpec := copyTestSpec{
		driver:          "sqlite",
		dsn:             ":memory:",
		randomDataQuery: sqliteRandomDataQuery,
		numInputRows:    sqliteNumInputRows,
		expectedType:    reflect.TypeFor[int64](),
	}
	csvqSourceSpec := copyTestSpec{
		driver:          "csvq",
		dsn:             ".",
		randomDataQuery: `select 1, "hello" union all select 2, "world"`,
		numInputRows:    2,
		expectedType:    reflect.TypeFor[any](),
	}

	for _, spec := range []copyTestSpec{sqliteSourceSpec, csvqSourceSpec} {
		t.Run(spec.driver, func(t *testing.T) {
			runCopyTests(t, spec)
		})
	}
}

func runCopyTests(t *testing.T, spec copyTestSpec) {
	sourceDb, err := sql.Open(spec.driver, spec.dsn)
	require.NoError(t, err)
	defer sclerr.CloseQuietly(sourceDb)

	copyFunc := gen.SimpleCopyWithInsert(gen.FixedPlaceholder("?"))

	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	defer sclerr.CloseQuietly(db)

	ctx := context.Background()

	_, err = db.ExecContext(ctx, "create table hello(a integer, b varchar)")
	require.NoError(t, err)

	defer func() {
		_, err := db.ExecContext(ctx, "drop table hello")
		require.NoError(t, err)
	}()

	randomDataQuery := spec.randomDataQuery
	numInputRows := spec.numInputRows

	t.Run("scantype is expected", func(t *testing.T) {
		someRows, err := sourceDb.QueryContext(ctx, randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		colTypes, err := someRows.ColumnTypes()
		require.NoError(t, err)
		require.Equal(t, colTypes[0].ScanType(), spec.expectedType)
	})

	t.Run("copy to table with name only", func(t *testing.T) {

		someRows, err := sourceDb.QueryContext(ctx, randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		rowsAdded, err := copyFunc(ctx, db, someRows, "hello")
		require.NoError(t, err)

		require.EqualValues(t, numInputRows, rowsAdded)

		checkHelloColumn(t, db)
	})

	t.Run("copy to table with name and columns", func(t *testing.T) {
		someRows, err := sourceDb.QueryContext(ctx, randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		rowsAdded, err := copyFunc(ctx, db, someRows, "hello(a, b)")
		require.NoError(t, err)

		require.EqualValues(t, numInputRows, rowsAdded)
		checkHelloColumn(t, db)
	})

	t.Run("copy to table with insert", func(t *testing.T) {
		someRows, err := sourceDb.QueryContext(ctx, randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		rowsAdded, err := copyFunc(ctx, db, someRows, "insert into hello(a, b) values (?, ?)")
		require.NoError(t, err)

		require.EqualValues(t, numInputRows, rowsAdded)
		checkHelloColumn(t, db)
	})
}

func checkHelloColumn(t *testing.T, db *sql.DB) {
	var bColumn string
	ctx := context.Background()
	require.NoError(t, db.QueryRowContext(ctx, "select distinct b from hello").Scan(&bColumn))
	require.Equal(t, "hello", bColumn)
}

type copyTestSpec struct {
	driver          string
	dsn             string
	randomDataQuery string
	numInputRows    int
	expectedType    reflect.Type
}
