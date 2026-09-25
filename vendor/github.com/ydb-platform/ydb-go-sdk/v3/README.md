# `ydb-go-sdk` - pure Go native and `database/sql` driver for [YDB](https://github.com/ydb-platform/ydb)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/ydb-platform/ydb/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/ydb-platform/ydb-go-sdk.svg?style=flat-square)](https://github.com/ydb-platform/ydb-go-sdk/releases)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ydb-platform/ydb-go-sdk/v3)](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3)
![tests](https://github.com/ydb-platform/ydb-go-sdk/workflows/tests/badge.svg)
![lint](https://github.com/ydb-platform/ydb-go-sdk/workflows/lint/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/ydb-platform/ydb-go-sdk/v3)](https://goreportcard.com/report/github.com/ydb-platform/ydb-go-sdk/v3)
[![codecov](https://codecov.io/gh/ydb-platform/ydb-go-sdk/badge.svg?precision=2)](https://app.codecov.io/gh/ydb-platform/ydb-go-sdk)
![Code lines](https://sloc.xyz/github/ydb-platform/ydb-go-sdk/?category=code)
[![View examples](https://img.shields.io/badge/learn-examples-brightgreen.svg)](https://github.com/ydb-platform/ydb-go-sdk/tree/master/examples)
[![Telegram](https://img.shields.io/badge/chat-on%20Telegram-2ba2d9.svg)](https://t.me/ydb_en)
[![WebSite](https://img.shields.io/badge/website-ydb.tech-blue.svg)](https://ydb.tech)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/ydb-platform/ydb-go-sdk/blob/master/CONTRIBUTING.md)

Supports `discovery`, `operation`, `table`, `query`, `coordination`, `ratelimiter`, `scheme`, `scripting` and `topic` clients for [YDB](https://ydb.tech).
`YDB` is an open-source Distributed SQL Database that combines high availability and scalability with strict consistency and [ACID](https://en.wikipedia.org/wiki/ACID) transactions.
`YDB` was created primarily for [OLTP](https://en.wikipedia.org/wiki/Online_transaction_processing) workloads and supports some [OLAP](https://en.wikipedia.org/wiki/Online_analytical_processing) scenarious.

## Supported Go Versions

`ydb-go-sdk` supports all Go versions supported by the official [Go Release Policy](https://go.dev/doc/devel/release#policy).
That is, the latest two versions of Go (or more, but with no guarantees for extra versions).


## Versioning Policy

`ydb-go-sdk` comply to guidelines from [SemVer2.0.0](https://semver.org/) with an several [exceptions](VERSIONING.md).

## Installation

```sh
go get -u github.com/ydb-platform/ydb-go-sdk/v3
```

## Example Usage <a name="example"></a>

* connect to YDB
```go
db, err := ydb.Open(ctx, "grpc://localhost:2136/local")
if err != nil {
    log.Fatal(err)
}
```
* execute `SELECT` query with `Query` service client
 ```go
// Do retry operation on errors with best effort
err := db.Query().Do( // Do retry operation on errors with best effort
	ctx, // context manage exiting from Do
	func(ctx context.Context, s query.Session) (err error) { // retry operation
		streamResult, err := s.Query(ctx, `SELECT 42 as id, "myStr" as myStr;`)
		if err != nil {
			return err // for auto-retry with driver
		}
		defer func() { _ = streamResult.Close(ctx) }() // cleanup resources
		for rs, err := range streamResult.ResultSets(ctx) {
			if err != nil {
				return err
			}
			for row, err := range rs.Rows(ctx) {
				if err != nil {
					return err
				}
				type myStruct struct {
					Id  int32  `sql:"id"`
					Str string `sql:"myStr"`
				}
				var s myStruct
				if err = row.ScanStruct(&s); err != nil {
					return err // generally scan error not retryable, return it for driver check error
				}
			}
		}

		return nil
	},
	query.WithIdempotent(),
)
if err != nil {
    log.Fatal(err)
}
```

* usage with `database/sql` (see additional docs in [SQL.md](SQL.md) )
```go
import (
    "context"
    "database/sql"
    "log"

    _ "github.com/ydb-platform/ydb-go-sdk/v3"
)

...

db, err := sql.Open("ydb", "grpc://localhost:2136/local")
if err != nil {
    log.Fatal(err)
}
defer db.Close() // cleanup resources
var (
    id    int32
    myStr string
)
row := db.QueryRowContext(context.TODO(), `SELECT 42 as id, "my string" as myStr`)
if err = row.Scan(&id, &myStr); err != nil {
    log.Printf("select failed: %v", err)
    return
}
log.Printf("id = %d, myStr = \"%s\"", id, myStr)
```


More examples of usage placed in [examples](./examples) directory.

## Credentials <a name="credentials"></a>

Driver implements several ways for making credentials for `YDB`:
- `ydb.WithAnonymousCredentials()` (enabled by default unless otherwise specified)
- `ydb.WithAccessTokenCredentials("token")`
- `ydb.WithStaticCredentials("user", "password")`, 
- `ydb.WithOauth2TokenExchangeCredentials()` and `ydb,WithOauth2TokenExchangeCredentialsFile(configFilePath)`
- as part of connection string, like as `grpcs://user:password@endpoint/database`

Another variants of `credentials.Credentials` object provides with external packages:

| Package                                                                            | Type        | Description                                    | Link of example usage                                                                                                                                                                                                                                                                                                                                              |
|------------------------------------------------------------------------------------|-------------|------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [ydb-go-yc](https://github.com/ydb-platform/ydb-go-yc)                             | credentials | credentials provider for Yandex.Cloud          | [yc.WithServiceAccountKeyFileCredentials](https://github.com/ydb-platform/ydb-go-yc/blob/master/internal/cmd/connect/main.go#L22) [yc.WithInternalCA](https://github.com/ydb-platform/ydb-go-yc/blob/master/internal/cmd/connect/main.go#L22) [yc.WithMetadataCredentials](https://github.com/ydb-platform/ydb-go-yc/blob/master/internal/cmd/connect/main.go#L24) |
| [ydb-go-yc-metadata](https://github.com/ydb-platform/ydb-go-yc-metadata)           | credentials | metadata credentials provider for Yandex.Cloud | [yc.WithInternalCA](https://github.com/ydb-platform/ydb-go-yc-metadata/blob/master/options.go#L23) [yc.WithCredentials](https://github.com/ydb-platform/ydb-go-yc-metadata/blob/master/options.go#L17)                                                                                                                                                             |
| [ydb-go-sdk-auth-environ](https://github.com/ydb-platform/ydb-go-sdk-auth-environ) | credentials | create credentials from environ                | [ydbEnviron.WithEnvironCredentials](https://github.com/ydb-platform/ydb-go-sdk-auth-environ/blob/master/env.go#L11)                                                                                                                                                                                                                                                |

## Ecosystem of debug tools over `ydb-go-sdk` <a name="debug"></a>

Package `ydb-go-sdk` provide debugging over trace events in package `trace`.
Next packages provide debug tooling:

| Package                                                                          | Type    | Description                                                                                                               | Link of example usage                                                                                                          |
|----------------------------------------------------------------------------------|---------|---------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------|
| [ydb-go-sdk-zap](https://github.com/ydb-platform/ydb-go-sdk-zap)                 | logging | logging ydb-go-sdk events with `zap` package                                                                                | [ydbZap.WithTraces](https://github.com/ydb-platform/ydb-go-sdk-zap/blob/master/internal/cmd/bench/main.go#L64)                 |
| [ydb-go-sdk-zerolog](https://github.com/ydb-platform/ydb-go-sdk-zerolog)             | logging | logging ydb-go-sdk events with `zerolog` package                                                                            | [ydbZerolog.WithTraces](https://github.com/ydb-platform/ydb-go-sdk-zerolog/blob/master/internal/cmd/bench/main.go#L47)         |
| [ydb-go-sdk-logrus](https://github.com/ydb-platform/ydb-go-sdk-logrus)             | logging | logging ydb-go-sdk events with `logrus` package                                                                            | [ydbLogrus.WithTraces](https://github.com/ydb-platform/ydb-go-sdk-logrus/blob/master/internal/cmd/bench/main.go#L48)         |
| [ydb-go-sdk-prometheus](https://github.com/ydb-platform/ydb-go-sdk-prometheus/v2)   | metrics | prometheus wrapper over [ydb-go-sdk/v3/metrics](https://github.com/ydb-platform/ydb-go-sdk/tree/master/metrics)            | [ydbPrometheus.WithTraces](https://github.com/ydb-platform/ydb-go-sdk-prometheus/blob/master/internal/cmd/bench/main.go#L56)   |
| [ydb-go-sdk-opentracing](https://github.com/ydb-platform/ydb-go-sdk-opentracing) | tracing | OpenTracing plugin for trace internal ydb-go-sdk calls                                                                    | [ydbOpentracing.WithTraces](https://github.com/ydb-platform/ydb-go-sdk-opentracing/blob/master/internal/cmd/bench/main.go#L86) |
| [ydb-go-sdk-otel](https://github.com/ydb-platform/ydb-go-sdk-otel) | tracing | OpenTelemetry plugin for trace internal ydb-go-sdk calls                                                                    | [ydbOtel.WithTraces](https://github.com/ydb-platform/ydb-go-sdk-otel/blob/master/internal/cmd/bench/main.go#L98) |

## Environment variables <a name="environ"></a>

`ydb-go-sdk` supports next environment variables  which redefines default behavior of driver

| Name                             | Type      | Default | Description                                                                                                              |
|----------------------------------|-----------|---------|--------------------------------------------------------------------------------------------------------------------------|
| `YDB_SSL_ROOT_CERTIFICATES_FILE` | `string`  |         | path to certificates file                                                                                                |
| `YDB_LOG_SEVERITY_LEVEL`         | `string`  | `quiet` | severity logging level of internal driver logger. Supported: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `quiet` |
| `YDB_LOG_DETAILS`                | `string`  | `.*`    | regexp for lookup internal logger logs                                                                                   |
| `GRPC_GO_LOG_VERBOSITY_LEVEL`    | `integer` |         | set to `99` to see grpc logs                                                                                             |
| `GRPC_GO_LOG_SEVERITY_LEVEL`     | `string`  |         | set to `info` to see grpc logs                                                                                           |
