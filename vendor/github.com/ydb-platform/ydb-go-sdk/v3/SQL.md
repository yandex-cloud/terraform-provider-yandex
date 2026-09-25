# `database/sql` driver for `YDB`

In addition to native YDB Go driver APIs, package `ydb-go-sdk` provides standard APIs for `database/sql`.
It allows to use "regular" Go facilities to access YDB databases.
Behind the scene, `database/sql` APIs are implemented using the native interfaces.


## Table of contents
1. [Initialization of `database/sql` driver](#init)
   * [Configure driver with `ydb.Connector` (recommended way)](#init-connector)
   * [Configure driver with data source name or connection string](#init-dsn)
2. [Client balancing](#balancing)
3. [Session pooling](#session-pool)
4. [Query execution](#queries)
   * [Queries on database object](#queries-db)
   * [Queries on transaction object](#queries-tx)
5. [Query modes (DDL, DML, DQL, etc.)](#query-modes)
6. [Retry helpers for `YDB` `database/sql` driver](#retry)
   * [Over `sql.Conn` object](#retry-conn)
   * [Over `sql.Tx`](#retry-tx)
7. [Query args types](#arg-types)
8. [Query bindings](#bindings)
9. [Accessing the native driver from `*sql.DB`](#unwrap)
   * [Driver with go's 1.18 supports also `*sql.Conn` for unwrapping](#unwrap-cc)
10. [Troubleshooting](#troubleshooting)
   * [Logging driver events](#logging)
   * [Add metrics about SDK's events](#metrics)
   * [Add `Jaeger` traces about driver events](#jaeger)
11. [Example of usage](#example)

## Initialization of `database/sql` driver <a name="init"></a>

### Configure driver with `ydb.Connector` (recommended way) <a name="init-connector"></a>
```go
import (
    "database/sql"
  
    "github.com/ydb-platform/ydb-go-sdk/v3"
)

func main() {
    // init native ydb-go-sdk driver
    nativeDriver, err := ydb.Open(context.TODO(), "grpc://localhost:2136/local",
        // See many ydb.Option's for configure driver https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#Option
    )
    if err != nil {
        // fallback on error
    }
    defer nativeDriver.Close(context.TODO())
    connector, err := ydb.Connector(nativeDriver,
        // See ydb.ConnectorOption's for configure connector https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#ConnectorOption
    )
    if err != nil {
        // fallback on error
    }
    db := sql.OpenDB(connector)
    defer db.Close()
    db.SetMaxOpenConns(100)
    db.SetMaxIdleConns(100)
    db.SetConnMaxIdleTime(time.Second) // workaround for background keep-aliving of YDB sessions
}
```

### Configure driver with data source name or connection string <a name="init-dsn"></a>
```go
import (
    "database/sql"
  
    _ "github.com/ydb-platform/ydb-go-sdk/v3"
)

func main() {
    db, err := sql.Open("ydb", "grpc://localhost:2136/local")
    defer db.Close()
}
```
Data source name parameters:
* `token` – access token to be used during requests (required)
* static credentials with authority part of URI, like `grpcs://root:password@localhost:2135/local`
* `query_mode=scripting` - you can redefine default [DML](https://en.wikipedia.org/wiki/Data_manipulation_language) query mode

## Client balancing <a name="balancing"></a>

`database/sql` driver for `YDB` like as native driver for `YDB` use client balancing, which happens on `CreateSession` request.
At this time, native driver choose known node for execute request by according balancer algorithm.
Default balancer algorithm is a `random choice`.
Client balancer may be re-configured with option `ydb.WithBalancer`:
```go
import (
    "github.com/ydb-platform/ydb-go-sdk/v3/balancers"
)
func main() {
    nativeDriver, err := ydb.Open(context.TODO(), "grpc://localhost:2136/local",
        ydb.WithBalancer(
            balancers.PreferLocationsWithFallback(
                balancers.RandomChoice(), "a", "b",
            ),
        ),
    )
    if err != nil {
        // fallback on error
    }
    connector, err := ydb.Connector(nativeDriver)
    if err != nil {
        // fallback on error
    }
    db := sql.OpenDB(connector)
    db.SetMaxOpenConns(100)
    db.SetMaxIdleConns(100)
    db.SetConnMaxIdleTime(time.Second) // workaround for background keep-aliving of YDB sessions
}
```

## Session pooling <a name="session-pool"></a>

Native driver `ydb-go-sdk/v3` implements the internal session pool, which uses with `db.Table().Do()` or `db.Table().DoTx()` methods.
Internal session pool are configured with options like `ydb.WithSessionPoolSizeLimit()` and other.
Unlike the session pool in the native driver, `database/sql` contains a different implementation of the session pool, which is configured with `*sql.DB.SetMaxOpenConns` and `*sql.DB.SetMaxIdleConns`.
Lifetime of a `YDB` session depends on driver configuration and error occurance, such as `sql.driver.ErrBadConn`.
`YDB` driver for `database/sql` includes the logic to transform the internal `YDB` error codes into `sql.driver.ErrBadConn` and other retryable and non-retryable error types.

In most cases the implementation of `database/sql` driver for YDB allows to complete queries without user-visible errors.
But some errors need to be handled on the client side, by re-running not a single operation, but a complete transaction.
Therefore the transaction logic needs to be wrapped with retry helpers, such as `retry.Do` or `retry.DoTx` (see more about retry helpers in the [retry section](#retry)).

## Query execution <a name="queries"></a>

### Queries on database object <a name="queries-db"></a>
```go
rows, err := db.QueryContext(ctx,
    "SELECT series_id, title, release_date FROM `/full/path/of/table/series`;"
)
if err != nil {
    log.Fatal(err)
}
defer rows.Close() // always close rows 
var (
    id          *string
    title       *string
    releaseDate *time.Time
)
for rows.Next() { // iterate over rows
    // apply database values to go's type destinations
    if err = rows.Scan(&id, &title, &releaseDate); err != nil {
        log.Fatal(err)
    }
    log.Printf("> [%s] %s (%s)", *id, *title, releaseDate.Format("2006-01-02"))
}
if err = rows.Err(); err != nil { // always check final rows err
    log.Fatal(err)
}
```

### Queries on transaction object <a name="queries-tx"></a>

`database/sql` driver over `ydb-go-sdk/v3` supports next isolation leveles:
- read-write (mapped to `SerializableReadWrite` transaction control)
  ```go
  rw := sql.TxOption{
    ReadOnly: false,
    Isolation: sql.LevelDefault,
  }
  ```
- read-only (mapped to `OnlineReadOnly` transaction settings on each request, will be mapped to true `SnapshotReadOnly` soon)
  ```go
  ro := sql.TxOption{
    ReadOnly: true,
    Isolation: sql.LevelSnapshot,
  }
  ```

Example of works with transactions:
```go
tx, err := db.BeginTx(ctx, sql.TxOption{
  ReadOnly: true,
  Isolation: sql.LevelSnapshot,
})
if err != nil {
    log.Fatal(err)
}
defer tx.Rollback()
rows, err := tx.QueryContext(ctx,
    "SELECT series_id, title, release_date FROM `/full/path/of/table/series`;"
)
if err != nil {
    log.Fatal(err)
}
defer rows.Close() // always close rows 
var (
    id          *string
    title       *string
    releaseDate *time.Time
)
for rows.Next() { // iterate over rows
    // apply database values to go's type destinations
    if err = rows.Scan(&id, &title, &releaseDate); err != nil {
        log.Fatal(err)
    }
    log.Printf("> [%s] %s (%s)", *id, *title, releaseDate.Format("2006-01-02"))
}
if err = rows.Err(); err != nil { // always check final rows err
    log.Fatal(err)
}
if err = tx.Commit(); err != nil {
    log.Fatal(err)
}
```

## Query modes (DDL, DML, DQL, etc.) <a name="query-modes"></a>
Currently the `YDB` server APIs require the use of a proper GRPC service method depending on the specific request type.
In particular, [DDL](https://en.wikipedia.org/wiki/Data_definition_language) must be called through `table.session.ExecuteSchemeQuery`,
[DML](https://en.wikipedia.org/wiki/Data_manipulation_language) needs `table.session.Execute`, and
[DQL](https://en.wikipedia.org/wiki/Data_query_language) should be passed via `table.session.Execute` or `table.session.StreamExecuteScanQuery`.
`YDB` also has a so-called "scripting" service, which supports different query types within a single method,
but without support for transactions.

Unfortunately, this leads to the need to choose the proper query mode on the application side.

`YDB` team has a roadmap goal to implement a single universal service method for executing
different query types and without the limitations of the "scripting" service method.

`database/sql` driver implementation for `YDB` supports the following query modes:
* `ydb.DataQueryMode` - default query mode, for lookup [DQL](https://en.wikipedia.org/wiki/Data_query_language) queries and [DML](https://en.wikipedia.org/wiki/Data_manipulation_language) queries.
* `ydb.ExplainQueryMode` - for gettting plan and [AST](https://en.wikipedia.org/wiki/Abstract_syntax_tree) of the query
* `ydb.ScanQueryMode` - for heavy [OLAP](https://en.wikipedia.org/wiki/Online_analytical_processing) style scenarious, with [DQL-only](https://en.wikipedia.org/wiki/Data_query_language) queries. Read more about scan queries in [ydb.tech](https://ydb.tech/en/docs/concepts/scan_query)
* `ydb.SchemeQueryMode` - for [DDL](https://en.wikipedia.org/wiki/Data_definition_language) queries
* `ydb.ScriptingQueryMode` - for [DDL](https://en.wikipedia.org/wiki/Data_definition_language), [DML](https://en.wikipedia.org/wiki/Data_manipulation_language), [DQL](https://en.wikipedia.org/wiki/Data_query_language) queries (not a [TCL](https://en.wikipedia.org/wiki/SQL#Transaction_controls)). Be careful: queries execute longer than with other query modes, and consume more server-side resources

Example for changing the default query mode:
```go
res, err = db.ExecContext(ydb.WithQueryMode(ctx, ydb.SchemeQueryMode),
   "DROP TABLE `/full/path/to/table/series`",
)
```

## Changing the transaction control mode <a name="tx-control"></a>

Default `YDB`'s transaction control mode is a `SerializableReadWrite`. 
Default transaction control mode can be changed outside of interactive transactions by updating the context object:
```go
rows, err := db.QueryContext(ydb.WithTxControl(ctx, table.OnlineReadOnlyTxControl()),
    "SELECT series_id, title, release_date FROM `/full/path/of/table/series`;"
)
```

## Retry helpers for `YDB` `database/sql` driver <a name="retry"></a>

`YDB` is a distributed `RDBMS` with non-stop 24/7 releases flow.
It means some nodes may be unavailable for queries at some point in time.
Network errors may also occur.
That's why some queries may complete with errors.
Most of those errors are transient.
`ydb-go-sdk`'s "knows" what to do on specific error: retry or not, with or without backoff, with or without the need to re-establish the session, etc.
`ydb-go-sdk` provides retry helpers which can work either with the database connection object, or with the transaction object.

### Retries over `sql.Conn` object <a name="retry-conn"></a>

`retry.Do` helper accepts custom lambda, which must return error if it happens during the processing,
or nil if the operation succeeds.
```go
import (
   "github.com/ydb-platform/ydb-go-sdk/v3/retry"
)
...
err := retry.Do(context.TODO(), db, func(ctx context.Context, cc *sql.Conn) error {
   // work with cc
   rows, err := cc.QueryContext(ctx, "SELECT 1;")
   if err != nil {
       return err // return err to retryer
   }
   ...
   return nil // good final of retry operation
}, retry.WithIdempotent(true))

```

### Retries over `sql.Tx` <a name="retry-tx"></a>

`retry.DoTx` helper accepts custom lambda, which must return error if it happens during processing,
or nil if the operation succeeds.

`tx` object is a prepared transaction object.

The logic within the custom lambda does not need the explicit commit or rollback at the end - `retry.DoTx` does it automatically.

```go
import (
    "github.com/ydb-platform/ydb-go-sdk/v3/retry"
)
...
err := retry.DoTx(context.TODO(), db, func(ctx context.Context, tx *sql.Tx) error {
    // work with tx
    rows, err := tx.QueryContext(ctx, "SELECT 1;")
    if err != nil {
        return err // return err to retryer
    }
    ...
    return nil // good final of retry tx operation
}, retry.WithIdempotent(true), retry.WithTxOptions(&sql.TxOptions{
    Isolation: sql.LevelSnapshot,
    ReadOnly:  true,
}))
```

## Specifying query parameters <a name="arg-types"></a>

`database/sql` driver for `YDB` supports the following types of query parameters:
* `sql.NamedArg` types with native Go's types and ydb types:
   ```go
   rows, err := db.QueryContext(ctx, 
      `
        DECLARE $title AS Text;
        DECLARE $views AS Uint64;
        DECLARE $ts AS Datetime;
        SELECT season_id FROM seasons WHERE title LIKE $title AND views > $views AND first_aired > $ts;
      `,
      sql.Named("$title", "%Season 1%"), // argument name with prefix `$` 
      sql.Named("views", uint64(1000)),  // argument name without prefix `$` (driver will prepend `$` if necessary)
      sql.Named("$ts", types.DatetimeValueFromTime( // native ydb type
        time.Now().Add(-time.Hour*24*365), 
      )),
   )
   ```
* `table.ParameterOption` arguments:
   ```go
   rows, err := db.QueryContext(ctx, 
      `
        DECLARE $title AS Text;
        DECLARE $views AS Uint64;
        SELECT season_id FROM seasons WHERE title LIKE $title AND views > $views;
      `,
      table.ValueParam("$seasonTitle", types.TextValue("%Season 1%")),
      table.ValueParam("$views", types.Uint64Value((1000)),
   )
   ```
* single `*table.QueryParameters` argument:
   ```go
   rows, err := db.QueryContext(ctx, 
      `
        DECLARE $title AS Text;
        DECLARE $views AS Uint64;
        SELECT season_id FROM seasons WHERE title LIKE $title AND views > $views;
      `,
      table.NewQueryParameters(
          table.ValueParam("$seasonTitle", types.TextValue("%Season 1%")),
          table.ValueParam("$views", types.Uint64Value((1000)),
      ),
   )
   ```

## Query bindings <a name="bindings"></a>

YQL is a SQL dialect with YDB specific strict types. This is great for performance and correctness, but sometimes can be a bit dauting to express in a query, especially when then need to be parametrized externally from the application side. For instance, when a YDB query needs to be parametrized, each parameter has name and type provided via `DECLARE` statement.

Also, because YDB tables reside in virtual filesystem-like structure their names can be quite lengthy. There's a `PRAGMA TablePathPrefix` that can scope the rest of the query inside a given prefix, simplifying table names. For example, a query to the table `/local/path/to/tables/seasons` might look like this:

```
DECLARE $title AS Text;
DECLARE $views AS Uint64;
SELECT season_id
FROM `/local/path/to/tables/seasons`
WHERE title LIKE $title AND views > $views;
```

Using the `PRAGMA` statement, you can simplify the prefix part in the name of all tables involved in the `YQL`-query:

```
PRAGMA TablePathPrefix("/local/path/to/tables/");
DECLARE $title AS Text;
DECLARE $views AS Uint64;
SELECT season_id FROM seasons WHERE title LIKE $title AND views > $views;
```

`database/sql` driver for `YDB` (part of [YDB Go SDK](https://github.com/ydb-platform/ydb-go-sdk)) supports query enrichment for:

* specifying **_TablePathPrefix_**
* **_declaring_** types of parameters
* **_numeric_** or **_positional_** parameters

These query enrichments can be enabled explicitly on the initializing step using connector options or connection string parameter `go_auto_bind`. By default `database/sql` driver for `YDB` doesn’t modify queries.

The following example without bindings demonstrates explicit working with `YDB` types:

```go
import (
  "context"
  "database/sql"

  "github.com/ydb-platform/ydb-go-sdk/v3"
  "github.com/ydb-platform/ydb-go-sdk/v3/table"
  "github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

func main() {
  db := sql.Open("ydb", "grpc://localhost:2136/local")
  defer db.Close()

  // positional args
  row := db.QueryRowContext(context.TODO(), `
    PRAGMA TablePathPrefix("/local/path/to/my/folder");
    DECLARE $p0 AS Int32;
    DECLARE $p1 AS Utf8;
    SELECT $p0, $p1`, 
    sql.Named("$p0", 42), 
    table.ValueParam("$p1", types.TextValue("my string")),
  )
  // process row ...
}
```

As you can see, this example also required importing of `ydb-go-sdk` packages and working with them directly.

With enabled bindings enrichment, the same result can be achieved much easier:

- with _**connection string_** params:
```go
import (
  "context"
  "database/sql"

  _ "github.com/ydb-platform/ydb-go-sdk/v3" // anonymous import for registering driver
)

func main() {
  var (
    ctx = context.TODO()
    db = sql.Open("ydb", 
      "grpc://localhost:2136/local?"+
        "go_auto_bind="+
          "table_path_prefix(/local/path/to/my/folder),"+
          "declare,"+
          "positional", 
    )
  )
  defer db.Close() // cleanup resources

  // positional args
  row := db.QueryRowContext(ctx, `SELECT ?, ?`, 42, "my string")
```

- with _**connector options**_:
```go
import (
  "context"
  "database/sql"

  "github.com/ydb-platform/ydb-go-sdk/v3" 
)

func main() {
  var (
    ctx = context.TODO()
    nativeDriver = ydb.MustOpen(ctx, "grpc://localhost:2136/local")
    db = sql.OpenDB(ydb.MustConnector(nativeDriver,
      query.TablePathPrefix("/local/path/to/my/folder"), // bind pragma TablePathPrefix
      query.Declare(),                                   // bind parameters declare
      query.Positional(),                                // bind positional args
    ))
  )
  defer nativeDriver.Close(ctx) // cleanup resources
  defer db.Close()

  // positional args
  row := db.QueryRowContext(ctx, `SELECT ?, ?`, 42, "my string")
```

The original simple query `SELECT ?, ?` will be expanded on the driver side to the following:

```sql
-- bind TablePathPrefix 
PRAGMA TablePathPrefix("/local/path/to/my/folder");

-- bind declares
DECLARE $p0 AS Int32;
DECLARE $p1 AS Utf8;

-- origin query with positional args replacement
SELECT $p0, $p1
```

This expanded query will be sent to `ydbd` server instead of the original one.

For additional examples of query enrichment, see `ydb-go-sdk` documentation:

* specifying `TablePathPrefix`:
    * using [connection string parameter](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindTablePathPrefix)
    * using [connector option](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindTablePathPrefixOverConnector)
* declaring bindings:
    * using [connection string parameter](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindDeclare)
    * using [connector option](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindDeclareOverConnector)
* positional arguments binding:
    * using [connection string parameter](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindPositionalArgs)
    * using [connector option](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindPositionalArgsOverConnector)
* numeric arguments binding:
    * using [connection string parameter](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindNumericlArgs)
    * using [connector option](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3#example-package-DatabaseSQLBindNumericArgsOverConnector)

For a deep understanding of query enrichment see also [unit-tests](https://github.com/ydb-platform/ydb-go-sdk/blob/master/query/bind_test.go).

You can write your own unit tests to check correct binding of your queries like this:

```go
func TestBinding(t *testing.T) {
  bindings := query.NewBind(
    query.TablePathPrefix("/local/path/to/my/folder"), // bind pragma TablePathPrefix
    query.Declare(),                                   // bind parameters declare
    query.Positional(),                                // auto-replace positional args
  )
  query, params, err := bindings.ToYQL("SELECT ?, ?, ?", 1, uint64(2), "3")
  require.NoError(t, err)
  require.Equal(t, `-- bind TablePathPrefix
PRAGMA TablePathPrefix("/local/path/to/my/folder");

-- bind declares
DECLARE $p0 AS Int32;
DECLARE $p1 AS Uint64;
DECLARE $p2 AS Utf8;

-- origin query with positional args replacement
SELECT $p0, $p1, $p2`, query)
  require.Equal(t, table.NewQueryParameters(
    table.ValueParam("$p0", types.Int32Value(1)),
    table.ValueParam("$p1", types.Uint64Value(2)),
    table.ValueParam("$p2", types.TextValue("3")),
  ), params)
}
```

## Accessing the native driver from `*sql.DB` <a name="unwrap"></a>

```go
db, err := sql.Open("ydb", "grpc://localhost:2136/local")
if err != nil {
  t.Fatal(err)
}
nativeDriver, err = ydb.Unwrap(db)
if err != nil {
  t.Fatal(err)
}
nativeDriver.Table().Do(ctx, func(ctx context.Context, s table.Session) error {
  // doing with native YDB session
  return nil
})
```

### Driver with go's 1.18 supports also `*sql.Conn` for unwrapping <a name="unwrap-cc"></a>

For example, this feature may be helps with `retry.Do`:
```go
err := retry.Do(context.TODO(), db, func(ctx context.Context, cc *sql.Conn) error {
    nativeDriver, err := ydb.Unwrap(cc)
    if err != nil {
        return err // return err to retryer
    }
    res, err := nativeDriver.Scripting().Execute(ctx,
        "SELECT 1+1",
        table.NewQueryParameters(),
    )
    if err != nil {
        return err // return err to retryer
    }
    // work with cc
    rows, err := cc.QueryContext(ctx, "SELECT 1;")
    if err != nil {
        return err // return err to retryer
    }
    ...
    return nil // good final of retry operation
}, retry.WithIdempotent(true))
```

## Troubleshooting <a name="troubleshooting"></a>

### Logging driver events <a name="logging"></a>

Adding a logging driver events allowed only if connection to `YDB` opens over [connector](##init-connector).
Adding of logging provides with [debug adapters](README.md#debug) and wrotes in [migration notes](MIGRATION_v2_v3.md#logs).

Example of adding `zap` logging:
```go
import (
    "github.com/ydb-platform/ydb-go-sdk/v3/trace"
    ydbZap "github.com/ydb-platform/ydb-go-sdk-zap"
)
...
nativeDriver, err := ydb.Open(ctx, connectionString,
    ...
    ydbZap.WithTraces(log, trace.DetailsAll),
)
if err != nil {
    // fallback on error
}
connector, err := ydb.Connector(nativeDriver)
if err != nil {
    // fallback on error
}
db := sql.OpenDB(connector)
```

### Add metrics about SDK's events <a name="metrics"></a>

Adding of driver events monitoring allowed only if connection to `YDB` opens over [connector](##init-connector).
Monitoring of driver events provides with [debug adapters](README.md#debug) and wrotes in [migration notes](MIGRATION_v2_v3.md#metrics).

Example of adding `Prometheus` monitoring:
```go
import (
    "github.com/ydb-platform/ydb-go-sdk/v3/trace"
    ydbMetrics "github.com/ydb-platform/ydb-go-sdk-prometheus"
)
...
nativeDriver, err := ydb.Open(ctx, connectionString,
    ...
    ydbMetrics.WithTraces(registry, ydbMetrics.WithDetails(trace.DetailsAll)),
)
if err != nil {
    // fallback on error
}
connector, err := ydb.Connector(nativeDriver)
if err != nil {
    // fallback on error
}
db := sql.OpenDB(connector)
```

### Add `Jaeger` traces about driver events <a name="jaeger"></a>

Adding of `Jaeger` traces about driver events allowed only if connection to `YDB` opens over [connector](##init-connector).
`Jaeger` tracing provides with [debug adapters](README.md#debug) and wrotes in [migration notes](MIGRATION_v2_v3.md#jaeger).

Example of adding `Jaeger` tracing:
```go
import (
    "github.com/ydb-platform/ydb-go-sdk/v3/trace"
    ydbOpentracing "github.com/ydb-platform/ydb-go-sdk-opentracing"
)
...
nativeDriver, err := ydb.Open(ctx, connectionString,
    ...
    ydbOpentracing.WithTraces(trace.DriverConnEvents | trace.DriverClusterEvents | trace.DriverRepeaterEvents | trace.DiscoveryEvents),
)
if err != nil {
    // fallback on error
}
connector, err := ydb.Connector(nativeDriver)
if err != nil {
    // fallback on error
}
db := sql.OpenDB(connector)
```

## Example of usage <a name="example"></a>

[Basic example](https://github.com/ydb-platform/ydb-go-sdk/tree/master/examples/basic/native/query) about series written with `database/sql` driver for `YDB` placed in [examples repository](https://github.com/ydb-platform/ydb-go-sdk/tree/master/examples/database/sql)  
