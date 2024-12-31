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

In most cases, using or contributing `usql` directly is best. `usqlgen` is useful when you want to avoid editing `usql` code and:

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

Running without installing is useful in minimal container builds. 

If you don't have Go 1.21+, you can run `usqlgen` with Docker:

```shell
docker run --rm golang:1.23 \
  go run github.com/sclgo/usqlgen@latest build ...add parameters here... -o - > ./usql

chmod +x ./usql
```

Ensure that the Docker image matches the OS you are using - Windows container for Windows, Alpine for Alpine Linux, etc.
The default golang image works for Debian-based Linux distributions like Ubuntu.

## Quickstart

To install `usql` with support for an additional driver, first review your driver documentation
to find the Go driver name, DSN format and package that needs to be imported to install the
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

To connect to the database, refer to [`usql` documentation](https://github.com/xo/usql#readme).
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
[rqlite](https://github.com/rqlite/gorqlite), 
[influxdb](https://pkg.go.dev/github.com/influxdata/influxdb-iox-client-go/v2/ioxsql),
[Dremio or Apache Drill](https://github.com/factset/go-drill), etc.

`usqlgen` also allows you to use alternative drivers of supported databases. Examples include:

- [github.com/sclgo/impala-go](https://github.com/sclgo/impala-go) - modernized variant of the built-in Impala driver
- [github.com/mailru/go-clickhouse/v2](https://github.com/mailru/go-clickhouse) - HTTP-only alternative of the built-in Clickhouse driver

For more options, see `usqlgen --help` or review the examples below.

## Limitations

Most `usql` [backslash (meta) commands](https://github.com/xo/usql?tab=readme-ov-file#backslash-commands) work 
with new drivers added with `--import`, including 
[cross-database `\copy`](https://github.com/xo/usql?tab=readme-ov-file#copying-between-databases). 
Informational commands and autocomplete won't work though.

## Examples

### Installing the customized `usql`

Use `usqlgen install ...` to install the customized `usql` to `GOPATH/bin` which is
typically on the search path.

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

For example, one of the authors of `usql` created a SQL driver 
[github.com/kenshaw/go-impala](https://github.com/kenshaw/go-impala),
fork of the abandoned Apache Impala driver currently used in `usql` - 
[github.com/bippio/go-impala](https://github.com/bippio/go-impala).

The fork was needed, because the abandoned driver used in `usql` 
doesn't work in recent Go releases:

```shell
go install -tags impala github.com/xo/usql@latest
# a bunch of error messages
```

To use the fork, run:

```shell
usqlgen build --replace "github.com/bippio/go-impala=github.com/kenshaw/go-impala@master" -- -tags impala
```

To test the compiled `usql` binary:

```shell
# Start local Impala
docker run -d --rm -p 21050:21050 --memory=4096m \
  apache/kudu:impala-latest impala
  
# Connect to local Impala
# We use a usql DB URL as opposed to a driver:dsn URL because we use a built-in driver config
./usql impala://localhost:21050 -c "select 'Hello World'" -t -q
# prints Hello World
```

We included `-- -tags impala` in the command-line so the original driver code in `usql`
is included in the build.

Note that this works only with forks that keep the original module name - 
in this case `github.com/bippio/go-impala` - in their 
[go.mod](https://github.com/kenshaw/go-impala/blob/master/go.mod).
Such forks can only be used as replacements and can't be imported directly. 
For example, this command doesn't work:

```shell
usqlgen build --import "github.com/kenshaw/go-impala"
# the error output includes:
#	module declares its path as: github.com/bippio/go-impala
#	        but was required as: github.com/kenshaw/go-impala	       
```

Forks that changed the module name to match their repository location can be imported with `--import`,
e.g. [github.com/sclgo/impala-go](https://github.com/sclgo/impala-go) .

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

## Support

If you encounter problems, please review [open issues](https://github.com/sclgo/usqlgen/issues) and create one if nessesary.
Note that [usql](https://github.com/xo/usql) authors are aware of this project but support
only [their regular releases](https://github.com/xo/usql?tab=readme-ov-file#installing).
