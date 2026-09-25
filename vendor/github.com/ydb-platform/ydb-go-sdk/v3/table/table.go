package table

import (
	"context"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Formats"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/closer"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Operation is the interface that holds an operation for retry.
// if Operation returns not nil - operation will retry
// if Operation returns nil - retry loop will break
type Operation func(ctx context.Context, s Session) error

// TxOperation is the interface that holds an operation for retry.
// if TxOperation returns not nil - operation will retry
// if TxOperation returns nil - retry loop will break
type TxOperation func(ctx context.Context, tx TransactionActor) error

type ClosableSession interface {
	closer.Closer

	Session
}

type Client interface {
	// CreateSession returns session or error for manually control of session lifecycle
	//
	// CreateSession implements internal busy loop until one of the following conditions is met:
	// - context was canceled or deadlined
	// - session was created
	//
	// Deprecated: not for public usage. Because explicit session often leaked on server-side due to bad client-side usage.
	// Will be removed after Oct 2024.
	// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
	CreateSession(ctx context.Context, opts ...Option) (s ClosableSession, err error)

	// Do provide the best effort for execute operation.
	//
	// Do implements internal busy loop until one of the following conditions is met:
	// - deadline was canceled or deadlined
	// - retry operation returned nil as error
	//
	// Warning: if context without deadline or cancellation func than Do can run indefinitely.
	Do(ctx context.Context, op Operation, opts ...Option) error

	// DoTx provide the best effort for execute transaction.
	//
	// DoTx implements internal busy loop until one of the following conditions is met:
	// - deadline was canceled or deadlined
	// - retry operation returned nil as error
	//
	// DoTx makes auto begin (with TxSettings, by default - SerializableReadWrite), commit and
	// rollback (on error) of transaction.
	//
	// If op TxOperation returns nil - transaction will be committed
	// If op TxOperation return non nil - transaction will be rollback
	// Warning: if context without deadline or cancellation func than DoTx can run indefinitely
	DoTx(ctx context.Context, op TxOperation, opts ...Option) error

	// BulkUpsert upserts a batch of rows non-transactionally.
	//
	// Returns success only when all rows were successfully upserted. In case of an error some rows might
	// be upserted and some might not.
	BulkUpsert(ctx context.Context, table string, data BulkUpsertData, opts ...Option) error

	DescribeTable(ctx context.Context, path string,
		opts ...options.DescribeTableOption,
	) (desc *options.Description, err error)

	// ReadRows reads a batch of rows non-transactionally.
	ReadRows(
		ctx context.Context, path string, keys types.Value,
		readRowOpts []options.ReadRowsOption, retryOptions ...Option,
	) (_ result.Result, err error)
}

type SessionStatus = string

const (
	SessionStatusUnknown = SessionStatus("unknown")
	SessionReady         = SessionStatus("ready")
	SessionBusy          = SessionStatus("busy")
	SessionClosing       = SessionStatus("closing")
	SessionClosed        = SessionStatus("closed")
)

type SessionInfo interface {
	ID() string
	NodeID() uint32
	Status() SessionStatus
	LastUsage() time.Time
}

type Session interface {
	SessionInfo

	CreateTable(ctx context.Context, path string,
		opts ...options.CreateTableOption,
	) (err error)

	DescribeTable(ctx context.Context, path string,
		opts ...options.DescribeTableOption,
	) (desc options.Description, err error)

	DropTable(ctx context.Context, path string,
		opts ...options.DropTableOption,
	) (err error)

	AlterTable(ctx context.Context, path string,
		opts ...options.AlterTableOption,
	) (err error)

	// Deprecated: use CopyTables method instead
	// Will be removed after Oct 2024.
	// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
	CopyTable(ctx context.Context, dst, src string,
		opts ...options.CopyTableOption,
	) (err error)

	CopyTables(ctx context.Context,
		opts ...options.CopyTablesOption,
	) (err error)

	RenameTables(ctx context.Context,
		opts ...options.RenameTablesOption,
	) (err error)

	Explain(ctx context.Context, sql string) (exp DataQueryExplanation, err error)

	// Prepare prepares query for executing in the future
	Prepare(ctx context.Context, sql string) (stmt Statement, err error)

	// Execute executes query.
	//
	// By default, Execute have a flag options.WithKeepInCache(true) if params is not empty. For redefine behavior -
	// append option options.WithKeepInCache(false)
	Execute(ctx context.Context, tx *TransactionControl, sql string, params *params.Params,
		opts ...options.ExecuteDataQueryOption,
	) (txr Transaction, r result.Result, err error)

	ExecuteSchemeQuery(ctx context.Context, sql string,
		opts ...options.ExecuteSchemeQueryOption,
	) (err error)

	DescribeTableOptions(ctx context.Context) (desc options.TableOptionsDescription, err error)

	StreamReadTable(ctx context.Context, path string,
		opts ...options.ReadTableOption,
	) (r result.StreamResult, err error)

	StreamExecuteScanQuery(ctx context.Context, sql string, params *params.Params,
		opts ...options.ExecuteScanQueryOption,
	) (_ result.StreamResult, err error)

	// Deprecated: use Client instance instead.
	BulkUpsert(ctx context.Context, table string, rows types.Value,
		opts ...options.BulkUpsertOption,
	) (err error)

	// Deprecated: use Client instance instead.
	ReadRows(ctx context.Context, path string, keys types.Value,
		opts ...options.ReadRowsOption,
	) (_ result.Result, err error)

	BeginTransaction(ctx context.Context, tx *TransactionSettings) (x Transaction, err error)

	KeepAlive(ctx context.Context) error
}

type (
	TransactionSettings = tx.Settings
	// Transaction control options
	TxOption = tx.SettingsOption
)

// Explanation is a result of Explain calls.
type Explanation struct {
	Plan string
}

// ScriptingYQLExplanation is a result of Explain calls.
type ScriptingYQLExplanation struct {
	Explanation

	ParameterTypes map[string]types.Type
}

// DataQueryExplanation is a result of ExplainDataQuery call.
type DataQueryExplanation struct {
	Explanation

	AST string
}

// DataQuery only for tracers
type DataQuery interface {
	String() string
	ID() string
	YQL() string
}

type TransactionIdentifier = tx.Identifier

type TransactionActor interface {
	TransactionIdentifier

	Execute(
		ctx context.Context,
		sql string,
		params *params.Params,
		opts ...options.ExecuteDataQueryOption,
	) (result.Result, error)
	ExecuteStatement(
		ctx context.Context,
		stmt Statement,
		params *params.Params,
		opts ...options.ExecuteDataQueryOption,
	) (result.Result, error)
}

type Transaction interface {
	TransactionActor

	CommitTx(
		ctx context.Context,
		opts ...options.CommitTransactionOption,
	) (r result.Result, err error)
	Rollback(
		ctx context.Context,
	) (err error)
}

type Statement interface {
	Execute(
		ctx context.Context,
		tx *TransactionControl,
		params *params.Params,
		opts ...options.ExecuteDataQueryOption,
	) (txr Transaction, r result.Result, err error)
	NumInput() int
	Text() string
}

// TxSettings returns transaction settings
func TxSettings(opts ...TxOption) *TransactionSettings {
	settings := tx.NewSettings(opts...)

	return &settings
}

// BeginTx returns begin transaction control option
func BeginTx(opts ...TxOption) TxControlOption {
	return tx.BeginTx(opts...)
}

func WithTx(t TransactionIdentifier) TxControlOption {
	return tx.WithTx(t)
}

func WithTxID(txID string) TxControlOption {
	return tx.WithTxID(txID)
}

// CommitTx returns commit transaction control option
func CommitTx() TxControlOption {
	return tx.CommitTx()
}

func WithSerializableReadWrite() TxOption {
	return tx.WithSerializableReadWrite()
}

func WithSnapshotReadOnly() TxOption {
	return tx.WithSnapshotReadOnly()
}

func WithStaleReadOnly() TxOption {
	return tx.WithStaleReadOnly()
}

func WithOnlineReadOnly(opts ...TxOnlineReadOnlyOption) TxOption {
	return tx.WithOnlineReadOnly(opts...)
}

type (
	TxOnlineReadOnlyOption = tx.OnlineReadOnlyOption
)

func WithInconsistentReads() TxOnlineReadOnlyOption {
	return tx.WithInconsistentReads()
}

type (
	TxControlOption    = tx.ControlOption
	TransactionControl = tx.Control
)

// TxControl makes transaction control from given options
func TxControl(opts ...TxControlOption) *TransactionControl {
	return tx.NewControl(opts...)
}

// DefaultTxControl returns default transaction control with serializable read-write isolation mode and auto-commit
func DefaultTxControl() *TransactionControl {
	return TxControl(
		BeginTx(WithSerializableReadWrite()),
		CommitTx(),
	)
}

// SerializableReadWriteTxControl returns transaction control with serializable read-write isolation mode
func SerializableReadWriteTxControl(opts ...TxControlOption) *TransactionControl {
	return TxControl(
		append([]TxControlOption{
			BeginTx(WithSerializableReadWrite()),
		}, opts...)...,
	)
}

// OnlineReadOnlyTxControl returns online read-only transaction control
func OnlineReadOnlyTxControl(opts ...TxOnlineReadOnlyOption) *TransactionControl {
	return TxControl(
		BeginTx(WithOnlineReadOnly(opts...)),
		CommitTx(), // open transactions not supported for OnlineReadOnly
	)
}

// StaleReadOnlyTxControl returns stale read-only transaction control
func StaleReadOnlyTxControl() *TransactionControl {
	return TxControl(
		BeginTx(WithStaleReadOnly()),
		CommitTx(), // open transactions not supported for StaleReadOnly
	)
}

// SnapshotReadOnlyTxControl returns snapshot read-only transaction control
func SnapshotReadOnlyTxControl() *TransactionControl {
	return TxControl(
		BeginTx(WithSnapshotReadOnly()),
		CommitTx(), // open transactions not supported for StaleReadOnly
	)
}

// QueryParameters
type (
	ParameterOption = params.NamedValue
	QueryParameters = params.Params
)

func NewQueryParameters(opts ...ParameterOption) *QueryParameters {
	qp := QueryParameters(make([]*params.Parameter, len(opts)))
	for i, opt := range opts {
		if opt != nil {
			qp[i] = params.Named(opt.Name(), opt.Value())
		}
	}

	return &qp
}

func ValueParam(name string, v types.Value) ParameterOption {
	switch len(name) {
	case 0:
		panic("empty name")
	default:
		if name[0] != '$' {
			name = "$" + name
		}
	}

	return params.Named(name, v)
}

type Options struct {
	Label           string
	Idempotent      bool
	TxSettings      *TransactionSettings
	TxCommitOptions []options.CommitTransactionOption
	RetryOptions    []retry.Option
	Trace           *trace.Table
}

type Option interface {
	ApplyTableOption(opts *Options)
}

var _ Option = labelOption("")

type labelOption string

func (label labelOption) ApplyTableOption(opts *Options) {
	opts.Label = string(label)
	opts.RetryOptions = append(opts.RetryOptions, retry.WithLabel(string(label)))
}

func WithLabel(label string) labelOption {
	return labelOption(label)
}

var _ Option = retryOptionsOption{}

type retryOptionsOption []retry.Option

func (retryOptions retryOptionsOption) ApplyTableOption(opts *Options) {
	opts.RetryOptions = append(opts.RetryOptions, retryOptions...)
}

// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithRetryBudget(b budget.Budget) retryOptionsOption {
	return []retry.Option{retry.WithBudget(b)}
}

// Deprecated: redundant option
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithRetryOptions(retryOptions []retry.Option) retryOptionsOption {
	return retryOptions
}

func WithIdempotent() retryOptionsOption {
	return []retry.Option{retry.WithIdempotent(true)}
}

var _ Option = txSettingsOption{}

type txSettingsOption struct {
	txSettings *TransactionSettings
}

func (opt txSettingsOption) ApplyTableOption(opts *Options) {
	opts.TxSettings = opt.txSettings
}

func WithTxSettings(txSettings *TransactionSettings) txSettingsOption {
	return txSettingsOption{txSettings: txSettings}
}

var _ Option = txCommitOptionsOption(nil)

type txCommitOptionsOption []options.CommitTransactionOption

func (txCommitOpts txCommitOptionsOption) ApplyTableOption(opts *Options) {
	opts.TxCommitOptions = append(opts.TxCommitOptions, txCommitOpts...)
}

func WithTxCommitOptions(opts ...options.CommitTransactionOption) txCommitOptionsOption {
	return opts
}

var _ Option = traceOption{}

type traceOption struct {
	t *trace.Table
}

func (opt traceOption) ApplyTableOption(opts *Options) {
	opts.Trace = opts.Trace.Compose(opt.t)
}

func WithTrace(t trace.Table) traceOption { //nolint:gocritic
	return traceOption{t: &t}
}

type BulkUpsertData interface {
	ToYDB(tableName string) (*Ydb_Table.BulkUpsertRequest, error)
}
type bulkUpsertRows struct {
	rows types.Value
}

func (data bulkUpsertRows) ToYDB(tableName string) (*Ydb_Table.BulkUpsertRequest, error) {
	return &Ydb_Table.BulkUpsertRequest{
		Table: tableName,
		Rows:  value.ToYDB(data.rows),
	}, nil
}

func BulkUpsertDataRows(rows types.Value) bulkUpsertRows {
	return bulkUpsertRows{
		rows: rows,
	}
}

type bulkUpsertCsv struct {
	data []byte
	opts []csvFormatOption
}

type csvFormatOption interface {
	applyCsvFormatOption(dataFormat *Ydb_Table.BulkUpsertRequest_CsvSettings) (err error)
}

func (data bulkUpsertCsv) ToYDB(tableName string) (*Ydb_Table.BulkUpsertRequest, error) {
	var (
		request = &Ydb_Table.BulkUpsertRequest{
			Table: tableName,
			Data:  data.data,
		}
		dataFormat = &Ydb_Table.BulkUpsertRequest_CsvSettings{
			CsvSettings: &Ydb_Formats.CsvSettings{},
		}
	)

	for _, opt := range data.opts {
		if opt != nil {
			if err := opt.applyCsvFormatOption(dataFormat); err != nil {
				return nil, err
			}
		}
	}

	request.DataFormat = dataFormat

	return request, nil
}

func BulkUpsertDataCsv(data []byte, opts ...csvFormatOption) bulkUpsertCsv {
	return bulkUpsertCsv{
		data: data,
		opts: opts,
	}
}

type csvHeaderOption struct{}

func (opt *csvHeaderOption) applyCsvFormatOption(dataFormat *Ydb_Table.BulkUpsertRequest_CsvSettings) error {
	dataFormat.CsvSettings.Header = true

	return nil
}

// First not skipped line is a CSV header (list of column names).
func WithCsvHeader() csvFormatOption {
	return &csvHeaderOption{}
}

type csvNullValueOption []byte

func (nullValue csvNullValueOption) applyCsvFormatOption(dataFormat *Ydb_Table.BulkUpsertRequest_CsvSettings) error {
	dataFormat.CsvSettings.NullValue = nullValue

	return nil
}

// String value that would be interpreted as NULL.
func WithCsvNullValue(value []byte) csvFormatOption {
	return csvNullValueOption(value)
}

type csvDelimiterOption []byte

func (delimeter csvDelimiterOption) applyCsvFormatOption(dataFormat *Ydb_Table.BulkUpsertRequest_CsvSettings) error {
	dataFormat.CsvSettings.Delimiter = delimeter

	return nil
}

// Fields delimiter in CSV file. It's "," if not set.
func WithCsvDelimiter(value []byte) csvFormatOption {
	return csvDelimiterOption(value)
}

type csvSkipRowsOption uint32

func (skipRows csvSkipRowsOption) applyCsvFormatOption(dataFormat *Ydb_Table.BulkUpsertRequest_CsvSettings) error {
	dataFormat.CsvSettings.SkipRows = uint32(skipRows)

	return nil
}

// Number of rows to skip before CSV data. It should be present only in the first upsert of CSV file.
func WithCsvSkipRows(skipRows uint32) csvFormatOption {
	return csvSkipRowsOption(skipRows)
}

type bulkUpsertArrow struct {
	data []byte
	opts []arrowFormatOption
}

type arrowFormatOption interface {
	applyArrowFormatOption(req *Ydb_Table.BulkUpsertRequest_ArrowBatchSettings) (err error)
}

func (data bulkUpsertArrow) ToYDB(tableName string) (*Ydb_Table.BulkUpsertRequest, error) {
	var (
		request = &Ydb_Table.BulkUpsertRequest{
			Table: tableName,
			Data:  data.data,
		}
		dataFormat = &Ydb_Table.BulkUpsertRequest_ArrowBatchSettings{
			ArrowBatchSettings: &Ydb_Formats.ArrowBatchSettings{},
		}
	)

	for _, opt := range data.opts {
		if opt != nil {
			if err := opt.applyArrowFormatOption(dataFormat); err != nil {
				return nil, err
			}
		}
	}

	request.DataFormat = dataFormat

	return request, nil
}

func BulkUpsertDataArrow(data []byte, opts ...arrowFormatOption) bulkUpsertArrow {
	return bulkUpsertArrow{
		data: data,
		opts: opts,
	}
}

type arrowSchemaOption []byte

func (schema arrowSchemaOption) applyArrowFormatOption(
	dataFormat *Ydb_Table.BulkUpsertRequest_ArrowBatchSettings,
) error {
	dataFormat.ArrowBatchSettings.Schema = schema

	return nil
}

func WithArrowSchema(schema []byte) arrowFormatOption {
	return arrowSchemaOption(schema)
}
