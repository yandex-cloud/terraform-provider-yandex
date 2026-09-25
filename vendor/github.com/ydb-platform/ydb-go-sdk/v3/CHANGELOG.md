## v3.115.7
* Added support for `PartitionBy` in `DescribeTable` results

## v3.115.6
* Fixed context cancellation issues in the `QueryService` stream results

## v3.115.5
* Fixed error logging in `topic.UpdateOffsetsInTransaction`

## v3.115.4
*  Supported reusing one IP address and host name after nodes restart

## v3.115.3
* Fixed error catching in `retry.DoTxWithResult`

## v3.115.2
* Fixed bug with wrong literal YQL representation for signed integer YDB types

## v3.115.1
* Added `ydb.Param.Range()` range iterator

## v3.115.0
* Added public package `pkg/xtest` with test helpers

## v3.114.1
* Fixed depth for `pkg/xerrors.WithStackTrace()` error 

## v3.114.0
* Added public packages:
  - `pkg/xerrors` - helpers for wrap errors with stacktrace and join errors
  - `pkg/xslices` - helpers for work with slices
  - `pkg/xstring` - helpers for work with strings
* Added `table.Client.DescribeTable()` helper method
* Added `table/types/IsNull()` and `table/types/Unwrap()` methods for work with optional values

## v3.113.6
* Removed experimental label from `retry.DoWithResult` and `retry.DoTxWithResult`

## v3.113.5
* Fixed incorrect string conversion of `Uint64Value` values greater than `int64` max value

## v3.113.4
* Added original error to `listValue` and `setValue` cast error

## v3.113.3
* Marked as non-retryable operation error `ABORTED` with internal issue `200509` ("Datashard program size limit exceeded")

## v3.113.2
* Fixed panic in `database/sql` traces

## v3.113.1
* Fixed several issues in `database/sql` error handling
* Improved context management for query result streams.

## v3.113.0
* Added experimental `config.WithDisableOptimisticUnban` option to disable fast node unban after pessimization
* Fixed respect start offset from `topicoptions.WithReaderGetPartitionStartOffset` for commit messages
* Added `query.AllowImplicitSessions()` option for execute queries through `query.Client.{Exec,Query,QueryResultSet,QueryRow}` without explicit sessions

## v3.112.0
* Added support for the `json.Unmarshaler` interface in the `CastTo` function for use in scanners, such as the `ScanStruct` method
* Fixed the support of server-side session balancing in `database/sql` driver
* Added `ydb.WithDisableSessionBalancer()` driver option for disable server-side session balancing on table and query clients

## v3.111.3
* Fixed session closing in `ydb.WithExecuteDataQueryOverQueryClient(true)` scenario

## v3.111.2
* Changed discovery and dns resolving log level to DEBUG

## v3.111.1
* Replaced minimal requirements of `go` from `1.23` to `1.22`
* Migrated `golangci-lint` to version `v2.1.6`

## v3.111.0
* Added `sugar.PrintErrorWithoutStack` helper for remove stack records from error string
* Added `sugar.UnwrapError` helper for unwrap source error to root errors

## v3.110.1
* Added the ability to send BulkRequest exceeding the GrpcMaxMessageSize

## v3.110.0
* Added read partitions in parallel for topic listener.

## v3.109.0
* Added control plane fields for split-merge topics (Create,Alter,Describe) 

## v3.108.5
* Fixed stop topic reader after TLI in transaction

## v3.108.4
* Removed `experimental` from coordination API
* Added `WithReaderLogContext`, `WithWriterLogContext` options to topic reader/writer to supply log entries with user context fields

## v3.108.3
* Fixed handling of zero values for DyNumber
* Fixed the decimal yql slice bounds out of range

## v3.108.2
* Bumped dependencies:
  - `github.com/jonboulle/clockwork` from v0.3.0 to v0.5.0
  - `google.golang.org/grpc` from v1.62.1 to v1.69.4
  - `google.golang.org/protobuf` from v1.33.0 to v1.35.1
  - `github.com/golang-jwt/jwt/v4` from v4.5.0 to v4.5.2 (see security alerts https://github.com/golang-jwt/jwt/security/advisories/GHSA-mh63-6h87-95cp and https://github.com/ydb-platform/ydb-go-sdk/security/dependabot/45)
  - `golang.org/x/net` from v0.33.0 to v0.38.0 (see security alert https://github.com/ydb-platform/ydb-go-sdk/security/dependabot/59)
  - `golang.org/x/crypto` from 0.31.0 to 0.36.0 (see security alerts https://github.com/ydb-platform/ydb-go-sdk/security/dependabot/56 and https://github.com/ydb-platform/ydb-go-sdk/security/dependabot/52)
  - `golang.org/x/sync` from v0.10.0 to v0.12.0
  - `golang.org/x/sys` from v0.28.0 to v0.31.0
  - `golang.org/x/text` from v0.21.0 to v0.23.0

## v3.108.1
* Supported `json.Marshaller` query parameter in `database/sql` driver

## v3.108.0
* Added `query.EmptyTxControl()` for empty transaction control (server-side defines transaction control by internal logic)
* Marked as deprecated `query.NoTx()` because this is wrong name for server-side transaction control inference

## v3.107.0
* Refactored internal client balancer: added singleton for getting gRPC-connection (auto dial and auto reconnect on non-ready state) for use in discovery attempts
* Added `topicoptions.IncludePartitionStats()` for `Topic().Describe()` in order to get partition stats from server

## v3.106.1
* Dropped `internal/allocator` package and all usages of it for further switch (test) protobuf opaque API

## v3.106.0
* Added option WithReaderSupportSplitMergePartitions for topic manage support of split-merge partitions on client side (enabled by default).
* Allowed overflow queue limit for one goroutine at time for topic writer
* Removed delay before send commit in sync mode of a topic reader

## v3.105.2
* Improved the `ydb.WithSessionPoolSessionUsageLimit` option for allow `time.Duration` as argument type for limit max session time to live since create time 

## v3.105.1
* Changed the gRPC DNS balancer policy to `round_robin` for internal `discovery/ListEndpoints` call (reverted v3.90.2 changes)

## v3.105.0
* Supported topic split merge server feature for topic reader (no api changed)

## v3.104.7
* Added public type alias `ydb.Params` to `internal/params.Parameters` for external usage

## v3.104.6
* Refactored `table.TransactionControl` and `query.TransactionControl` for use single implementation in `internal/tx`
* Changed `ydb.WithTxControl` context modifier for allow both `table.TransactionControl` and `query.TransactionControl`

## v3.104.5
* Added query client session pool metrics: create_in_progress, in_use, waiters_queue
* Added pool item closing for not-alived item

## v3.104.4
* Fixed bug with session query latency metric collector

## v3.104.3
* Changed argument types in `table.Client.ReadRows` to public types for compatibility with mock-generation 

## v3.104.2
* Added bindings options into `ydb.ParamsFromMap` for bind wide time types
* Changed `ydb.WithWideTimeTypes(bool)` for allow boolean argument

## v3.104.1
* Added export of advanced metric information for QueryService calls

## v3.104.0
* Added binding `ydb.WithWideTimeTypes()` which interprets `time.Time` and `time.Duration` as `Timestamp64` and `Interval64` YDB types

## v3.103.0
* Supported wide `Interval64` type

## v3.102.0
* Supported wide `Date32`, `Datetime64` and `Timestamp64` types

## v3.101.4
* Switched internal type of result `ydb.Driver.Query()` from `*internal/query.Client` to `query.Client` interface 

## v3.101.3
* Added `query.TransactionActor` type alias to `query.TxActor` for compatibility with `table.Client` API's 
* Removed comment `experimental` from `ydb.ParamsBuilder` and `ydb.ParamsFromMap`
* Fixed panic on closing `internal/query/sessionCore.done` channel twice
* Fixed hangup when try to send batch of messages with size more, then grpc limits from topic writer internals

## v3.101.2
* Added a new metric `ydb_go_sdk_ydb_info` with the current version of the SDK

## v3.101.1
* Changed allowBanned=false for preferred node connections

## v3.101.0
* Added `table.Client.ReadRows` method with internal retries

## v3.100.3
* Fixed bug with concurrent rewrites source slice of `grpc.DialOption` on dial step

## v3.100.2
* Fixed bug in `internal/xcontext.WithDone` (not listening chan done)

## v3.100.1
* Refactored behaviour on `retry.Retryable` error for retry object (such as session, connection or transaction)

## v3.100.0
* Added `table.DescribeTable.StoreType` to table description result from `table.Session.DescribeTable` request

## v3.99.13
* Added checking errors for conditionally delete item from pool

## v3.99.12
* Internal debug improved

## v3.99.11
* Added stacktrace record to row scan errors for detect broken client code
* Fixed DescribeConsumer ignoring PartitionConsumerStats
* Added virtualtimestamps field to cdc description

## v3.99.10
* Returned legacy behaviour for interpret as `time.Time` YDB types `Date`, `Datetime` and `Timestamp` 

## v3.99.9
* Fixed broken compatibility `database/sql` driver which worked on query engine (usnig `ydb.WithQueryService(true)` connector option):
  - fixed list of valid data types for `database/sql.Row.Scan()`
  - allowed legacy option `ydb.WithTxControl(ctx, txControl)` for query engine

## v3.99.8
* Added details to all log messages
* Fixed sometime panic on stats receive in query service

## v3.99.7
* Fixed not passing request context to topic event logs
* Fixed deadlock on closing table session with internal query session core

## v3.99.6
* Added log grpc messages metadata on trace log level for topic writer

## v3.99.5
* Fixed error `Empty query text` using prepared statements and `ydb.WithExecuteDataQueryOverQueryClient(true)` option
* Prepared statements always send query text on Execute call from now (previous behaviour - send query ID)  
* Prevented create decoder instance until start read a message from topics

## v3.99.4
* Fixed bug with wrong context on session closing
* Fixed goroutine leak on closing `database/sql` driver
* "No endpoints" is retriable error now

## v3.99.3
* Fixed potential infinity loop for local dc detection (CWE-835)
* Fixed nil pointer dereferenced in a topic listener (CWE-476)

## v3.99.2
* Fixed panic when error returned from parsing sql params
* Fixed explicit null dereferenced issue in internal/credentials/static.go (CWE-476)

## v3.99.1
* Bumped dependencies:
  - `golang.org/x/net` from v0.23.0 to v0.33.0
  - `golang.org/x/sync` from v0.6.0 to v0.10.0
  - `golang.org/x/sys` from v0.18.0 to v0.28.0
  - `golang.org/x/text` from v0.14.0 to v0.21.0
  - `github.com/golang-jwt/jwt/v4` from v4.4.1 to v4.5.0

## v3.99.0
* Added `ydb.WithExecuteDataQueryOverQueryClient(bool)` option to execute data queries from table service 
  client using query client API. Using this option you can execute queries from legacy table service client 
  through `table.Session.Execute` using internal query client API without limitation of 1000 rows in response.
  Be careful: an OOM problem may happen because bigger result requires more memory

## v3.98.0
* Supported pool of encoders, which implement ResetableWriter interface

## v3.97.0
* Added immutable range iterators from go1.23 into query stats to iterate over query phases and accessed tables without query stats object mutation

## v3.96.2
* Fixed broken metric `ydb_go_sdk_ydb_database_sql_conns`

## v3.96.1
* Fixed drop session from pool unnecessary in query service 

## v3.96.0
* Supported of list, set and struct for unmarshall using `sugar.Unmarshall...`

## v3.95.6
* Fixed panic on span reporting in `xsql/Tx`

## v3.95.5
* Fixed goroutine leak on failed execute call in query client

## v3.95.4
* Fixed connections pool leak on closing sessions
* Fixed an error in logging session deletion events

## v3.95.3
* Supported of `database/sql/driver.Valuer` interfaces for params which passed to query using sql driver 
* Exposed `credentials/credentials.OAuth2Config` OAuth2 config

## v3.95.2
* Fixed panic on multiple closing driver

## v3.95.1
* Added alias from `ydb.WithFakeTx(ydb.ScriptingQueryMode)` to `ydb.WithFakeTx(ydb.QueryExecuteQueryMode)` for compatibility with legacy code   

## v3.95.0
* Added implementation of `database/sql` driver over query service client
* Added `ydb.WithQueryService(bool)` option to explicitly enable `database/sql` driver over query service client
* Added environment parameter `YDB_DATABASE_SQL_OVER_QUERY_SERVICE` to enable `database/sql` driver over query service client without code rewriting

## v3.94.0
* Refactored golang types mapping into ydb types using `ydb.ParamsFromMap` and `database/sql` query arguments
* Small breaking change: type mapping for `ydb.ParamsFromMap` and `database/sql` type `uuid.UUID` changed from ydb type `Text` to ydb type `UUID`

## v3.93.3
* Supported raw protobuf `*Ydb.TypedValue` using `ydb.ParamsBuilder()`

## v3.93.2
* Removed experimental helper `ydb.MustParamsFromMap`
* Changed result of experimental helper `ydb.ParamsFromMap` from tuple <`params.Parameters`, `error`> to `params.Parameters` only 

## v3.93.1
* Published `query.ExecuteOption` as alias to `internal/query/options.Execute`

## v3.93.0
* Added `ydb.WithStaticCredentialsLogin` and `ydb.WithStaticCredentialsPassword` options

## v3.92.6
* Fixed string representation of `TzTimestamp`, `TzDatetime` and `TzDate` type values
* Added `database/sql/driver.Value` as type destination for almost ydb values

## v3.92.5
* Avoid retrying requests finished with `UNAUTHORIZED` errors

## v3.92.4
* Fixed connections pool leak on closing

## v3.92.3
* Fixed error with incompleted data return from transaction.ReadQueryResult method
* Added option `query/WithResponsePartLimitSizeBytes(...)` for queries with query service

## v3.92.2
* Added `table/options.WithShardNodesInfo()` experimental option to get shard nodeId for describe table call

## v3.92.1
* Added `sugar.WithUserPassword(user,password)` option for `sugar.DSN()` helper
* Added `sugar.WithSecure(bool)` option for `sugar.DSN()` helper
* Small breaking change: `sugar.DSN` have only two required parameters (endpoint and database) from now on. 
  Third parameter `secure` must be passed as option `sugar.WithSecure(bool)`

## v3.92.0
* Added experimental ydb.ParamsFromMap and ydb.MustParamsFromMap for build query parameters
* Refactored coordination traces
* gRPC connection will be forcefully closed on DNS resolver errors from now on

## v3.91.0
* Added `ydb.WithPreferredNodeID(ctx, nodeID)` context modifier for trying to execute queries on given nodeID

## v3.90.2
* Set the `pick_first` balancer for short-lived grpc connection inside ydb cluster discovery attempt

## v3.90.1
* Small broken change: added method `ID()` into `spans.Span` interface (need to implement in adapter) 
* Fixed traceparent header for tracing grpc requests

## v3.90.0
* Fixed closing of child driver with shared balancer

## v3.89.6
* Refactored `database/sql` driver internals for query-service client support in the future 

## v3.89.5
* Fixed nil pointer dereference in metabalancer initialization

## v3.89.4
* Changed behaviour on re-discovery: always open new grpc connection for discovery request

## v3.89.3
* Wrapped internal balancer with metadata middleware

## v3.89.2
* Returned log.XXX methods for create fields, removed from public at v3.85.0

## v3.89.1
* Added option `ydb.WithSharedBalancer(*Driver)` for child drivers

## v3.89.0
* Fixed send optional arguments to the server with `ydb.ParamsBuilder`

## v3.88.0
* Removed UUID methods from ydb.ParamsBuilder()

## v3.87.0
* BREAK OLD STYLE WORK WITH UUID. See https://github.com/ydb-platform/ydb-go-sdk/issues/1501 for details.
  At the version you must explicit choose way for work with uuid: old with bug or new (fixed).

## v3.86.1
* Fixed scan to optional uuid

## v3.86.0
* Add workaround for bug in uuid send/receive from server. It is migration version. All native code and most database sql code worked with uuid continue to work.
Dedicated version for migrate code for workaround/fix uuid bug. See https://github.com/ydb-platform/ydb-go-sdk/issues/1501 for details.

## v3.85.3
* Renamed `query.WithPoolID()` into `query.WithResourcePool()`

## v3.85.2
* Added experimental `query.WithPoolID()` execute option for define resource pool for execute query

## v3.85.1
* Added `spans.Retry` constructor of `trace.Retry`

## v3.85.0
* Added experimental package `spans` with tracing adapter interfaces for OpenTelemetry, OpenTracing, etc.
* Added `db.Topic().DescribeTopicConsumer()` method for displaying consumer information
* Marked as deprecated options `ydb.WithDatabase(database)` and `ydb.WithEndpoint(endpoint)`

## v3.84.1
* Added session info into `trace.TableSessionBulkUpsertStartInfo`

## v3.84.0
* Added `meta.WithTraceParent` context modifier for explicit putting traceparent header into grpc calls

## v3.83.0
* Supported `db.Table().BulkUpsert()` from scv, arrow and ydb rows formats

## v3.82.0
* Fixed error on experimental `TopicListener.Close`
* Disabled reporting of `ydb_go_sdk_query_session_count` when metrics are disabled
* Disabled reporting of `ydb_go_sdk_ydb_query_session_create_latency` histogram metrics when metrics are disabled
* Allowed skip column for `ScanStruct` by tag `-`

## v3.81.4
* Returned `topicwriter.ErrQueueLimitExceed`, accidental removed at `v3.81.0`

## v3.81.3
* Fixed tracing details check for some metrics

## v3.81.2
* Removed `experimantal` comment for query service client

## v3.81.1
* Fixed nil pointer dereference panic on failed `ydb.Open`
* Added ip discovery. Server can show own ip address and target hostname in the ListEndpoint message. These fields are used to bypass DNS resolving.

## v3.81.0
* Added error ErrMessagesPutToInternalQueueBeforeError to topic writer
* Added write to topics within transactions

## v3.80.10
* Added `ydb.WithSessionPoolSessionUsageLimit()` option for limitation max count of session usage
* Refactored experimental topic iterators in `topicsugar` package

## v3.80.9
* Fixed bug in experimental api: `ydb.ParamsBuilder().Param().Optional()` receive pointer and really produce optional value.

## v3.80.8
* Added `ydb.WithLazyTx(bool)` option for create lazy transactions on `query.Session.Begin` call
* Added initial experimental topic and cdc-helpers, see examples in [tests/integration/topic_helpers_test.go](https://github.com/ydb-platform/ydb-go-sdk/blob/master/tests/integration/topic_helpers_test.go)
* Added experimental `sugar.UnmarshalRows` for user unmarshaller structs in own code in go 1.23, change example for use the iterator.
* Added `ydb_go_sdk_ydb_query_pool_size_index` metrics

## v3.80.7
* Fixed bug with doesn't rollback the transaction on the operation error in table service

## v3.80.6
* Fixed concurrent map writes in metrics
* Renamed method at experimental API `reader.PopBatchTx` to `reader.PopMessagesBatchTx`

## v3.80.5
* Fixed connections pool leak on failed `ydb.Open` call

## v3.80.4
* Fixed panic on usage metrics package from prometheus adapter on `trace.Driver.OnNewStream` callback

## v3.80.3
* Added option `ydb.WithSessionPoolSessionIdleTimeToLive` for restrict idle time of query sessions
* Fixed bug with leak of query transactions
* Changed `ydb_go_sdk_ydb_driver_conn_requests` metrics splitted to `ydb_go_sdk_ydb_driver_conn_request_statuses` and `ydb_go_sdk_ydb_driver_conn_request_methods`
* Fixed metadata for operation service connection
* Fixed composing query traces in call `db.Query.Do[Tx]` using option `query.WithTrace`

## v3.80.2
* Added `balancers.PreferNearestDC[WithFallback]` balancers
* Marked as deprecated `balancers.PreferLocalDC[WithFallback]` balancers because `local` word is ambiguous for balancer idea

## v3.80.1
* Added `lastErr` from previous attempt in `retry.RetryWithResult`

## v3.80.0
* Replaced internal table client pool entities to `internal/pool`

## v3.79.2
* Enabled by default usage of `internal/pool` in `internal/query.Client`

## v3.79.1
* Changed `trace.Table` and `trace.Query` traces
* Implemented `internal/pool` the same as table client pool from `internal/table.Client`

## v3.79.0
* Added commit messages for topic listener
* EOF error in RecvMsg is no longer logged

## v3.78.0
* Changed result type of method `query.Executor.QueryResultSet` from `query.ResultSet` to `query.ClosableResultSet`
* Added `table/types.DecimalValueFromString` decimal type constructor

## v3.77.1
* Added log topic writer ack
* Replaced `operation.Client.List` to five methods for listing operations `operation.List{BuildIndex,ImportFromS3,ExportToS3,ExportToYT,ExecuteQuery}`

## v3.77.0
* Changed log message about send topic message
* Added experimental support for executing scripts over query service client (`query.Client.ExecuteScript` and `query.CLient.FetchScriptResults`)
* Removed tx result from `query.Session.Execute` (tx can be obtained from `query.Session.Begin`)
* Changed behaviour of `query.Session.Begin` to `noop` for lazy initialization with first call `query.TxActor.Execute`
* Splitted experimental method `query.Client.Execute` to methods `query.Client.Exec` without result and `query.Client.Query` with result
* Splitted experimental method `query.TxActor.Execute` to methods `query.TxActor.Exec` without result and `query.TxActor.Query` with result
* Renamed experimental method `query.Client.ReadResultSet` to `query.Client.QueryResultSet`
* Renamed experimental method `query.Client.ReadRow` to `query.Client.QueryRow`
* Removed experimental methods `query.Session.ReadResultSet` and  `query.Session.ReadRows`
* Removed experimental methods `query.TxActor.ReadResultSet` and  `query.TxActor.ReadRows`
* Removed experimental method `query.Client.Stats`
* Option `query.WithIdempotent()` allowed for `query.Client.{Exec,Query,QueryResultSet,QueryRow}` methods now
* Added experimental support for operation service client through `db.Operation()` method (supports methods `Get`, `List`, `Cancel` and `Forget`)

## v3.76.6
* Replaced requirements from go1.22 + experimantal flag to go1.23 for experimental range-over interface

## v3.76.5
* Fixed out of index item creation in `internal/pool.Pool`
* Fixed tracing of `(*grpcClientStream).finish` event

## v3.76.4
* Added traces and logs for read messages from topic within transaction
* Changed result type of `query.Session.NodeID()` from `int64` to `uint32` for compatibility with table session and discovery
* Removed experimental method `query.Result.Err()`
* Added the finishing reading the grpc stream on `query.Result.Close()` call
* Renamed experimental method `query.Result.Range()` to `query.Result.ResultSets()`
* Renamed experimental method `query.ResultSet.Range()` to `query.ResultSet.Rows()`
* Removed support of `go1.20`
* Added PopMessages from topic within transaction

## v3.76.3
* Changed interface `table.TransactionIdentifier` (added private method) for prohibition of any implementations outside ydb-go-sdk

## v3.76.2
* Fixed bug with nil pointer dereference on trace callback from `query.createSession`
* Fixed test message builder, now all method return itself pointer
* Fixed handle reconnection timeout error
* Fixed experimental topic listener handle stop partition event

## v3.76.1
* Fixed `query.WithCommit()` flag behaviour for `tx.Execute` in query service
* OAuth 2.0 token exchange: allowed multiple resource parameters in according to https://www.rfc-editor.org/rfc/rfc8693

## v3.76.0
* Added experimental topic listener implementation
* Fixed `internal/xstrings.Buffer()` leak without call `buffer.Free()`
* Removed double quotas from goroutine labels background workers for prevent problem with pprof

## v3.75.2
* Fixed build for go1.20

## v3.75.1
* Fixed return more than one row error if real error raised on try read next row
* Fixed checking errors for session must be deleted
* Changed signature of filter func in balancers (replaced argument from `conn.Conn` type to `endpoint.Info`)

## v3.75.0
* Improve config validation before start topic reader
* Added metrics over `db.Table().Do()` and `db.Table().DoTx()`
* Added method `ydb.ParamsBuilder().Param(name).Any(value)` to add custom `types.Value`
* Upgraded dependencies:
  * `google.golang.org/grpc` - from `v1.57.1` to `v1.62.1`
  * `github.com/google/uuid` - from `v1.3.0` to `v1.6.0`
  * `golang.org/x/sync` - from `v0.3.0` to `v0.6.0`
* Fixed goroutine leak on close reader
* Fixed topic reader and writer WaitInit hunging on unretriable connection error
* Added `query.Client.Stats()` method
* Added `query.Result.Stats()` method
* Added `query.ResultSet.Index()` method
* Support loading OAuth 2.0 token exchange credentials provider from config file
* Added options for JWT tokens for loading EC private keys and HMAC secrets
* Add retries to OAuth 2.0 token exchange credentials

## v3.74.5
* Fixed bug with reading empty result set parts.
* Fixed nil pointer dereference when closing result set

## v3.74.4
* Fixed bug with fail cast of grpc response to `operation.{Response,Status}`

## v3.74.3
* Removed check the node is available for query and table service sessions
* Refactored the `balancers.PreferLocations()` function - it is a clean/pure function
* Added experimental `balancers.WithNodeID()` context modifier for define per request the YDB endpoint by NodeID
* Reverted the allowing the casts from signed YDB types to unsigned destination types if source value is not negative
* Replaced internal query session pool by default to stub for exclude impact from internal/pool

## v3.74.2
* Added description to scan errors with use query service client scanner

## v3.74.1
* Allowed the use of DSN without specifying the protocol/scheme
* Allowed casts from signed YDB types to unsigned destination types if source value is not negative
* Removed public `query.TxIdentifier` interface for exclude any external implementations for use with YDB

## v3.74.0
* Added experimental range functions to the `query.Result` and `query.ResultSet` types, available as for-range loops starting with Go version 1.22. These features can be enabled by setting the environment variable `GOEXPERIMENT=rangefunc`.
* Added public types for `tx.Option`, `options.DoOption` and `options.DoTxOption`

## v3.73.1
* Changed `query.DefaultTxControl()` from `query.SerializableReadWrite()` with commit to `query.NoTx()`

## v3.73.0
* Added experimental `retry.DoWithResult` and `retry.DoTxWithResult` helpers for retry lambda and return value from lambda

## v3.72.0
* Excluded `Query()` method from interface `ydb.Connection`. Method `Query()` remains accessible from `ydb.Driver`

## v3.71.0
* Added `query/ResultSet.{Columns,ColumnTypes}` methods for get column names and types from query result set
* Added experimental `retry.RetryWithResult` helper for retry lambda and return value from lambda

## v3.70.0
* Fixed `config.WithDatabase` behaviour with empty database in DSN string
* Added experimental method `query/Client.Execute` for execute query and read materialized result

## v3.69.0
* Added experimental method for execute query and read only one row from result:
  * `query/Client.ReadRow`
  * `query/Session.ReadRow`
  * `query/Transaction.ReadRow`
* Added experimental method for execute query and read only one result set from result:
  * `query/Client.ReadResultSet`
  * `query/Session.ReadResultSet`
  * `query/Transaction.ReadResultSet`
* Added experimental `sugar.UnmarshallRow[T]` and `sugar.UnmarshallResultSet[T]` helpers for converts YDB rows to typed objects

## v3.68.1
* Downgraded minimal version of Go to 1.20
* Refactored internal packages by `ifshort` linter issues

## v3.68.0
* Added experimental `ydb.{Register,Unregister}DsnParser` global funcs for register/unregister external custom DSN parser for `ydb.Open` and `sql.Open` driver constructor
* Simple implement option WithReaderWithoutConsumer
* Fixed bug: topic didn't send specified partition number to a server

## v3.67.2
* Fixed incorrect formatting of decimal. Implementation of decimal has been reverted to latest working version

## v3.67.1 (retracted)
* Fixed race of stop internal processes on close topic writer
* Fixed goroutines leak within topic reader on network problems

## v3.67.0
* Added `ydb.WithNodeAddressMutator` experimental option for mutate node addresses from `discovery.ListEndpoints` response
* Added type assertion checks to enhance type safety and prevent unexpected panics in critical sections of the codebase

## v3.66.3
* Fixed the OAuth2 test

## v3.66.2
* Added `trace.DriverConnStreamEvents` details bit
* Added `trace.Driver.OnConnStreamFinish` event

## v3.66.1
* Added flush messages from buffer before close topic writer
* Added Flush method for topic writer

## v3.66.0
* Added experimental package `retry/budget` for limit second and subsequent retry attempts
* Refactored internals for enabling `containedctx` linter
* Fixed the hanging semaphore issue on coordination session reconnect

## v3.65.3
* Fixed data race in `internal/conn.grpcClientStream`

## v3.65.2
* Fixed data race using `log.WithNames`

## v3.65.1
* Updated dependency `ydb-go-genproto`
* Added processing of `Ydb.StatusIds_EXTERNAL_ERROR` in `retry.Retry`

## v3.65.0
* Supported OAuth 2.0 Token Exchange credentials provider

## v3.64.0
* Supported `table.Session.RenameTables` method
* Fixed out of range panic if next query result set part is empty
* Updated the indirect dependencies `golang.org/x/net` to `v0.17.0` and `golang.org/x/sys` to `v0.13.0` due to vulnerability issue

## v3.63.0
* Added versioning policy

## v3.62.0
* Restored `WithSessionPoolKeepAliveMinSize` and `WithSessionPoolKeepAliveTimeout` for backward compatibility.
* Fixed leak timers
* Changed default StartTime (time of retries for connect to server) for topic writer from 1 minute to infinite (can be overrided by WithWriterStartTimeout topic option)
* Added `Struct` support for `Variant` in `ydb.ParamsBuilder()`
* Added `go` with anonymous function case in `gstack`

## v3.61.2
* Changed default transaction control to `NoTx` for execute query through query service client

## v3.61.1
* Renamed `db.Coordination().CreateSession()` to `db.Coordination().Session()` for compatibility with protos

## v3.61.0
* Added `Tuple` support for `Variant` in `ydb.ParamsBuilder()`

## v3.60.1
* Added additional traces for coordination service client internals

## v3.60.0
* Added experimental support of semaphores over coordination service client

## v3.59.3
* Fixed `gstack` logic for parsing `ast.BlockStmt`

## v3.59.2
* Added internal `gstack` codegen tool for filling `stack.FunctionID` with value from call stack

## v3.59.1
* Fixed updating last usage timestamp for smart parking of the conns

## v3.59.0
* Added `Struct` support for `ydb.ParamsBuilder()`
* Added support of `TzDate`,`TzDateTime`,`TzTimestamp` types in `ydb.ParamsBuilder()`
* Added `trace.Query.OnTransactionExecute` event
* Added query pool metrics
* Fixed logic of query session pool
* Changed initialization of internal driver clients to lazy
* Removed `ydb.WithSessionPoolSizeLimit()` option
* Added async put session into pool if external context is done
* Dropped intermediate callbacks from `trace.{Table,Retry,Query}` events
* Wrapped errors from `internal/pool.Pool.getItem` as retryable
* Disabled the logic of background grpc-connection parking
* Improved stringification for postgres types

## v3.58.2
* Added `trace.Query.OnSessionBegin` event
* Added `trace.Query.OnResult{New,NextPart,NextResultSet,Close}` events
* Added `trace.Query.OnRow{Scan,ScanNamed,ScanStruct}` events

## v3.58.1
* Dropped all deprecated callbacks and events from traces
* Added `trace.Driver.OnConnStream{SendMsg,RecvMsg,CloseSend}` events
* Added `trace.Query.OnSessionExecute` event

## v3.58.0
* Changed `List` constructor from `ydb.ParamsBuilder().List().Build().Build()` to `ydb.ParamsBuilder().BeginList().EndList().Build()`
* Changed `Set` constructor from `ydb.ParamsBuilder().Set().Build().Build()` to `ydb.ParamsBuilder().BeginSet().EndSet().Build()`
* Changed `Dict` constructor from `ydb.ParamsBuilder().Dict().Build().Build()` to `ydb.ParamsBuilder().BeginDict().EndDict().Build()`
* Changed `Optional` constructor from `ydb.ParamsBuilder().Set().Build().Build()` to `ydb.ParamsBuilder().BeginOptional().EndOptional().Build()`
* Added events into `trace.Query` trace
* Rewrote `internal/pool` to buffered channel
* Added `internal/xcontext.WithDone()`
* Added `internal/xsync.{OnceFunc,OnceValue}`
* Updated `google.golang.org/protobuf` from `v1.31.0` to `v.33.0`
* Added `ydb.ParamsBuilder().Pg().{Value,Int4,Int8,Unknown}` for postgres arguments
* Added `Tuple` support for `ydb.ParamsBuilder()`

## v3.57.4
* Added client pid to each gRPC requests to YDB over header `x-ydb-client-pid`
* Added `ydb.WithApplicationName` option
* Added `Dict` support for `ydb.ParamsBuilder()`

## v3.57.3
* Added metrics over query service internals
* Added session create and delete events into `trace.Query`
* Moved public type `query.SessionStatus` into `internal/query` package

## v3.57.2
* Fixed cases when some option is nil

## v3.57.1
* Added logs over query service internals
* Changed `trace.Query` events
* Changed visibility of `query.{Do,DoTx}Options` from public to private

## v3.57.0
* Added experimental implementation of query service client
* Fixed sometime panic on topic writer closing
* Added experimental query parameters builder `ydb.ParamsBuilder()`
* Changed types of `table/table.{QueryParameters,ParameterOption}` to aliases on `internal/params.{Parameters,NamedValue}`
* Fixed bug with optional decimal serialization

## v3.56.2
* Fixed return private error for commit to stopped partition in topic reader.
* Stopped wrapping err error as transport error at topic streams (internals)

## v3.56.1
* Fixed fixenv usage (related to tests only)

## v3.56.0
* Fixed handle of operational errors in topic streams
* The minimum version of Go in `ydb-go-sdk` has been raised to `go1.21`
* Fixed topic writer infinite reconnections in some cases
* Refactored nil on err `internal/grpcwrapper/rawydb/issues.go`, when golangci-lint nilerr enabled
* Refactored nil on err `internal/grpcwrapper/rawtopic/describe_topic.go`, when golangci-lint nilerr enabled

## v3.55.3
* Fixed handle of operational errors in topic streams (backported fix only)

## v3.55.2
* Fixed init info in topic writer, when autoseq num turned off.

## v3.55.1
* Supported column name prefix `__discard_column_` for discard columns in result sets
* Made `StatusIds_SESSION_EXPIRED` retriable for idempotent operations

## v3.55.0
* Refactored `internal/value/intervalValue.Yql()`
* The minimum version of Go in `ydb-go-sdk` has been raised to `go1.20`

## v3.54.3
* Added per message metadata support for topic api
* Context for call options now have same lifetime as driver (previous - same lifetime as context for call Open function).
* Extended metrics (fill database.sql callbacks, recognize TLI error)
* Refactored config prefix in metrics
* Removed excess status labels from metrics
* Implement `fmt.Stringer` interface for `Driver` struct

## v3.54.2
* Added context to some internal methods for better tracing
* Added `trace.FunctionID` helper and `FunctionID` field to trace start info's
* Replaced lazy initialization of ydb clients (table, topic, etc.) to explicit initialization on `ydb.Open` step

## v3.54.1
* Fixed inconsistent labels in `metrics`

## v3.54.0
* Allowed `sql.LevelSerializable` isolation level in read-write mode in `database/sql` transactions
* Refactored traces and metrics
* Added `{retry,table}.WithLabel` options for mark retriers calls
* Added `ydb.WithTraceRetry` option
* Moved `internal/allocator.Buffers` to package `internal/xstring`
* Bumped `golang.org/x/sync` to `v0.3.0`
* Bumped `google.golang.org/protobuf` to `v1.31.0`
* Bumped `google.golang.org/grpc` to `v1.57.1`
* Allowed grpc status error as arg in `internal/xerrors.TransportError(err)`
* Added `interanl/xtest.CurrentFileLine()` helper for table tests
* Added `internal/credentials.IsAccessError(err)` helper for check access errors
* Changed period for re-fresh static credentials token from `1/2` to `1/10` to expiration time
* Added `table.SnapshotReadOnlyTxControl()` helper for get transaction control with snapshot read-only

## v3.53.4
* Downgrade `golang.org/x/net` from `0.17.0` to `0.15.0`
* Downgrade `golang.org/x/sys` from `v0.13.0` to `v0.12.0`
* Downgrade `golang.org/x/crypto` from `v0.14.0` to `v0.13.0`

## v3.53.3
* Refactored credentials options (from funcs to interfaces and types)
* Fixed stringification of credentials object

## v3.53.2
* Fixed panic when try to unwrap values with more than 127 columns with custom ydb unmarshaler

## v3.53.1
* Bumps `github.com/ydb-platform/ydb-go-genproto` for support `query` service
* Bumps `golang.org/x/net` from `0.7.0` to `0.17.0`
* Bumps `golang.org/x/sys` from `v0.5.0` to `v0.13.0`
* Bumps `golang.org/x/text` from `v0.7.0` to `v0.13.0`

## v3.53.0
* Removed `internal/backoff.Backoff.Wait` interface method for exclude resource leak with bug-provoked usage of `time.After` method
* Marked as deprecated `retry.WithDoRetryOptions` and `retry.WithDoTxRetryOptions`
* Added receiving first result set on construct `internal/table/scanner.NewStream()`
* Added experimental package `metrics` with SDK metrics
* Fixed redundant trace call for finished `database/sql` transactions
* Added repeater event type to wake-up func context
* Refactored default logger format
* Refactored `internal/conn.coonError` format
* Fixed data race on `internal/conn.conn.cc` access

## v3.52.3
* Removed almost all experimental marks from topic api.
* Rename some topic APIs (old names was deprecated and will be removed in one of next versions).
* Deprecated topic options (the option will be removed): min size of read messages batch
* Deprecated WithOnWriterFirstConnected callback, use Writer.WaitInitInfo instead.
* Changed topic Codec base type from int to int32 (was experimental code)
* Added `WaitInit` and `WaitInitInfo` method to the topic reader and writer
* Remove extra allocations in `types.TupleValue`, `types.ListValue` and `types.SetValue`

## v3.52.2
* Removed support of placeholder "_" for ignoring columns in `database/sql` result sets

## v3.52.1
* Merged `internal/xsql/conn.{GetTables,GetAllTables}` methods for `DRY`
* Replaced `internal/xsql.Connector.PathNormalizer` default from `nopPathNormalizer` to `bind.TablePathPrefix` with database name as path prefix
* Supported placeholder "_" for ignored column names in `database/sql` result sets

## v3.52.0
* Added `table.Session.CopyTables` method
* Added `x-ydb-trace-id` header into grpc calls
* Improved topic reader logs
* Fixed `internal/xstring` package with deprecated warning in `go1.21` about `reflect.{String,Slice}Header`

## v3.51.3
* Added `internal/xstring.{FromBytes([]byte),ToBytes(string)` for increase performance on `string` from/to `[]byte` conversion

## v3.51.2
* Added `table/options.ReadFromSnapshot(bool)` option for `session.StreamReadTable()`

## v3.51.1
* Added checking condition for `tx.Rollback()` in `retry.DoTx`

## v3.51.0
* Added node info to grpc errors

## v3.50.0
* Added methods `TotalCPUTime()` and `TotalDuration()` to `table/stats/QueryStats` interface
* Added check if commit order is bad in sync mode

## v3.49.1
* Added `table.options.WithIgnoreTruncated` option for `session.Execute` method
* Added `table.result.ErrTruncated` error for check it with `errors.Is()` outside of `ydb-go-sdk`

## v3.49.0
* Added `table.Session.ReadRows` method for getting rows by keys
* Added `table/options.ChangefeedFormatDynamoDBStreamsJSON` format of `DynamoDB` change feeds

## v3.48.8
* Fixed `sugar.RemoveRecursive()` for column table type

## v3.48.7
* Added `sugar.StackRecord()` helper for stringification of current file path and line
* Updated `google.golang.org/grpc` from `v1.49.0` to `v1.53.0` due to vulnerability
* Updated `google.golang.org/protobuf` from `v1.28.0` to `v1.28.1` due to vulnerability
* Implemented implicit standard interface `driver.RowsColumnTypeNullable` in `internal/xsql.rows`
* Upgraded errors description from `retry.Retry` with attempts info

## v3.48.6
* Added builder for topic reader message (usable for tests)

## v3.48.5
* Removed `log.Secret` helper as unnessesarry in public API after refactoring logging subsystem
* Enriched the error with important details from initial discovery
* Added `internal.{secret,stack}` packages
* Implemented `fmt.Stringer` interface in credential types

## v3.48.4
* Added `ydb.IsOperationErrorTransactionLocksInvalidated(err)` helper for checks `TLI` flag in err

## v3.48.3
* Added `table/types.IsOptional()` helper

## v3.48.2
* Refactored tests

## v3.48.1
* Added `sugar.Is{Entry,ColumnTable}Exists` helper

## v3.48.0
* Fixed stopping topic reader by grpc stream shutdown
* Fixed `database/sql` driver for get and parse container ydb types
* Changed `table/scanner.scanner.Any()` behaviour: for non-primitive types returns raw `table/types.Value` instead nil from previous behaviour
* Added `table/types.{ListItems,VariantValue,DictValues}` helpers for get internal content of abstract `table/types.Value`
* Marked as deprecated `table/types.DictFields` (use `table/types.DictValues` instead)

## v3.47.5
* Added `scheme.Entry.IsColumnTable()` helper

## v3.47.4
* Disabled check of node exists with `balancers.SingleConn`
* Improved code with `go-critic` linter
* Added session info into `database/sql` event `connected`

## v3.47.3
* Added `table/options.Description.Tiering` field

## v3.47.2
* Refactored `internal/cmd/gtrace` tool (prefer pointers instead trace struct copies) for bust performance
* Fixed usage of generated traces in code

## v3.47.1
* Removed test artifacts from repository

## v3.47.0
* Added `table/types.ToDecimal()` converter from `table/types.Value` to `table/types.Decimal`

## v3.46.1
* Implemented `internal/xcontext.With{Cancel,Timeout}` with stack record and switched all usages from standard `context.With{Cancel,Timeout}`

## v3.46.0
* Refactored package `log` for support typed fields in log messages

## v3.45.0
* Added `table/options.WithPartitions` for configure partitioning policy
* Marked as deprecated `table/options.WithPartitioningPolicy{UniformPartitions,ExplicitPartitions}` (use `table/options.With{UniformPartitions,ExplicitPartitions}` instead)

## v3.44.3
* Fixed bug of processing endpoint with `node_id=0`
* Refactored of checking node ID in cluster discovery before `Get` and during in `Put` of session into session pool

## v3.44.2
* Removed debug print

## v3.44.1
* Fixed bug with returning session into pool before second discovery

## v3.44.0
* Added `table/options.WithCallOptions` options for append custom grpc call options into `session.{BulkUpsert,Execute,StreamExecuteScanQuery}`
* Supported fake transactions in `database/sql` driver over connector option `ydb.WithFakeTx(queryMode)` and connection string param `go_fake_tx`
* Removed `testutil/timeutil` package (all usages replaced with `clockwork` package)
* Changed behaviour of retryer on transport errors `cancelled` and `deadline exceeded` - will retry idempotent operation if context is not done
* Added address of node to operation error description as optional
* Fixed bug with put session from unknown node
* Fixed bug with parsing of `TzTimestamp` without microseconds
* Fixed code -1 of retryable error if wrapped error with code
* Added `ydb.MustOpen` and `ydb.MustConnector` helpers
* Fixed `internal/xerrors.Transport` error wrapping for case when given error is not transport error
* Added grpc and operation codes to errors string description
* Extend `scheme.Client` interface with method `Database`
* Removed `driver.ResultNoRows` in `internal/xsql`
* Added `ydb.{WithTablePathPrefix,WithAutoDeclare,WithPositionalArgs,WithNumericalArgs}` query modifiers options
* Supported binding parameters for `database/sql` driver over connector option `ydb.WithAutoBind()` and connection string params `go_auto_bind={table_path_prefix(path),declare,numeric,positional}`
* Added `testutil.QueryBind` test helper
* Fixed topic retry policy callback call: not call it with nil error
* Fixed bug with no checking operation error on `discovery.Client` calls
* Allowed zero create session timeout in `ydb.WithSessionPoolCreateSessionTimeout(timeout)` (less than or equal to zero - no used timeout on create session request)
* Added examples with own `go.mod`
* Marked as deprecated `ydb.WithErrWriter(w)` and `ydb.WithOutWriter(w)` logger options
* Added `ydb.WithWriter(w)` logger option

## v3.43.0
**Small broken changes**

Most users can skip there notes and upgrade as usual because build break rare used methods (expiremental API and api for special cases, not need for common use YDB) and this version has no any behavior changes.

Changes for experimental topic API:
* Moved `producer_id` from required positional argument to option `WithProducerID` (and it is optional now)
* Removed `WithMessageGroupID` option (because not supported now)

Changes in ydb connection:
* Publish internal private struct `ydb.connection` as `ydb.Driver` (it is implement `ydb.Connection`)
* `ydb.Connection` marked as deprecated
* Changed return type of `ydb.Open(...)` from `ydb.Connection` to `*ydb.Driver`
* Changed return type of `ydb.New(...)` from `ydb.Connection` to `*ydb.Driver`
* Changed argument type for `ydb.GRPCConn` from `ydb.Connection` to `*ydb.Driver`
* Removed method `With` from `ydb.Connection` (use `*Driver.With` instead).

Changes in package `sugar`:
* Changed a type of database arg in `sugar.{MakeRecursive,RemoveRecursive}` from `ydb.Connection` to minimal required local interface

Dependencies:
* Up minimal supported version of `go` to `1.17` for update dependencies (new `golang.org/x` doesn't compiled for `go1.16`)
* Upgrade `golang.org/x/...`  for prevent issues: `CVE-2021-33194`, `CVE-2022-27664`, `CVE-2021-31525`, `CVE-2022-41723`

## v3.42.15
* Fixed checking `nil` error with `internal/xerrors.Is`

## v3.42.14
* Supported `scheme.EntryTopic` path child entry in `sugar.RemoveRecursive`

## v3.42.13
* Fixed default state of `internal/xerrors.retryableError`: it inherit properties from parent error as possible
* Marked event `grpc/stats.End` as ignored at observing status of grpc connection

## v3.42.12
* Replaced the balancer connection to discovery service from short-lived grpc connection to `internal/conn` lazy connection (revert related changes from `v3.42.6`)
* Marked as deprecated `trace.Driver.OnBalancerDialEntrypoint` event callback
* Deprecated `trace.Driver.OnConnTake` event callback
* Added `trace.Driver.OnConnDial` event callback

## v3.42.11
* Fixed validation error for `topicoptions.WithPartitionID` option of start topic writer.

## v3.42.10
* Added exit from retryer if got grpc-error `Unauthenticated` on `discovery/ListEndpoints` call

## v3.42.9
* Added `internal/xerrors.Errorf` error for wrap multiple errors and check them with `errors.Is` of `errors.As`
* Fixed corner cases of `internal/wait.Wait`
* Added check of port in connection string and error throw
* Fixed bug with initialization of connection pool before apply static credentials
* Refactored of applying grpc dial options with defaults
* Added `trace.Driver.{OnBalancerDialEntrypoint,OnBalancerClusterDiscoveryAttempt}` trace events
* Fixed compilation of package `internal/xresolver` with `google.golang.org/grpc@v1.53`
* Fixed returning `io.EOF` on `rows.Next` and `rows.NextResultSet`
* Added wrapping of errors from unary and stream results
* Added error throw on `database/sql.Conn.BeginTx()`, `*sql.Tx.ExecContext` and `*sql.Tx.QueryContext` if query mode is not `ydb.DataQueryMode`
* Added test for `database/sql` scan-query

## v3.42.8
* Fixed `internal/scheme/helpers/IsDirectoryExists(..)` recursive bug

## v3.42.7
* Fixed `sugar.IsTableExists` with recursive check directory exists
* Added `sugar.IsDirectoryExists`
* Changed type of `table/options.IndexType` for type checks
* Added constants `table/options.IndexTypeGlobal` and `table/options.IndexTypeGlobalAsync`
* Added `table/options.IndexDescription.Type` field with `table/options.IndexType` type

## v3.42.6
* Implemented `driver.RowsColumnTypeDatabaseTypeName` interface in `internal/xsql.rows` struct
* Extended `internal/xsql.conn` struct with methods for getting `YDB` metadata
* Added `scheme.Client` to `internal/xsql.connection` interface
* Added `helpers` package with method for checking existence of table, refactored `sugar.IsTableExists()`
* Added checks for nil option to all opts range loops
* Moved content of package `internal/ctxlabels` into `internal/xcontext`
* Implemented `GRPCStatus` method in `internal/xerrors/transportError`
* Added different implementations of stacktrace error for grpc errors and other
* Dropped `internal/xnet` package as useless
* Fixed default grpc dial options
* Replaced single connection for discovery repeater into connection which creates each time for discovery request
* Fixed retry of cluster discovery on initialization
* Fixed dial timeout processing

## v3.42.5
* Fixed closing of `database/sql` connection (aka `YDB` session)
* Made `session.Close()` as `nop` for idled session
* Implemented goroutine for closing idle connection in `database/sql` driver
* Separated errors of commit from other reader and to expired session
* Fixed wrapping error in `internal/balancer/Balancer.wrapCall()`

## v3.42.4
* Added `ydb.WithDisableServerBalancer()` database/sql connector option

## v3.42.3
* Added `credentials.NewStaticCredentials()` static credentials constructor
* Changed `internal/credentials.NewStaticCredentials()` signature and behaviour for create grpc connection on each call to auth service
* Downgrade `google.golang.org/grpc` to `v1.49.0`

## v3.42.2
* Added `trace.Details.Details()` method for use external detailer

## v3.42.1
* Fixed lazy transaction example for `godoc`

## v3.42.0
* Added retry policy options for topics: `topic/topicoptions.WithReaderCheckRetryErrorFunction`, `topic/topicoptions.WithReaderStartTimeout`, `topic/topicoptions.WithWriterCheckRetryErrorFunction`, `topic/topicoptions.WithWriterStartTimeout`
* Refactored `internal/conn` middlewares
* Added `trace.tableSessionInfo.LastUsage()` method for get last usage timestamp
* Reverted `tx.WithCommit()` changes for fix unstable behaviour of lazy transactions
* Added `options.WithCommit()` option for execute query with auto-commit flag
* Removed `trace.TableTransactionExecuteStartInfo.KeepInCache` field as redundant

## v3.41.0
* Added option for set interval of auth token update in topic streams
* Supported internal allocator in `{session,statement}.Execute` for decrease memory usage
* Fixed typo in `topic/README.md`
* Upgraded `ydb-go-genproto` dependency
* Fixed duplicating of traces in `table.Client.Do()` call
* Supported `table.Transaction.WithCommit()` method for execute query and auto-commit after
* Added `DataColumns` to `table.options.IndexDescription`
* Added `scheme.EntryColumnStore` and `scheme.EntryColumnColumn` entry types
* Added `table.options.WithPartitioningBy(columns)` option

## v3.40.1
* Added constructor of `options.TimeToLiveSettings` and fluent modifiers

## v3.40.0
* Added `options.WithAddAttribute` and `options.WithDropAttribute` options for `session.AlterTable` request
* Added `options.WithAddIndex` and `options.WithDropIndex` options for `session.AlterTable` request
* Added return error while create topic writer with not equal producer id and message group id.
* Added package `meta` with methods about `YDB` metadata
* Added `meta.WithTrailerCallback(ctx, callback)` context modifier for attaching callback function which will be called on incoming metadata
* Added `meta.ConsumedUnits(metadata.MD)` method for getting consumed units from metadata
* Added `NestedCall` field to retry trace start infos for alarm on nested calls
* Added `topicoptions.WithWriterTrace` option for attach tracer into separated writer
* Added `sugar.IsTableExists()` helper for check existence of table

## v3.39.0
* Removed message level partitioning from experimental topic API. It is unavailable on server side yet.
* Supported `NullValue` type as received type from `YDB`
* Supported `types.SetValue` type
* Added `types.CastTo(types.Value, destination)` public method for cast `types.Value` to golang native type value destination
* Added `types.TupleItem(types.Value)`, `types.StructFields(types.Value)` and `types.DictValues(types.Value)` funcs (extractors of internal fields of tuple, struct and dict values)
* Added `types.Value.Yql()` func for getting values string representation as `YQL` literal
* Added `types.Type.Yql()` func for getting `YQL` representation of type
* Marked `table/types.WriteTypeStringTo` as deprecated
* Added `table/options.WithDataColumns` for supporting covering indexes
* Supported `balancer` query string parameter in `DSN`
* Fixed bug with scanning `YSON` value from result set
* Added certificate caching in `WithCertificatesFromFile` and `WithCertificatesFromPem`

## v3.38.5
* Fixed bug from scan unexpected column name

## v3.38.4
* Changed type of `table/options.{Create,Alter,Drop}TableOption` from func to interface
* Added implementations of `table/options.{Create,Alter,Drop}Option`
* Changed type of `topic/topicoptions.{Create,Alter,Drop}Option` from func to interface
* Added implementations of `topic/topicoptions.{Create,Alter}Option`
* Fix internal race-condition bugs in internal background worker

## v3.38.3
* Added retries to initial discovering

## v3.38.2
* Added missing `RetentionPeriod` parameter for topic description
* Fixed reconnect problem for topic client
* Added queue limit for sent messages and split large grpc messages while send to topic service
* Improved control plane for topic services: allow list topic in schema, read cdc feeds in table, retry on contol plane operations in topic client, full info in topic describe result
* Allowed writing zero messages to topic writer

## v3.38.1
* Fixed deadlock with implicit usage of `internal.table.Client.internalPoolAsyncCloseSession`

## v3.38.0
* Fixed commit errors for experimental topic reader
* Updated `ydb-go-genproto` dependency
* Added `table.WithSnapshotReadOnly()` `TxOption` for supporting `SnapshotReadOnly` transaction control
* Fixed bug in `db.Scripting()` queries (not checked operation results)
* Added `sugar.ToYdbParam(sql.NamedArg)` helper for converting `sql.NamedArg` to `table.ParameterOption`
* Changed type `table.ParameterOption` for getting name and value from `table.ParameterOption` instance
* Added topic writer experimental api with internal logger

## v3.37.8
* Refactored the internal closing behaviour of table client
* Implemented the `sql.driver.Validator` interface
* Fixed update token for topic reader
* Marked sessions which creates from `database/sql` driver as supported server-side session balancing

## v3.37.7
* Changed type of truncated result error from `StreamExecuteScanQuery` to retryable error
* Added closing sessions if node removed from discovery results
* Moved session status type from `table/options` package to `table`
* Changed session status source type from `uint32` to `string` alias

## v3.37.6
* Added to balancer notifying mechanism for listening in table client event about removing some nodes and closing sessions on them
* Removed from public client interfaces `closer.Closer` (for exclude undefined behaviour on client-side)

## v3.37.5
* Refactoring of `xsql` errors checking

## v3.37.4
* Revert the marking of context errors as required to delete session

## v3.37.3
* Fixed alter topic request - stop send empty setSupportedCodecs if customer not set them
* Marked the context errors as required to delete session
* Added log topic api reader for internal logger

## v3.37.2
* Fixed nil pointer exception in topic reader if reconnect failed

## v3.37.1
* Refactored the `xsql.badconn.Error`

## v3.37.0
* Supported read-only `sql.LevelSnapshot` isolation with fake transaction and `OnlineReadOnly` transaction control (transient, while YDB clusters are not updated with true snapshot isolation mode)
* Supported the `*sql.Conn` as input type `ydb.Unwrap` helper for go's 1.18

## v3.36.2
* Changed output of `sugar.GenerateDeclareSection` (added error as second result)
* Specified `sugar.GenerateDeclareSection` for `go1.18` (supports input types `*table.QueryParameters` `[]table.ParameterOption` or `[]sql.NamedArg`)
* Supports different go's primitive value types as arg of `sql.Named("name", value)`
* Added `database/sql` example and docs

## v3.36.1
* Fixed `xsql.Rows` error checking

## v3.36.0
* Changed behavior on `result.Err()` on truncated result (returns non-retryable error now, exclude `StreamExecuteScanQuery`)
* Added `ydb.WithIgnoreTruncated` option for disabling errors on truncated flag
* Added simple transaction control constructors `table.OnlineReadOnlyTxControl()` and `table.StaleReadOnlyTxControl()`
* Added transaction control specifier with context `ydb.WithTxControl`
* Added value constructors `types.BytesValue`, `types.BytesValueFromString`, `types.TextValue`
* Removed auto-prepending declare section on `xsql` queries
* Supports `time.Time` as type destination in `xsql` queries
* Defined default dial timeout (5 seconds)

## v3.35.1
* Removed the deprecation warning for `ydb.WithSessionPoolIdleThreshold` option

## v3.35.0
* Replaced internal table client background worker to plain wait group for control spawned goroutines
* Replaced internal table client background session keeper to internal background session garbage collector for idle sessions
* Extended the `DescribeTopicResult` struct

## v3.34.2
* Added some description to error message from table pool get
* Moved implementation `sugar.GenerateDeclareSection` to `internal/table`
* Added transaction trace callbacks and internal logging with them
* Stored context from `BeginTx` to `internal/xsql` transaction
* Added automatically generated declare section to query text in `database/sql` usage
* Removed supports `sql.LevelSerializable`
* Added `retry.Do` helper for retry custom lambda with `database/sql` without transactions
* Removed `retry.WithTxOptions` option (only default isolation supports)

## v3.34.1
* Changed `database/sql` driver `prepare` behaviour to `nop` with proxing call to conn exec/query with keep-in-cache flag
* Added metadata to `trace.Driver.OnInvoke` and `trace.Driver.OnNewStream` done events

## v3.34.0
* Improved the `xsql` errors mapping to `driver.ErrBadConn`
* Extended `retry.DoTx` test for to achieve equivalence with `retry.Retry` behaviour
* Added `database/sql` events for tracing `database/sql` driver events
* Added internal logging for `database/sql` events
* Supports `YDB_LOG_DETAILS` environment variable for specify scope of log messages
* Removed support of `YDB_LOG_NO_COLOR` environment variable
* Changed default behaviour of internal logger to without coloring
* Fixed coloring (to true) with environment variable `YDB_LOG_SEVERITY_LEVEL`
* Added `ydb.WithStaticCredentials(user, password)` option for make static credentials
* Supports static credentials as part of connection string (dsn - data source name)
* Changed minimal supported version of go from 1.14 to 1.16 (required for jwt library)


## v3.33.0
* Added `retry.DoTx` helper for retrying `database/sql` transactions
* Implemented `database/sql` driver over `ydb-go-sdk`
* Marked as deprecated `trace.Table.OnPoolSessionNew` and `trace.Table.OnPoolSessionClose` events
* Added `trace.Table.OnPoolSessionAdd` and `trace.Table.OnPoolSessionRemove` events
* Refactored session lifecycle in session pool for fix flaked `TestTable`
* Fixed deadlock in topicreader batcher, while add and read raw server messages
* Fixed bug in `db.Topic()` with send response to stop partition message

## v3.32.1
* Fixed flaky TestTable
* Renamed topic events in `trace.Details` enum

## v3.32.0
* Refactored `trace.Topic` (experimental) handlers
* Fixed signature and names of helpers in `topic/topicsugar` package
* Allowed parallel reading and committing topic messages

## v3.31.0
* Extended the `ydb.Connection` interface with experimental `db.Topic()` client (control plane and reader API)
* Removed `ydb.RegisterParser()` function (was needed for `database/sql` driver outside `ydb-go-sdk` repository, necessity of `ydb.RegisterParser()` disappeared with implementation `database/sql` driver in same repository)
* Refactored `db.Table().CreateSession(ctx)` (maked retryable with internal create session timeout)
* Refactored `internal/table/client.createSession(ctx)` (got rid of unnecessary goroutine)
* Supported many user-agent records

## v3.30.0
* Added `ydb.RegisterParser(name string, parser func(value string) []ydb.Option)` function for register parser of specified param name (supporting additional params in connection string)
* Fixed writing `KeepInCacheFlag` in table traces

## v3.29.5
* Fixed regression of `table/types.WriteTypeStringTo`

## v3.29.4
* Added touching of last updated timestamp in existing conns on stage of applying new endpoint list

## v3.29.3
* Reverted `xerrors.IsTransportError(err)` behaviour for raw grpc errors to false

## v3.29.2
* Enabled server-side session balancing for sessions created from internal session pool
* Removed unused public `meta.Meta` methods
* Renamed `meta.Meta.Meta(ctx)` public method to `meta.Meta.Context(ctx)`
* Reverted default balancer to `balancers.RandomChoice()`

## v3.29.1
* Changed default balancer to `balancers.PreferLocalDC(balancers.RandomChoice())`

## v3.29.0
* Refactored `internal/value` package for decrease CPU and memory workload with GC
* Added `table/types.Equal(lhs, rhs)` helper for check equal for two types

## v3.28.3
* Fixed false-positive node pessimization on receiving from stream io.EOF

## v3.28.2
* Upgraded dependencies (grpc, protobuf, testify)

## v3.28.1
* Marked dial errors as retryable
* Supported node pessimization on dialing errors
* Marked error from `Invoke` and `NewStream` as retryable if request not sended to server

## v3.28.0
* Added `sugar.GenerateDeclareSection()` helper for make declare section in `YQL`
* Added check when parameter name not started from `$` and automatically prepends it to name
* Refactored connection closing

## v3.27.0
* Added internal experimental packages `internal/value/exp` and `internal/value/exp/allocator` with alternative value implementations with zero-allocation model
* Supported parsing of database name from connection string URI path
* Added `options.WithExecuteScanQueryStats` option
* Added to query stats plan and AST
* Changed behaviour of `result.Stats()` (if query result have no stats - returns `nil`)
* Added context cancel with specific error
* Added mutex wrapper for mutex, rwmutex for guarantee unlock and better show critical section

## v3.26.10
* Fixed syntax mistake in `trace.TablePooStateChangeInfo` to `trace.TablePoolStateChangeInfo`

## v3.26.9
* Fixed bug with convert ydb value to `time.Duration` in `result.Scan[WithDefaults,Named]()`
* Fixed bug with make ydb value from `time.Duration` in `types.IntervalValueFromDuration(d)`
* Marked `table/types.{IntervalValue,NullableIntervalValue}` as deprecated

## v3.26.8
* Removed the processing of trailer metadata on stream calls

## v3.26.7
* Updated the `ydb-go-genproto` dependency

## v3.26.6
* Defined the `SerializableReadWrite` isolation level by default in `db.Table.DoTx(ctx, func(ctx, tx))`
* Updated the `ydb-go-genproto` dependency

## v3.26.5
* Disabled the `KeepInCache` policy for queries without params

## v3.26.4
* Updated the indirect dependency to `gopkg.in/yaml.v3`

## v3.26.3
* Removed `Deprecated` mark from `table/session.Prepare` method
* Added comments for `table/session.Execute` method

## v3.26.2
* Refactored of making permissions from scheme entry

## v3.26.1
* Removed deprecated traces

## v3.26.0
* Fixed data race on session stream queries
* Renamed `internal/router` package to `internal/balancer` for unambiguous understanding of package mission
* Implemented detection of local data-center with measuring tcp dial RTT
* Added `trace.Driver.OnBalancer{Init,Close,ChooseEndpoint,Update}` events
* Marked the driver cluster events as deprecated
* Simplified the balancing logic

## v3.25.3
* Changed primary license to `Apache2.0` for auto-detect license
* Refactored `types.Struct` value creation

## v3.25.2
* Fixed repeater initial force timeout from 500 to 0.5 second

## v3.25.1
* Fixed bug with unexpected failing of call `Invoke` and `NewStream` on closed cluster
* Fixed bug with releasing `internal/conn/conn.Pool` in cluster
* Replaced interface `internal/conn/conn.Pool` to struct `internal/conn/conn.Pool`

## v3.25.0
* Added `ydb.GRPCConn(ydb.Connection)` helper for connect to driver-unsupported YDB services
* Marked as deprecated `session.Prepare` callback
* Marked as deprecated `options.WithQueryCachePolicyKeepInCache` and `options.WithQueryCachePolicy` options
* Added `options.WithKeepInCache` option
* Enabled by default keep-in-cache policy for data queries
* Removed from `ydb.Connection` embedding of `grpc.ClientConnInterface`
* Fixed stopping of repeater
* Added log backoff between force repeater wake up's (from 500ms to 32s)
* Renamed `trace.DriverRepeaterTick{Start,Done}Info` to `trace.DriverRepeaterWakeUp{Start,Done}Info`
* Fixed unexpected `NullFlag` while parse nil `JSONDocument` value
* Removed `internal/conn/conn.streamUsages` and `internal/conn/conn.usages` (`internal/conn.conn` always touching last usage timestamp on API calls)
* Removed auto-reconnecting for broken conns
* Renamed `internal/database` package to `internal/router` for unambiguous understanding of package mission
* Refactored applying actual endpoints list after re-discovery (replaced diff-merge logic to swap cluster struct, cluster and balancers are immutable now)
* Added `trace.Driver.OnUnpessimizeNode` trace event

## v3.24.2
* Changed default balancer to `RandomChoice()` because `PreferLocalDC()` balancer works incorrectly with DNS-balanced call `Discovery/ListEndpoints`

## v3.24.1
* Refactored initialization of coordination, ratelimiter, scheme, scripting and table clients from `internal/lazy` package to each client initialization with `sync.Once`
* Removed `internal/lazy` package
* Added retry option `retry.WithStackTrace` for wrapping errors with stacktrace

## v3.24.0
* Fixed re-opening case after close lazy-initialized clients
* Removed dependency of call context for initializing lazy table client
* Added `config.AutoRetry()` flag with `true` value by default. `config.AutoRetry()` affects how to errors handle in sub-clients calls.
* Added `config.WithNoAutoRetry` for disabling auto-retry on errors in sub-clients calls
* Refactored `internal/lazy` package (supported check `config.AutoRetry()`, removed all error wrappings with stacktrace)

## v3.23.0
* Added `WithTLSConfig` option for redefine TLS config
* Added `sugar.LoadCertificatesFromFile` and `sugar.LoadCertificatesFromPem` helpers

## v3.22.0
* Supported `json.Unmarshaler` type for scanning row to values
* Reimplemented `sugar.DSN` with `net/url`

## v3.21.0
* Fixed gtrace tool generation code style bug with leading spaces
* Removed accounting load factor (unused field) in balancers
* Enabled by default anonymous credentials
* Enabled by default internal dns resolver
* Removed from defaults `grpc.WithBlock()` option
* Added `ydb.Open` method with required param connection string
* Marked `ydb.New` method as deprecated
* Removed package `dsn`
* Added `sugar.DSN` helper for make dsn (connection string)
* Refactored package `retry` (moved `retryBackoff` and `retryMode` implementations to `internal`)
* Refactored `config.Config` (remove interface `Config`, renamed private struct `config` to `Config`)
* Moved `discovery/config` to `internal/discovery/config`
* Moved `coordination/config` to `internal/coordination/config`
* Moved `scheme/config` to `internal/scheme/config`
* Moved `scripting/config` to `internal/scripting/config`
* Moved `table/config` to `internal/table/config`
* Moved `ratelimiter/config` to `internal/ratelimiter/config`

## v3.20.2
* Fixed race condition on lazy clients first call

## v3.20.1
* Fixed gofumpt linter issue on `credentials/credentials.go`

## v3.20.0
* Added `table.DefaultTxControl()` transaction control creator with serializable read-write isolation mode and auto-commit
* Fixed passing nil query parameters
* Fixed locking of cluster during call `cluster.Get`

## v3.19.1
* Simplified README.md for godoc documentation in pkg.go.dev

## v3.19.0
* Added public package `dsn` for making piped data source name (connection string)
* Marked `ydb.WithEndpoint`, `ydb.WithDatabase`, `ydb.WithSecure`, `ydb.WithInsecure` options as deprecated
* Moved `ydb.RegisterParser` to package `dsn`
* Added version into all error and warn log messages

## v3.18.5
* Fixed duplicating `WithPanicCallback` proxying to table config options
* Fixed comments for `xerrros.Is` and `xerrros.As`

## v3.18.4
* Renamed internal packages `errors`, `net` and `resolver` to `xerrors`, `xnet` and `xresolver` for excluding ambiguous interpretation
* Renamed internal error wrapper `xerrors.New` to `xerrors.Wrap`

## v3.18.3
* Added `WithPanicCallback` option to all service configs (discovery, coordination, ratelimiter, scheme, scripting, table) and auto-applying from `ydb.WithPanicCallback`
* Added panic recovering (if defined `ydb.WithPanicCallback` option) which thrown from retry operation

## v3.18.2
* Refactored balancers (makes concurrent-safe)
* Excluded separate balancers lock from cluster
* Refactored `cluster.Cluster` interface (`Insert` and `Remove` returning nothing now)
* Replaced unsafe `cluster.close` boolean flag to `cluster.done` chan for listening close event
* Added internal checker `cluster.isClosed()` for check cluster state
* Extracted getting available conn from balancer to internal helper `cluster.get` (called inside `cluster.Get` as last effort)
* Added checking `conn.Conn` availability with `conn.Ping()` in prefer nodeID case

## v3.18.1
* Added `conn.Ping(ctx)` method for check availability of `conn.Conn`
* Refactored `cluster.Cluster.Get(ctx)` to return only available connection (instead of returning any connection from balancer)
* Added address to error description thrown from `conn.take()`
* Renamed package `internal/db` to `internal/database` to exclude collisions with variable name `db`

## v3.18.0
* Added `go1.18` to test matrix
* Added `ydb.WithOperationTimeout` and `ydb.WithOperationCancelAfter` context modifiers

## v3.17.0
* Removed redundant `trace.With{Table,Driver,Retry}` and `trace.Context{Table,Driver,Retry}` funcs
* Moved `gtrace` tool from `./cmd/gtrace` to `./internal/cmd/gtrace`
* Refactored `gtrace` tool for generate `Compose` options
* Added panic recover on trace calls in `Compose` call step
* Added `trace.With{Discovery,Driver,Coordination,Ratelimiter,Table,Scheme,Scripting}PanicCallback` options
* Added `ydb.WithPanicCallback` option

## v3.16.12
* Fixed bug with check acquire error over `ydb.IsRatelimiterAcquireError`
* Added full changelog link to github release description

## v3.16.11
* Added stacktrace to errors with issues

## v3.16.10
* Refactored `cluster.Cluster` and `balancer.Balancer` interfaces (removed `Update` method)
* Replaced `cluster.Update` with `cluster.Remove` and `cluster.Insert` calls
* Removed `trace.Driver.OnClusterUpdate` event
* Fixed bug with unexpected changing of local datacenter flag in endpoint
* Refactored errors wrapping (stackedError are not ydb error now, checking `errors.IsYdb(err)` with `errors.As` now)
* Wrapped retry operation errors with `errors.WithStackTrace(err)`
* Changed `trace.RetryLoopStartInfo.Context` type from `context.Context` to `*context.Context`

## v3.16.9
* Refactored internal operation and transport errors

## v3.16.8
* Added `config.ExcludeGRPCCodesForPessimization()` opttion for exclude some grpc codes from pessimization rules
* Refactored pessimization node conditions
* Added closing of ticker in `conn.Conn.connParker`
* Removed `config.WithSharedPool` and usages it
* Removed `conn.Creator` interface and usage it
* Removed unnecessary options append in `ydb.With`

## v3.16.7
* Added closing `conn.Conn` if discovery client build failure
* Added wrapping errors with stacktrace
* Added discharging banned state of `conn.Conn` on `cluster.Update` step

## v3.16.6
* Rollback moving `meta.Meta` call to conn exclusively from `internal/db` and `internal/discovery`
* Added `WithMeta()` discovery config option

## v3.16.5
* Added `config.SharedPool()` setting and `config.WithSharedPool()` option
* Added management of shared pool flag on change dial timeout and credentials
* Removed explicit checks of conditions for use (or not) shared pool in `ydb.With()`
* Renamed `internal/db` interfaces
* Changed signature of `conn.Conn.Release` (added error as result)

## v3.16.4
* Removed `WithMeta()` discovery config option
* Moved `meta.Meta` call to conn exclusively

## v3.16.3
* Replaced panic on cluster close to error issues

## v3.16.2
* Fixed bug in `types.Nullable()`
* Refactored package `meta`
* Removed explicit call meta in `db.New()`

## v3.16.1
* Added `WithMeta()` discovery config option
* Fixed bug with credentials on discovery

## v3.16.0
* Refactored internal dns-resolver
* Added option `config.WithInternalDNSResolver` for use internal dns-resolver and use resolved IP-address for dialing instead FQDN-address

## v3.15.1
* Removed all conditions for trace retry errors
* Fixed background color of warn messages
* Added to log messages additional information about error, such as retryable (or not), delete session (or not), etc.

## v3.15.0
* Added github action for publish release tags
* Refactored version constant (split to major, minor and patch constants)
* Added `table.types.Nullable{*}Value` helpers and `table.types.Nullable()` common helper
* Fixed race on check trailer on closing table grpc-stream
* Refactored traces (start and done struct names have prefix about trace)
* Replaced `errors.Error`, `errors.Errorf` and `errors.ErrorfSkip` to single `errors.WithStackTrace`
* Refactored table client options
* Declared and implemented interface `errors.isYdbError` for checking ybd/non-ydb errors
* Fixed double tracing table do events
* Added `retry.WithFastBackoff` and `retry.WithFastBackoff` options
* Refactored `table.CreateSession` as retry operation with options
* Moved log level from root of repository to package `log`
* Added details and address to transport error
* Fixed `recursive` param in `ratelimiter.ListResource`
* Added counting stream usages for exclude park connection if it in use
* Added `trace.Driver` events about change stream usage and `conn.Release()` call

## 3.14.4
* Implemented auto-removing `conn.Conn` from `conn.Pool` with counting usages of `conn.Conn`
* Refactored naming of source files which declares service client interfaces

## 3.14.3
* Fixed bug with update balancer element with nil handle

## 3.14.2
* Refactored internal error wrapping (with file and line identification) - replaced `fmt.Printf("%w", err)` error wrapping to internal `stackError`

## 3.14.1
* Added `balacers.CreateFromConfig` balancer creator
* Added `Create` method to interface `balancer.Balancer`

## 3.14.0
* Added `balacers.FromConfig` balancer creator

## 3.13.3
* Fixed linter issues

## 3.13.2
* Fixed race with read/write pool conns on closing conn

## 3.13.1
* Improved error messages
* Defended `cluster.balancer` with `sync.RWMutex` on `cluster.Insert`, `cluster.Update`, `cluster.Remove` and `cluster.Get`
* Excluded `Close` and `Park` methods from `conn.Conn` interface
* Fixed bug with `Multi` balancer `Create()`
* Improved `errors.IsTransportError` (check a few transport error codes instead check single transport error code)
* Improved `errors.Is` (check a few errors instead check single error)
* Refactored YDB errors checking API on client-side
* Implemented of scripting traces

## 3.13.0
* Refactored `Connection` interface
* Removed `CustomOption` and taking client with custom options
* Removed `proxy` package
* Improved `db.With()` helper for child connections creation
* Set shared `conn.Pool` for all children `ydb.Connection`
* Fixed bug with `RoundRobin` and `RandomChoice` balancers `Create()`

## 3.12.1
* Added `trace.Driver.OnConnPark` event
* Added `trace.Driver.OnConnClose` event
* Fixed bug with closing nil session in table retryer
* Restored repeater `Force` call on pessimize event
* Changed mutex type in `conn.Conn` from `sync.Mutex` to `sync.RWMutex` for exclude deadlocks
* Reverted applying empty `discovery` results to `cluster`

## 3.12.0
* Added `balancers.Prefer` and `balancers.PreferWithFallback` constructors

## 3.11.13
* Added `trace.Driver.OnRepeaterWakeUp` event
* Refactored package `repeater`

## 3.11.12
* Added `trace.ClusterInsertDoneInfo.Inserted` boolean flag for notify about success of insert endpoint into balancer
* Added `trace.ClusterRemoveDoneInfo.Removed` boolean flag for notify about success of remove endpoint from balancer

## 3.11.11
* Reverted usage of `math/rand` (instead `crypto/rand`)

## 3.11.10
* Imported tool gtrace to `./cmd/gtrace`
* Changed minimal version of go from 1.13 to 1.14

## 3.11.9
* Fixed composing of service traces
* Fixed end-call of `trace.Driver.OnConnStateChange`

## 3.11.8
* Added `trace.EndpointInfo.LastUpdated()` timestamp
* Refactored `endpoint.Endpoint` (split to struct `endopint` and interface `Endpoint`)
* Returned safe-thread copy of `endpoint.Endpoint` to trace callbacks
* Added `endpoint.Endpoint.Touch()` func for refresh endpoint info
* Added `conn.conn.onClose` slice for call optional funcs on close step
* Added removing `conn.Conn` from `conn.Pool` on `conn.Conn.Close()` call
* Checked cluster close/empty on keeper goroutine
* Fixed `internal.errors.New` wrapping depth
* Added context flag for no wrapping operation results as error
* Refactored `trace.Driver` conn events

## 3.11.7
* Removed internal alias-type `errors.IssuesIterator`

## 3.11.6
* Changed `trace.GetCredentialsDoneInfo` token representation from bool to string
* Added `log.Secret` helper for mask token

## 3.11.5
* Replaced meta in `proxyConnection.Invoke` and `proxyConnection.NewStream`

## 3.11.4
* Refactored `internal/cluster.Cluster` (add option for notify about external lock, lock cluster for update cluster endpoints)
* Reverted `grpc.ClientConnInterface` API to `ydb.Connection`

## 3.11.3
* Replaced in `table/types/compare_test.go` checking error by error message to checking with `errors.Is()`

## 3.11.2
* Wrapped internal errors in retry operations

## 3.11.1
* Excluded error wrapping from retry operations

## 3.11.0
* Added `ydb.WithTLSSInsecureSkipVerify()` option
* Added `trace.Table.OnPoolStateChange` event
* Wrapped internal errors with print <func, file, line>
* Removed `trace.Table.OnPoolTake` event (unused)
* Refactored `trace.Details` matching by string pattern
* Added resolver trace callback
* Refactored initialization step of grpc dial options
* Added internal package `net` with `net.Conn` proxy object
* Fixed closing proxy clients
* Added `ydb.Connection.With(opts ...ydb.CustomOption)` for taking proxy `ydb.Connection` with some redefined options
* Added `ydb.MetaRequestType` and `ydb.MetaTraceID` aliases to internal `meta` package constants
* Added `ydb.WithCustomCredentials()` option
* Refactored `ydb.Ratelimiter().AcquireResource()` method (added options for defining type of acquire request)
* Removed single point to define operation mode params (each grpc-call with `OperationParams` must explicit define `OperationParams`)
* Removed defining operation params over context
* Removed `config.RequestTimeout` and `config.StreamTimeout` (each grpc-call must manage context instead define `config.RequestTimeout` or `config.StreamTimeout`)
* Added internal `OperationTimeout` and `OperationCancelAfter` to each client (ratelimiter, coordination, table, scheme, scripting, discovery) config. `OperationTimeout` and `OperationCancelAfter` config params defined from root config

## 3.10.0
* Extended `trace.Details` constants for support per-service events
* Added `trace.Discovery` struct for traces discovery events
* Added `trace.Ratelimiter`, `trace.Coordination`, `trace.Scripting`, `trace.Scheme` stubs (will be implements in the future)
* Added `ratelimiter/config`, `coordination/config`, `scripting/config`, `scheme/config`, `discovery/config` packages for specify per-service configs
* Removed `trace.Driver.OnDiscovery` callback (moved to `trace.Discovery`)
* Refactored initialization step (firstly makes discovery client)
* Removed `internal/lazy.Discovery` (discovery client always initialized)
* Fixed `trace.Table` event structs
* Refactored grpc options for define dns-balancing configuration
* Refactored `retry.Retry` signature (added `retry.WithID`, `retry.WithTrace` and `retry.WithIdempotent` opt-in args, required param `isIdempotentOperation` removed)
* Refactored package `internal/repeater`

## 3.9.4
* Fixed data race on closing session pool

## 3.9.3
* Fixed busy loop on call internal logger with external logger implementation of `log.Logger`

## 3.9.2
* Fixed `WithDiscoveryInterval()` option with negative argument (must use `SingleConn` balancer)

## 3.9.1
* Added `WithMinTLSVersion` option

## 3.9.0
* Removed `ydb.EndpointDatabase`, `ydb.ConnectionString` and `ydb.MustConnectionString` helpers
* Removed `ydb.ConnectParams` struct and `ydb.WithConnectParams` option creator
* Added internal package `dsn` for register external parsers and parse connection string
* Added `ydb.RegisterParser` method for registering external parser of connection string

## 3.8.12
* Unwrap sub-tests called as `t.Run(...)` in integration tests
* Updated `grpc` dependency (from `v1.38.0` to `v1.43.0`)
* Updated `protobuf` dependency (from `v1.26.0` to `v1.27.1`)
* Added internal retryers into `lazy.Ratelimiter`
* Added internal retryers into `lazy.Coordination`
* Added internal retryers into `lazy.Discovery`
* Added internal retryers into `lazy.Scheme`
* Added internal retryers into `lazy.Scripting`
* Added internal retryer into `lazy.Table.CreateSession`

## 3.8.11
* Fixed version

## 3.8.10
* Fixed misspell linter issue

## 3.8.9
* Removed debug print to log

## 3.8.8
* Refactored session shutdown test

## 3.8.7
* Ignored session shutdown test if no defined `YDB_SHUTDOWN_URLS` environment variable

## 3.8.6
* Added `ydb.WithInsecure()` option

## 3.8.5
* Fixed version

## 3.8.4
* Fixed syntax error in `CHANGELOG.md`

## 3.8.3
* Fixed `CHANGELOG.md`

## 3.8.2
* Updated `github.com/ydb-platform/ydb-go-genproto`

## 3.8.1
* Fixed `trace.Table.OnPoolDoTx` - added `Idempotent` flag to `trace.PoolDoTxStartInfo`

## 3.8.0
* Added `table.result.Result.ScanNamed()` scan function
* Changed connection secure to `true` by default
* Renamed public package `balancer` to `balancers` (this package contains only constructors of balancers)
* Moved interfaces from package `internal/balancer/ibalancer` to `internal/balancer`
* Added `NextResultSetErr()` func for select next result set and return error
* Added package `table/result/indexed` with interfaces `indexed.Required`, `indexed.Optional`, `indexed.RequiredOrOptional`
* Replaced abstract `interface{}` in `Scan` to `indexed.RequiredOrOptional`
* Replaced abstract `interface{}` in `ScanWithDefaults` to `indexed.Required`
* Replaced `trace.Table.OnPoolRetry` callback to `trace.Table.OnPoolDo` and `trace.Table.OnPoolDoTx` callbacks
* Supports server hint `session-close` for gracefully shutdown session

## 3.7.2
* Retry remove directory in `sugar.RemoveRecursive()` for retryable error

## 3.7.1
* Fixed panic on `result.Reset(nil)`

## 3.7.0
* Replaced `Option` to `CustomOption` on `Connection` interface methods
* Implements `WithCustom[Token,Database]` options for redefine database and token
* Removed experimental `balancer.PreferEndpoints[WithFallback][RegEx]` balancers
* Supported connections `TTL` with `Option` `WithConnectionTTL`
* Remove unnecessary `WithFastDial` option (lazy connections are always fast inserts into cluster)
* Added `Scripting` service client with API methods `Execute()`, `StreamExecute()` and `Explain()`
* Added `String()` method to `table.types.Type` interface
* Added `With[Custom]UserAgent()` `Option` and `CustomOption` constructors
* Refactored `log.Logger` interface and internal implementation
* Added `retry.RetryableError()` for returns user-defined error which must be retryed
* Renamed internal type `internal.errors.OperationCompleted` to `internal.errors.OperationStatus`
* Added `String()` method to `table.KeyRange` and `table.Value` types
* Replaced creation of goroutine on each stream call to explicit call stream.Recv() on NextResultSet()

## 3.6.2
* Refactored table retry helpers
* Added new `PreferLocations[WithFallback][RegEx]` balancers
* Added `trace.Details.String()` and `trace.Details.Strings()` helpers
* Added `trace.DetailsFromString(s)` and `trace.DetailsFromStrings(s)` helper

## 3.6.1
* Switched closing cluster after closing all sub-services
* Added windows and macOS runtimes to unit and integration tests

## 3.6.0
* Added `config/balancer` package with popular balancers
* Added new `PreferEndpoints[WithFallback][RegEx]` balancers
* Removed `config.BalancerConfig` struct
* Refactored internal packages (tree to flat, split balancers to different packages)
* Moved a taking conn to start of `conn.Invoke` /` conn.NewStream` for applying timeouts to alive conn instead lazy conn (previous logic applied timeouts to all request including dialing on lazy conn)

## 3.5.4
* Added auto-close stream result on end of stream

## 3.5.3
* Changed `Logger` interface for support custom loggers
* Added public type `LoggerOption` for proxies to internal `logger.Option`
* Fixed deadlock on table stream requests

## 3.5.2
* Fixed data race on closing table result
* Added custom dns-resolver to grpc options for use dns-balancing with round_robin balancing policy
* Wrapped with `recover()` system panic on getting system certificates pool
* Added linters and fixed issues from them
* Changed API of `sugar` package

## 3.5.1
* Added system certificates for `darwin` system
* Fixed `table.StreamResult` finishing
* Fixes `sugar.MakePath()`
* Added helper `ydb.MergeOptions()` for merge several `ydb.Option` to single `ydb.Option`

## 3.5.0
* Added `ClosabelSession` interface which extends `Session` interface and provide `Close` method
* Added `CreateSession` method into `table.Client` interface
* Added `Context` field into `trace.Driver.Net{Dial,Read,Write,Close}StartInfo` structs
* Added `Address` field into `trace.Driver.DiscoveryStartInfo` struct
* Improved logger options (provide err and out writers, provide external logger)
* Renamed package `table.resultset` to `table.result`
* Added `trace.Driver.{OnInit,OnClose}` events
* Changed unit/integration tests running
* Fixed/added YDB error checkers
* Dropped `ydb.WithDriverConfigOptions` (duplicate of `ydb.With`)
* Fixed freeze on closing driver
* Fixed `CGO` race on `Darwin` system when driver tried to expand tilde on certificates path
* Removed `EnsurePathExists` and `CleanupDatabase` from API of `scheme.Client`
* Added helpers `MakePath` and `CleanPath` to root of package `ydb-go-sdk`
* Removed call `types.Scanner.UnmarshalYDB()` inside `scanner.setDefaults()`
* Added `DoTx()` API method into `table.Client`
* Added `String()` method into `ConnectParams` for serialize params to connection string
* Added early exit from Rollback for committed transaction
* Moved `HasNextResultSet()` method from `Result` interface to common `result` interface. It provides access to `HasNextResultSet()` on both result interfaces (unary and stream results)
* Added public credentials constructors `credentials.NewAnonymousCredentials()` and `credentials.NewAccessTokenCredentials(token)`

## 3.4.4
* Prefer `ydb.table.types.Scanner` scanner implementation over `sql.Scanner`, when both available.

## 3.4.3
* Forced `round_robin` grpc load balancing instead default `pick_first`
* Added checker `IsTransportErrorCancelled`

## 3.4.2
* Simplified `Is{Transport,Operation}Error`
* Added `IsYdbError` helper

## 3.4.1
* Fixed retry reaction on operation error NotFound (non-retryable now)

## 3.4.0
* Fixed logic bug in `trace.Table.ExecuteDataQuery{Start,Done}Info`

## 3.3.3
* Cleared repeater context for discovery goroutine
* Fixed type of `trace.Details`

## 3.3.2
* Added `table.options.WithPartitioningSettings`

## 3.3.1
* Added `trace.DriverConnEvents` constant

## 3.3.0
* Stored node ID into `endpoint.Endpoint` struct
* Simplified <Host,Port> in `endpoint.Endpoint` to single fqdn Address
* On table session requests now preferred the endpoint by `ID` extracted from session `ID`. If
  endpoint by `ID` not found - using the endpoint from balancer
* Upgraded internal logger for print colored messages

## 3.2.7
* Fixed compare endpoints func

## 3.2.6
* Reverted `NodeID` as key for link between session and endpoint because yandex-cloud YDB
  installation not supported `Endpoint.ID` entity

## 3.2.5
* Dropped endpoint.Addr entity as unused. After change link type between session and endpoint
  to NodeID endpoint.Addr became unnecessary for internal logic of driver
* Enabled integration test table pool health
* Fixed race on session stream requests

## 3.2.4
* Returned context error when context is done on `session.StreamExecuteScanQuery`
  and `session.StreamReadTable`

## 3.2.3
* Fixed bug of interpret tilda in path of certificates file
* Added chapter to `README.md` about ecosystem of debug tools over `ydb-go-sdk`

## 3.2.2
* Fixed result type of `RawValue.String` (ydb string compatible)
* Fixed scans ydb types into string and slice byte receivers

## 3.2.1
* Upgraded dependencies
* Added `WithEndpoint` and `WithDatabase` Option constructors

## 3.2.0
* added package `log` with interface `log.Logger`
* implements `trace.Driver` and `trace.Table` with `log.Logger`
* added internal leveled logger which implement interface `log.Logger`
* supported environment variable `YDB_LOG_SEVERITY_LEVEL`
* changed name of the field `RetryAttempts` to` Attempts` in the structure `trace.PoolGetDoneInfo`.
  This change reduces back compatibility, but there are no external uses of v3 sdk, so this change is
  fine. We are sorry if this change broke your code

## 3.1.0
* published scheme Client interface

## 3.0.1
* refactored integration tests
* fixed table retry trace calls

## 3.0.0
* Refactored sources for splitting public interfaces and internal
  implementation for core changes in the future without change major version
* Refactored of transport level of driver - now we use grpc code generation by stock `protoc-gen-go` instead internal protoc codegen. New API provide operate from codegen grpc-clients with driver as a single grpc client connection. But driver hide inside self a pool of grpc connections to different cluster endpoints YDB. All communications with YDB (base services includes to driver: table, discovery, coordiantion and ratelimiter) provides stock codegen grpc-clients now.
* Much changed API of driver for easy usage.
* Dropped package `ydbsql` (moved to external project)
* Extracted yandex-cloud authentication to external project
* Extracted examples to external project
* Changed of traces API for next usage in jaeger и prometheus
* Dropped old APIs marked as `deprecated`
* Added integration tests with docker ydb container
* Changed table session and endpoint link type from string address to integer NodeID

## 2.11.0
* Added possibility to override `x-ydb-database` metadata value

## 2.10.9
* Fixed context cancellation inside repeater loop

## 2.10.8
* Fixed data race on cluster get/pessimize

## 2.10.7
* Dropped internal cluster connections tracker
* Switched initial connect to all endpoints after discovery to lazy connect
* Added reconnect for broken conns

## 2.10.6
* Thrown context without deadline into discovery goroutine
* Added `Address` param to `DiscoveryStartInfo` struct
* Forced `round_bobin` grpc load balancing config instead default `pick_first`
* Fixed applying driver trace from context in `connect.New`
* Excluded using session pool usage for create/take sessions in `database/sql`
  driver implementation. Package `ydbsql` with `database/sql` driver implementation
  used direct `CreateSession` table client call in the best effort loop

## 2.10.5
* Fixed panic when ready conns is zero

## 2.10.4
* Initialized repeater permanently regardless of the value `DriverConfig.DiscoveryInterval`
  This change allow forcing re-discovery depends on cluster state

## 2.10.3
* Returned context error when context is done on `StreamExecuteScanQuery`

## 2.10.2
* Fixed `mapBadSessionError()` in `ydbsql` package

## 2.10.1
* Fixed race on `ydbsql` concurrent connect. This hotfix only for v2 version

## 2.10.0
* Added `GlobalAsyncIndex` implementation of index interface

## 2.9.6
* Replaced `<session, endpoint>` link type from raw conn to plain endpoint address
* Moved checking linked endpoint from `driver.{Call,StreamRead}` to `cluster.Get`
* Added pessimization endpoint code for `driver.StreamRead` if transport error received
* Setted transport error `Cancelled` as needs to remove session from pool
* Deprecated connection use policy (used auto policy)
* Fixed goroutines leak on StreamRead call
* Fixed force re-discover on receive error after 1 second
* Added timeout to context in `cluster.Get` if context deadline not defined

## 2.9.5
* Renamed context idempotent operation flag

## 2.9.4
* Forced cancelled transport error as retriable (only idempotent operations)
* Renamed some internal retry mode types

## 2.9.3
* Forced grpc keep-alive PermitWithoutStream parameter to true

## 2.9.2
* Added errors without panic

## 2.9.1
* Added check nil grpc.ClientConn connection
* Processed nil connection error in keeper loop

## 2.9.0
* Added RawValue and supported ydb.Scanner in Scan

## 2.8.0
* Added NextResultSet for both streaming and non-streaming operations

## 2.7.0
* Dropped busy checker logic
* Refactoring of `RetryMode`, `RetryChecker` and `Retryer`
* Added fast/slow retry logic
* Supported context param for retry operation with no idempotent errors
* Added secondary indexes info to table describing method

## 2.6.1
* fix panic on lazy put to full pool

## 2.6.0
* Exported `SessionProvider.CloseSession` func
* Implements by default async closing session and putting busy
  session into pool
* Added some session pool trace funcs for execution control of
  goroutines in tests
* Switched internal session pool boolean field closed from atomic
  usage to mutex-locked usage

## 2.5.7
* Added panic on double scan per row

## 2.5.6
* Supported nil and time conventions for scanner

## 2.5.5
* Reverted adds async sessionGet and opDo into `table.Retry`.
* Added `sessionClose()` func into `SessionProvider` interface.

## 2.5.4
* Remove ready queue from session pool

## 2.5.3
* Fix put session into pool

## 2.5.2
* Fix panic on operate with result scanner

## 2.5.1
* Fix lock on write to chan in case when context is done

## 2.5.0
* Added `ScanRaw` for scan results as struct, list, tuple, map
* Created `RawScanner` interface in order to generate method With

## 2.4.1
* Fixed deadlock in the session pool

## 2.4.0
* Added new scanner API.
* Fixed dualism of interpret data (default values were deprecated for optional values)

## 2.3.3
* Fixed `internal/stats/series.go` (index out of range)
* Optimized rotate buckets in the `Series`

## 2.3.2
* Moved `api/wrap.go` to root for next replacement api package to external genproto

## 2.3.1
* Correct session pool tests
* Fixed conditions with KeepAliveMinSize and `IdleKeepAliveThreshold`

## 2.3.0
* Added credentials connect options:
  - `connect.WithAccessTokenCredentials(accessToken)`
  - `connect.WithAnonymousCredentials()`
  - `connect.WithMetadataCredentials(ctx)`
  - `connect.WithServiceAccountKeyFileCredentiials(serviceAccountKeyFile)`
* Added auth examples:
  - `example/auth/environ`
  - `example/auth/access_token_credentials`
  - `example/auth/anonymous_credentials`
  - `example/auth/metadata_credentials`
  - `example/auth/service_account_credentials`

## 2.2.1
* Fixed returning error from `table.StreamExecuteScanQuery`

## 2.2.0
* Supported loading certs from file using `YDB_SSL_ROOT_CERTIFICATES_FILE` environment variable

## 2.1.0
* Fixed erasing session from pool if session keep-alive count great then `IdleKeepAliveThreshold`
* Add major session pool config params as `connect.WithSessionPool*()` options

## 2.0.3
* Added panic for wrong `NextSet`/`NextStreamSet` call

## 2.0.2
* Fixed infinite keep alive session on transport errors `Cancelled` and `DeadlineExceeded`

## 2.0.1
* Fixed parser of connection string
* Fixed `EnsurePathExists` and `CleanupDatabase` methods
* Fixed `basic_example_v1`
* Renamed example cli flag `-link=connectionString` to `-ydb=connectionString` for connection string to YDB
* Added `-connect-timeout` flag to example cli
* Fixed some linter issues

## 2.0.0
* Renamed package ydbx to connect. New usage semantic: `connect.New()` instead `ydbx.Connect()`
* Added `healthcheck` example
* Fixed all examples with usage connect package
* Dropped `example/internal/ydbutil` package
* Simplified API of Traces - replace all pairs start/done to single handler with closure.

## 1.5.2
* Fixed `WithYdbCA` at nil certPool case

## 1.5.1
* Fixed package name of `ydbx`

## 1.5.0
* Added `ydbx` package

## 1.4.1
* Fixed `fmt.Errorf` error wrapping and some linter issues

## 1.4.0
* Added helper for create credentials from environ
* Added anonymous credentials
* Move YDB Certificate Authority from auth/iam package to root  package. YDB CA need to dial with
  dedicated YDB and not need to dial with IAM. YDB CA automatically added to all grpc calling

## 1.3.0
* Added `Compose` method to traces

## 1.2.0
* Load YDB certificates by default with TLS connection

## 1.1.0
* Support scan-query method in `ydbsql` (database/sql API)

## 1.0.7
* Use `github.com/golang-jwt/jwt` instead of `github.com/dgrijalva/jwt-go`

## 1.0.6
* Append (if not exits) SYNC Operation mode on table calls: *Session, *DataQuery, *Transaction, KeepAlive

## 1.0.5
* Remove unused ContextDeadlineMapping driver config (always used default value)
* Simplify operation params logic
* Append (if not exits) SYNC Operation mode on ExecuteDataQuery call

## 1.0.4
* Fixed timeout and cancellation setting for YDB operations
* Introduced possibility to use `ContextDeadlineNoMapping` once again

## 1.0.3
* Negative `table.Client.MaxQueryCacheSize` will disable a client query cache now
* Refactoring of `meta.go` for simple adding in the future new headers to requests
* Added support `x-ydb-trace-id` as standard SDK header

## 1.0.2
* Implements smart lazy createSession for best control of create/delete session balance. This feature fix leakage of forgotten sessions on server-side
* Some imporvements of session pool stats

## 1.0.1
* Fix closing sessions on PutBusy()
* Force setting operation timeout from client context timeout (if this timeout less then default operation timeout)
* Added helper `ydb.ContextWithoutDeadline` for clearing existing context from any deadlines

## 1.0.0
* SDK versioning switched to `Semantic Versioning 2.0.0`

## 2021.04.1
* Added `table.TimeToLiveSettings` struct and corresponding
  `table.WithTimeToLiveSettings`, `table.WithSetTimeToLive`
  and `table.WithDropTimeToLive` options.
* Deprecated `table.TTLSettings` struct alongside with
  `table.WithTTL`, `table.WithSetTTL` and `table.WithDropTTL` functions.

## 2021.03.2
* Add Truncated flag support.

## 2021.03.1
* Fixed a race between `SessionPool.Put` and `SessionPool.Get`, where the latter
  would end up waiting forever for a session that is already in the pool.

## 2021.02.1
* Changed semantics of `table.Result.O...` methods (e.g., `OUTF8`):
  it will not fail if current item is non-optional primitive.

## 2020.12.1
* added CommitTx method, which returns QueryStats

## 2020.11.4
* re-implementation of ydb.Value comparison
* fix basic examples

## 2020.11.3
* increase default and minimum `Dialer.KeepAlive` setting

## 2020.11.2
* added `ydbsql/connector` options to configure default list of `ExecDataQueryOption`

## 2020.11.1
* tune `grpc.Conn` behaviour

## 2020.10.4
* function to compare two ydb.Value

## 2020.10.3
* support scan query execution

## 2020.10.2
* add table Ttl options

## 2020.10.1
* added `KeyBloomFilter` support for `CreateTable`, `AlterTable` and `DescribeTalbe`
* added `PartitioningSettings` support for `CreateTable`, `AlterTable` and `DescribeTalbe`. Move to `PartitioningSettings` object

## 2020.09.3
* add `FastDial` option to `DriverConfig`.
  This will allow `Dialer` to return `Driver` as soon as the 1st connection is ready.

## 2020.09.2
* parallelize endpoint operations

## 2020.09.1
* added `ProcessCPUTime` method to `QueryStats`
* added `ReadReplicasSettings` support for `CreateTable`, `AlterTable` and `DescribeTalbe`
* added `StorageSettings` support for `CreateTable`, `AlterTable` and `DescribeTalbe`

## 2020.08.2
* added `PartitioningSettings` support for `CreateTable` and `AlterTable`

## 2020.08.1
* added `CPUTime` and `AffectedShards` fields to `QueryPhase` struct
* added `CompilationStats` statistics

## 2020.07.7
* support manage table attributes

## 2020.07.6
* support Column Families

## 2020.07.5
* support new types: DyNumber, JsonDocument

## 2020.07.4
* added coordination service
* added rate_limiter service

## 2020.07.3
* made `api` wrapper for `internal` api subset

## 2020.07.2
* return TableStats and PartitionStats on DescribeTable request with options
* added `ydbsql/connector` option to configure `DefaultTxControl`

## 2020.07.1
* support go modules tooling for ydbgen

## 2020.06.2
* refactored `InstanceServiceAccount`: refresh token in background.
  Also, will never produce error on creation
* added getting `ydb.Credentials` examples

## 2020.06.1

* exported internal `api.Wrap`/`api.Unwrap` methods and linked structures

## 2020.04.5

* return on discovery only endpoints that match SSL status of driver

## 2020.04.4

* added GCP metadata auth style with `InstanceServiceAccount` in `auth.iam`

## 2020.04.3

* fix race in `auth.metadata`
* fix races in test hooks

## 2020.04.2

* set limits to grpc `MaxCallRecvMsgSize` and `MaxCallSendMsgSize` to 64MB
* remove deprecated IAM (jwt) `Client` structure
* fix panic on nil dereference while accessing optional fields of `IssueMessage` message

## 2020.04.1

* added options to `DescribeTable` request
* added `ydbsql/connector` options to configure `pool`s  `KeepAliveBatchSize`, `KeepAliveTimeout`, `CreateSessionTimeout`, `DeleteTimeout`

## 2020.03.2

* set session keepAlive period to 5 min - same as in other SDKs
* fix panic on access to index after pool close

## 2020.03.1

* added session pre-creation limit check in pool
* added discovery trigger on more then half unhealthy transport connects
* await transport connect only if no healthy connections left

## 2020.02

* support cloud IAM (jwt) authorization from service account file
* minimum version of Go become 1.13. Started support of new `errors` features
