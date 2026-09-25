# Migration from `ydb-go-sdk/v2` to `ydb-go-sdk/v3` with `database/sql` API driver usages

> Article contains some cases for migrate from `github.com/yandex-cloud/ydb-go-sdk/v2/ydbsql` to `github.com/ydb-platform/ydb-go-sdk/v3`

> Note: the article is being updated.

## `sql.Connector` initialization

Package `database/sql` provides two ways for making driver:
- from connection string (see [sql.Open(driverName, connectionString)](https://pkg.go.dev/database/sql#Open))
- from custom connector (see [sql.OpenDB(connector)](https://pkg.go.dev/database/sql#OpenDB))

Second way (making driver from connector) are different in `v2` and `v3`:
- in `v2`:
  `ydbsql.Connector` returns single result (`sql.driver.Connector`) and init driver lazy on first request. This design causes rare errors. 
  ```go
  db := sql.OpenDB(ydbsql.Connector(
    ydbsql.WithDialer(dialer),
    ydbsql.WithEndpoint(params.endpoint),
    ydbsql.WithClient(&table.Client{
      Driver: driver,
    }),
  ))
  defer db.Close()
  ```
- in `v3`:
  `ydb.Connector` creates `sql.driver.Connector` from native `YDB` driver, returns two results (`sql.driver..Connector` and error) and exclude some lazy driver initialization.
  ```go
  nativeDriver, err := ydb.Open(context.TODO(), "grpc://localhost:2136/local")
  if err != nil {
    // fallback on error
  }
  defer nativeDriver.Close(context.TODO())
  connector, err := ydb.Connector(nativeDriver)
  if err != nil {
    // fallback on error
  }
  db := sql.OpenDB(connector)
  defer db.Close()
  ```

## Read-only isolation levels

In `ydb-go-sdk/v2/ydbsql` was allowed `sql.LevelReadCommitted` and `sql.LevelReadUncommitted` isolation levels for read-only interactive transactions. It implements with fake transaction with true `OnlineReadOnly` transaction control on each query inside transaction.

Transaction controls `OnlineReadOnly` and `StaleReadOnly` will deprecate in the future.

That's why `ydb-go-sdk/v3` allowed only `sql.LevelSnapshot` for read-only interactive transactions. Currently, snapshot isolation implements over `fake` transaction with true `OnlineReadOnly` transaction control on each request inside transaction.
YDB implements snapshot isolation, but this feature is not deployed on YDB clusters now. After full deploying on each YDB cluster fake transaction will be replaced to true read-only interactive transaction with snapshot isolation level.  