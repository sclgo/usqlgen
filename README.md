# usqlgen

`usqlgen` creates custom distributions of [github.com/xo/usql](https://github.com/xo/usql) - 
a universal single-binary SQL CLI, build by the [github.com/xo](https://github.com/xo) team, and 
inspired by [psql](https://www.postgresql.org/docs/current/app-psql.html).

`usql` is great because it is a multi-platform, multi-database SQL client in a single binary. 
Learn more about `usql` from its [README](https://github.com/xo/usql#readme).

`usqlgen` builds on usql's extensibility and allows you, among other things,
to include arbitrary Go SQL drivers, even for databases not known to `usql`.

`usqlgen` is useful when:

- you want to use a database driver which is not publicly available or is under active development
- you want to use alternative driver for a database supported by `usql`. 
- you want to use a different version of a bundled driver
- you want to use a fork of some of the bundled drivers
- you are working with an obscure or niche database which is not supported by `usql` yet
  - consider contributing the support to `usql` at some point 

`usqlgen` is itself inspired by the 
[OpenTelemetry Collector builder](https://opentelemetry.io/docs/collector/custom-collector/).

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
docker run murfffi/usqlgen build -o - > usql
```

See Docker examples section near the end to see if you need to use Docker - typically not the case.

## Quickstart

To install `usql` with support for an additional driver, review your driver documentation
to find the Go driver name, DSN format and package that needs to be imported to install the
driver. Let's take for example, [MonetDB](https://github.com/MonetDB/MonetDB-Go#readme),
a database, too niche to be supported by the regular `usql` distribution:

```shell
usqlgen build --import "github.com/MonetDB/MonetDB-Go/v2"
```

This creates `usql` executable in the current directory with its default built-in drivers 
together with the driver for MonetDB.
The additional driver is registered using a side-effect import (aka anonymous import)
of the package in the `--import` parameter. The respective module is automatically
determined by `go mod tidy` but can also be specified explicitly.

To connect to the database, refer to [`usql` documentation](https://github.com/xo/usql#readme).
Unlike built-in databases, the `usql` DB URL (connection string) for the new database 
is in the form `driver:dsn`. For example, to connect to MonetDB with the binary we
just built, run:

```shell

# The command below is configured to connect to a local MonetDB started like this:
# docker run -p 50000:50000 -e MDB_DB_ADMIN_PASS=password monetdb/monetdb:latest

./usql monetdb:monetdb:password@localhost:50000/monetdb -c "select 'Hello World'"
```

You can try the same with databases like [rqlite](https://github.com/rqlite/gorqlite), 
[influxdb](https://pkg.go.dev/github.com/influxdata/influxdb-iox-client-go/v2/ioxsql), etc.

For more options, see `usqlgen --help` or review the examples below.

## Limitations

In most cases, you will be only able to use SQL with `usql` distributions, created with `usqlgen`.
Informational commands and autocomplete won't work. 
You will be able to use all commands only when you can pair `--import` with `--copy`,
or if you use `--replace` without `--import`. See the examples below for details.

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

Go environment variables like GOPRIVATE or CGO_ENABLED affect the compilation
as usual.

### Using a driver fork

`usqlgen` can build `usql` with a `replace` directive so that you can use a
SQL driver fork while keeping the `usql` configuration for the target database.
Information commands, schema exploration, and autocomplete will continue to work
if the fork remains compatible enough with the original.

For example, the author of `usql` created a SQL driver 
[github.com/kenshaw/go-impala](https://github.com/kenshaw/go-impala),
fork of the abandoned Apache Impala driver currently used in `usql` - 
[github.com/bippio/go-impala](https://github.com/bippio/go-impala).

The fork was needed, because the abandoned driver used in `usql` 
doesn't work in the current Go release:

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
  
# Connect to local Impala like with the original driver
# We use a usql DB URL as opposed to a driver:dsn URL because we use a built-in driver config
./usql impala://localhost:21050 -c "select 'Hello World'" -t -q
# prints Hello World
```

We included `-- -tags impala` in the command-line so the original driver code in `usql`
is included in the build. The original driver code imports the bippio driver we replaced.

Note that this works only with forks that keep the original module name - 
in this case `github.com/bippio/go-impala` - in their 
[go.mod](https://github.com/kenshaw/go-impala/blob/master/go.mod).
Such forks can only be used as replacements and can't be imported directly. 
For example, the following doesn't work:

```shell
usqlgen build --import "github.com/kenshaw/go-impala"
# the error output includes the following:
#	module declares its path as: github.com/bippio/go-impala
#	        but was required as: github.com/kenshaw/go-impala	       
```

A fork of a driver that changed its module name can only be used as
an alternative driver as described in a following example.

### Using an alternative driver for a supported database

`usqlgen` can replace a driver with an alternative with a different Go module name.
This includes drivers that started as a fork, but changed their Go module name to be independently usable.

For example, [github.com/kprotoss/go-impala](https://github.com/kprotoss/go-impala)
is another fork of the abandoned Apache Impala driver currently used in `usql` (github.com/bippio/go-impala).
Unlike [the fork from the previous example](https://github.com/kenshaw/go-impala),
you can import this one directly using a package under its own module name:

```shell
usqlgen build --import "github.com/kprotoss/go-impala"
# Query Apache Impala running on localhost:
# Since we use --import, the DB URL is in the form 'driver:dsn'
# For this driver, this means we repeat 'impala:' twice, once as driver name,
# and once as a part of the DSN.
./usql impala:impala://localhost:21050 -c "select 'Hello World" -t -q
# prints Hello World
```

Since this driver started as a fork of `bippio/go-impala`, the driver name is the same.

Some alternative drivers diverge more from the built-in ones. For example, 
[github.com/mailru/go-clickhouse/v2](https://github.com/mailru/go-clickhouse) is an
alternative driver for Clickhouse, that shares practically nothing with the driver
included in `usql`. The mailru driver uses the Clickhouse HTTP API instead of the TCP API,
which may be preferable if HTTP middleware is used like a load-balancer or HTTP service mesh. 
The HTTP driver name is `chhttp`. You can directly import the driver as usual:

```shell
usqlgen build --import "github.com/mailru/go-clickhouse/v2"
# Query Clickhouse running on localhost:
./usql usqlgen build --import "github.com/mailru/go-clickhouse/v2" -t -q
# prints Hello World
```

Just importing the alternative driver makes it less functional that the built-in one, because
all Clickhouse-specific configuration in `usql` is lost. In such case, you can try copying the
configuration from the built-in driver to the new one:

```shell
usqlgen build --copy clickhouse:chhttp --import "github.com/mailru/go-clickhouse/v2" -- -tags clickhouse
# TODO Add example that demonstrates the config
```

In the `--copy clickhouse:chhttp` example above, `clickhouse` is the driver name which is the source of the
configuration and `chhttp` is the target. You can review the `usql` built-in configuration for 
any driver in <https://github.com/xo/usql/tree/master/drivers>.
Note that we included `-- -tags clickhouse` to the command-line to ensure the built-in driver 
and its configuration is included in the build. We won't be able to copy it otherwise.

### Using a specific version of a driver

...

## Docker examples