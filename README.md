# usqlgen

`usqlgen` creates custom distributions of [github.com/xo/usql](https://github.com/xo/usql) - 
a universal single-binary SQL CLI, built by the [github.com/xo](https://github.com/xo) team, and 
inspired by [psql](https://www.postgresql.org/docs/current/app-psql.html).

`usql` is great because it is a multi-platform, multi-database SQL client in a single binary. 
Learn more about it from its [README](https://github.com/xo/usql#readme).

`usqlgen` builds on usql's extensibility to allow including arbitrary drivers and other customizations,
without needing to fork.

![Tests](https://github.com/sclgo/usqlgen/actions/workflows/go.yml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/sclgo/usqlgen)](https://goreportcard.com/report/github.com/sclgo/usqlgen)

## When to use

In most cases, using `usql` directly is best. `usqlgen` helps when you want to avoid editing `usql` code and:

- you want to use a database driver which is not publicly available or is under active development
- you want to use alternative driver for a database supported by `usql`. 
- you want to use a different version or a fork of a bundled driver
- you are working with a niche database which is not supported by `usql` yet
  - consider contributing the support to `usql` at some point
 
The Examples section details those usecases.

`usqlgen` is itself inspired by the 
[OpenTelemetry Collector builder](https://opentelemetry.io/docs/collector/custom-collector/).

> [!IMPORTANT]
> [usql](https://github.com/xo/usql) authors are aware of this project but support
> only [their regular releases](https://github.com/xo/usql?tab=readme-ov-file#installing).
> Issues that appear on builds produced by `usqlgen` should be reported on this repository.
> Report issues on https://github.com/xo/usql/issues only if you can reproduce them on a regular release.

## Installing

Install `usqlgen` with Go 1.21+:

```shell
go install github.com/sclgo/usqlgen@latest
```

You can also run usqlgen without installing:

```shell
go run github.com/sclgo/usqlgen@latest
```

If you don't have Go 1.21+, you can run `usqlgen` with Docker:

```shell
docker run --rm golang \
  go run github.com/sclgo/usqlgen@latest build ...add parameters here... -o - > ./usql

chmod +x ./usql
```

Ensure that the Docker image matches the OS you are using - Windows container for Windows, Alpine for Alpine Linux, etc.
The default golang image works for Debian-based Linux distributions like Ubuntu.

## Quickstart

To install `usql` with support for an additional driver, first review your driver documentation
to find the Go driver name, DSN ([Data Source Name](https://pkg.go.dev/database/sql#Open)) format and package that needs to be imported to install the
driver. Let's take for example, [MonetDB](https://github.com/MonetDB/MonetDB-Go#readme),
which is not in `usql` yet. The docs state that the package that needs to be imported is
`github.com/MonetDB/MonetDB-Go/v2`. We use this command to build `usql`:

```shell
usqlgen build --import "github.com/MonetDB/MonetDB-Go/v2"
```

This creates `usql` executable in the current directory with its default built-in drivers 
together with the driver for MonetDB.
The additional driver is registered using a side-effect import (aka [blank import](https://go.dev/doc/effective_go#blank_import))
of the package in the `--import` parameter. The respective module is automatically
determined but can also be specified explicitly with `--get`.

Unlike built-in databases, the `usql` DB URL (connection string) for the new database 
is in the form `driverName:dsn`. For example, [MonetDB docs](https://github.com/MonetDB/MonetDB-Go#readme)
state that the driver name is `monetdb` and the DSN format is `username:password@hostname:50000/database`.
So to connect to MonetDB with the binary we just built, run:

```shell

# The command below is configured to connect to a local MonetDB started like this:
# docker run -p 50000:50000 -e MDB_DB_ADMIN_PASS=password monetdb/monetdb:latest

./usql monetdb:monetdb:password@localhost:50000/monetdb -c "select 'Hello World'"
```

Above `monetdb` is repeated in the beginning of the DB URL because it is both the Go driver name,
and the admin username in the beginning of the DSN.

You can try the same with databases or data engines like
[rqlite](https://github.com/rqlite/gorqlite?tab=readme-ov-file#driver-for-databasesql),
[Dremio or Apache Drill](https://github.com/factset/go-drill), etc.

`usqlgen` also allows you to use alternative drivers of supported databases. Examples include:

- [github.com/microsoft/gocosmos](https://github.com/microsoft/gocosmos) - an official mirror of the unofficial driver
  included in `usql`
- [github.com/yugabyte/pgx](https://github.com/yugabyte/pgx) - Postgres pgx variant by Yugabyte with cluster-aware load balancing
- [github.com/mailru/go-clickhouse/v2](https://github.com/mailru/go-clickhouse) - HTTP-only alternative of the built-in Clickhouse driver
- [github.com/sclgo/adbcduck-go](https://github.com/sclgo/adbcduck-go) - alternative driver for DuckDB that
  uses [its ADBC C API](https://duckdb.org/docs/clients/adbc)

For more options, see `usqlgen --help` or review the examples below.

## Limitations

Most `usql` [backslash (meta) commands](https://github.com/xo/usql?tab=readme-ov-file#backslash-commands) work 
with new drivers added with `--import`, including 
[cross-database `\copy`](https://github.com/xo/usql?tab=readme-ov-file#copying-between-databases).
Informational commands and autocomplete won't work though - [for now](https://github.com/sclgo/usqlgen/issues/50).

`usql` requires that connection strings are valid URIs or URLs, at least according to the Go `net/url` parsing algorithm.
If you get an error that parameter in the form `driverName:dsn` can't be parsed as a URL,
start `usql` without a connection string - in interactive mode or with the `-f` parameter. `usql` will start not connected to a DB.
Then use the `\connect` command with two arguments driverName and dsn. In the `monetdb` example above, that would be:
`\connect monetdb monetdb:password@localhost:50000/monetdb`.

## CGO usage

`usqlgen build` and `usqlgen install` don't need [CGO](https://pkg.go.dev/cmd/cgo) to compile
the generated `usql` codebase. You only need CGO if you:

- import drivers that use CGO e.g. `--import github.com/sclgo/adbcduck-go`. The driver documentation should mention such
  usage.
- add drivers that use CGO with tags e.g. `-- -tags duckdb` .

When CGO is not available, `usqlgen` modifies the default "base" driver set,
replacing `sqlite3` (that requires CGO) with `moderncsqlite` (that doesn't).
In that case, using the `sqlite3` scheme will run `moderncsqlite` underneath.
You can opt out of that behavior either:

- by adding `--dboptions keepcgo`
- by forcing Go to assume CGO is available, by setting environment variable `CGO_ENABLED=1`.
- by adding tags to explicitly include or exclude relevant drivers.

All in all, `usqlgen` and the `usql` binaries it produces are very portable - more portable than regular `xo/usql` -
as long as `usql` is generated and built in the same environment, it is expected to run in.

## Examples

### Installing the customized `usql`

Use `usqlgen install ...` to install the customized `usql` to `GOPATH/bin` which is
typically on the search path. This runs `go install` internally.

```shell
usqlgen install --import "github.com/MonetDB/MonetDB-Go/v2"
usql -c '\drivers' | grep monetdb
# prints
#   monetdb [mo]
```

### Adding compilation options

`usqlgen build` and `usqlgen install` call `go build` and `go install` respectively.
You can pass options directly to the `go` commands after the `--` parameter.
For example, the following command supplies go build tag no_base which removes
all built-in `usql` drivers so only the custom one we add remains:

```shell
usqlgen build --import "github.com/MonetDB/MonetDB-Go/v2" -- -tags no_base
./usql -c '\drivers'
# prints only a single driver
```

In this case, the binary will be smaller and faster to build.

Review <https://github.com/xo/usql?tab=readme-ov-file#building> for build tags, supported
by `usql` and the documentation of `go build` and `go install` for other options.

Go environment variables like `GOPRIVATE` or `CGO_ENABLED` affect the compilation
as usual. For example, `GOPRIVATE` allows you to compile `usql` with drivers which
are not publicly available.

### Using a driver fork

`usqlgen` can build `usql` with a `replace` directive so that you can use a
SQL driver fork while keeping the `usql` configuration for the target database.
Information commands, schema exploration, and autocomplete will continue to work
if the fork remains compatible enough with the original.

For example, [github.com/dlapko/go-mssqldb](https://github.com/dlapko/go-mssqldb)
is a fork of [github.com/microsoft/go-mssqldb](https://github.com/microsoft/go-mssqldb)
with several fixes at the time of writing.

To build `usql` with the fork, run:

```shell
usqlgen build --replace "github.com/microsoft/go-mssqldb=github.com/dlapko/go-mssqldb@main"
```

Note that this works only with forks that keep the original module name - 
in this case `github.com/microsoft/go-mssqldb` - in their 
[go.mod](https://github.com/dlapko/go-mssqldb/blob/main/go.mod).
Such forks can only be used as replacements and can't be imported directly. 
For example, this command doesn't work:

```shell
usqlgen build --import "github.com/dlapko/go-mssqldb"
# the error output includes:
module declares its path as: github.com/microsoft/go-mssqldb
        but was required as: github.com/dlapko/go-mssqldb      
```

Forks that changed the module name to match their repository location can be imported with `--import`,
e.g. [github.com/yugabyte/pgx](https://github.com/yugabyte/pgx) .

### Using a specific version of a driver

If you are not happy with some driver or library version bundled with `usql`, you can change it in two ways.

The preferred approach is adding `--get` parameter to execute `go get` while building.
`go get` [may adjust](https://go.dev/ref/mod#go-get) other dependencies to ensure compatibility with the updated version. 

```shell
usqlgen build --get "github.com/go-sql-driver/mysql@v1.7.1"
```

If the adjustments made by `go get` are not wanted, you may add a replace directive instead:

```shell
usqlgen build --replace "github.com/go-sql-driver/mysql=github.com/go-sql-driver/mysql@v1.7.1"
```

### "Off-label" usage

`usqlgen` may be useful even without changing drivers. For example, `usqlgen` provides the easiest way to
get a `usql` binary without installing it with a package manager or cloning its Git repository:

```bash
go run github.com/sclgo/usqlgen@latest build 
```

## Support

If you encounter problems, please review [open issues](https://github.com/sclgo/usqlgen/issues) and create one if nessesary.
Note that [usql](https://github.com/xo/usql) authors are aware of this project but support
only [their regular releases](https://github.com/xo/usql?tab=readme-ov-file#installing).
