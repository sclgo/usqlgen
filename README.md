# usqlgen

`usqlgen` creates custom distributions of [github.com/xo/usql](https://github.com/xo/usql) - 
a universal single-binary SQL CLI, build by the [github.com/xo](https://github.com/xo) team, and 
inspired by [psql](https://www.postgresql.org/docs/current/app-psql.html).

`usql` is great because it is a multi-platform, multi-database SQL client in a single binary,
with minimal dependencies. It support a bunch of databases by default and its build
automation allows it to be customized to support even more. Learn more about `usql` from its
[README](https://github.com/xo/usql#readme).

`usqlgen` builds on usql's tag-based distribution customization and allows you to include arbitrary Go SQL 
drivers, even for databases not known to `usql`

`usqlgen` is useful when:

- you want to use a database driver which is not publicly available or is under active development
- you want to use alternative driver for a database supported by `usql`. 
  <https://go.dev/wiki/SQLDrivers> references several other with multiple drivers alternatives.
- you want to use a specific version of a driver - older or newer than the one included in `usql`
- you want to use a fork of some of the bundled drivers
- you are working with an obscure or niche database which is not supported by `usql` yet
  - consider contributing to `usql` 

`usqlgen` is itself inspired by the 
[OpenTelemetry Collector builder](https://opentelemetry.io/docs/collector/custom-collector/).

## Installing

Install `usqlgen` with Go 1.21+:

```shell
go install github.com/sclgo/usqlgen@latest
```

You can run usqlgen without installing with:

```shell
go run github.com/sclgo/usqlgen@latest
```

Running without installing is useful in minimal container builds. 

If you don't have Go 1.21+, you can run it with Docker:

```shell
docker run murfffi/usqlgen
```

## Quickstart

To install `usql` with support for an additional driver, review your driver documentation
to find the Go driver name, DSN format and package that needs to be imported to install the
driver. Let's take for example, [MonetDB](https://github.com/MonetDB/MonetDB-Go#readme),
a database, too niche to be supported by the regular `usql` distribution:

```shell
usqlgen build --import "github.com/MonetDB/MonetDB-Go/v2"
```

This creates `usql` executable in the current directory with its default drivers and the driver for MonetDB.
The additional driver is registered using a side-effect import (aka anonymous import)
of the package in the `--import` parameter.

To connect to the database, refer to [`usql` documentation](https://github.com/xo/usql#readme).
The `usql` connection string for the new database is in the form `driver:dsn` as described
in the driver documentation. For example, to connect to MonetDB with the binary we
just built, run:

```shell

# The command below is configured to connect to a local MonetDB started like this:
# docker run -p 50000:50000 -e MDB_DB_ADMIN_PASS=password monetdb/monetdb:latest

./usql monetdb:monetdb:password@localhost:50000/monetdb -c "select 'Hello World'"
```

You can try the same with databases like [rqlite](https://github.com/rqlite/gorqlite).

For more options, see `usqlgen --help` or review the examples below.

## Limitations

In most cases, you will be only able to use SQL with `usql`
Informational commands and autocomplete won't work. The examples below
describe the exceptions where information commands and autocomplete will still work.

## Examples

### Installing the customized `usql`

Use `usqlgen install ...` to install the customized `usql` to `GOPATH/bin` which is
typically on the search path.

```shell
usqlgen build --import "github.com/MonetDB/MonetDB-Go/v2"
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

Go environment variables like GOPRIVATE or CGO_ENABLED affect the compilation
as usual.

### Using an alternative driver for a supported database

`usqlgen` can build `usql` with a replace directive so that you can use a specific
SQL driver while keeping the `usql` configuration for the target database.
Information commands, schema exploration, and autocomplete will continue to work
if the replacement driver is compatible enough with the original.

For example, you may want to use <https://github.com/glebarez/go-sqlite>
instead of the other SQLite implementations if you are exploring a database file created with that
specific one. <https://go.dev/wiki/SQLDrivers> references 
several other databases with multiple drivers alternatives.

```shell
usqlgen build --replace "modernc.org/sqlite=github.com/glebarez/go-sqlite@latest"
```

With the same approach, you can use a fork of the built-in drivers or a specific version of a driver. 

