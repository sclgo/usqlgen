package gen_test

import (
	"context"
	"database/sql"
	"errors"
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

type copyTestSpec struct {
	driver          string
	dsn             string
	randomDataQuery string
	numInputRows    int
	expectedType    reflect.Type
}

var sqliteSourceSpec = copyTestSpec{
	driver:          "sqlite",
	dsn:             ":memory:",
	randomDataQuery: sqliteRandomDataQuery,
	numInputRows:    sqliteNumInputRows,
	expectedType:    reflect.TypeFor[int64](),
}

var csvqSourceSpec = copyTestSpec{
	driver:          "csvq",
	dsn:             ".",
	randomDataQuery: `select 1, "hello" union all select 2, "world"`,
	numInputRows:    2,
	expectedType:    reflect.TypeFor[any](),
}

func TestRegisterNewDrivers(t *testing.T) {
	newDrivers := gen.RegisterNewDrivers(sql.Drivers())
	require.Empty(t, newDrivers)
}

func TestSimpleCopyWithInsert_SqliteDest(t *testing.T) {
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

	copyFunc := gen.BuildSimpleCopy(gen.FixedPlaceholder("?"))

	ctx := context.Background()
	db, cleanup := prepareTargetDb(t)
	defer cleanup()

	t.Run("scantype is expected", func(t *testing.T) {
		someRows, err := sourceDb.QueryContext(ctx, spec.randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		colTypes, err := someRows.ColumnTypes()
		require.NoError(t, err)
		require.Equal(t, colTypes[0].ScanType(), spec.expectedType)
	})

	t.Run("copy to table with name only", func(t *testing.T) {

		someRows, err := sourceDb.QueryContext(ctx, spec.randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		rowsAdded, err := copyFunc(ctx, db, someRows, "hello")
		require.NoError(t, err)

		require.EqualValues(t, spec.numInputRows, rowsAdded)

		checkHelloColumn(t, db)
	})

	t.Run("copy to table with name and columns", func(t *testing.T) {
		someRows, err := sourceDb.QueryContext(ctx, spec.randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		rowsAdded, err := copyFunc(ctx, db, someRows, "hello(a, b)")
		require.NoError(t, err)

		require.EqualValues(t, spec.numInputRows, rowsAdded)
		checkHelloColumn(t, db)
	})

	t.Run("copy to table with insert", func(t *testing.T) {
		someRows, err := sourceDb.QueryContext(ctx, spec.randomDataQuery)
		require.NoError(t, err)
		defer sclerr.CloseQuietly(someRows)

		rowsAdded, err := copyFunc(ctx, db, someRows, "insert into hello(a, b) values (?, ?)")
		require.NoError(t, err)

		require.EqualValues(t, spec.numInputRows, rowsAdded)
		checkHelloColumn(t, db)
	})
}

func prepareTargetDb(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)

	_, err = db.Exec("create table hello(a integer, b varchar)")
	require.NoError(t, err)
	return db, func() {
		_, err := db.Exec("drop table if exists hello")
		require.NoError(t, err)
		sclerr.CloseQuietly(db)
	}
}

func checkHelloColumn(t *testing.T, db *sql.DB) {
	var bColumn string
	ctx := context.Background()
	require.NoError(t, db.QueryRowContext(ctx, "select distinct b from hello").Scan(&bColumn))
	require.Equal(t, "hello", bColumn)
}

func TestSimpleCopyWithInsert_EdgeCases(t *testing.T) {
	spec := csvqSourceSpec
	sourceDb, err := sql.Open(spec.driver, spec.dsn)
	require.NoError(t, err)
	defer sclerr.CloseQuietly(sourceDb)

	targetDb, cleanup := prepareTargetDb(t)
	defer cleanup()

	db := testDb{
		DB:                targetDb,
		blockTransactions: true,
	}

	ctx := context.Background()

	someRows, err := sourceDb.QueryContext(ctx, spec.randomDataQuery)
	require.NoError(t, err)
	defer sclerr.CloseQuietly(someRows)

	rowsAdded, err := gen.SimpleCopyWithInsert(
		ctx,
		db,
		someRows,
		"insert into hello(a, b) values (?, ?)",
		gen.FixedPlaceholder("?"))
	require.NoError(t, err)

	require.EqualValues(t, spec.numInputRows, rowsAdded)
	checkHelloColumn(t, targetDb)
}

type testDb struct {
	*sql.DB

	blockTransactions bool
}

func (d testDb) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	if d.blockTransactions {
		return nil, errors.New("transactions are not supported")
	}
	return d.DB.BeginTx(ctx, opts)
}
