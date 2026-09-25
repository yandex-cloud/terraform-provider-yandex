package table

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Table_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_TableStats"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/conn"
	balancerContext "github.com/ydb-platform/ydb-go-sdk/v3/internal/endpoint"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/feature"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/meta"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/table/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/table/scanner"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xcontext"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsync"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type (
	dataQueryExecutor interface {
		execute(
			ctx context.Context,
			txControl *tx.Control,
			request *Ydb_Table.ExecuteDataQueryRequest,
			callOptions ...grpc.CallOption,
		) (*transaction, result.Result, error)
	}
	tableClientExecutor struct {
		client          Ydb_Table_V1.TableServiceClient
		ignoreTruncated bool
	}
	queryClientExecutor struct {
		client Ydb_Query_V1.QueryServiceClient
		core   query.Core
	}
)

func statsModeToStatsMode(src Ydb_Table.QueryStatsCollection_Mode) (dst Ydb_Query.StatsMode) {
	switch src {
	case Ydb_Table.QueryStatsCollection_STATS_COLLECTION_NONE:
		return Ydb_Query.StatsMode_STATS_MODE_NONE
	case Ydb_Table.QueryStatsCollection_STATS_COLLECTION_BASIC:
		return Ydb_Query.StatsMode_STATS_MODE_BASIC
	case Ydb_Table.QueryStatsCollection_STATS_COLLECTION_FULL:
		return Ydb_Query.StatsMode_STATS_MODE_FULL
	case Ydb_Table.QueryStatsCollection_STATS_COLLECTION_PROFILE:
		return Ydb_Query.StatsMode_STATS_MODE_PROFILE
	default:
		return Ydb_Query.StatsMode_STATS_MODE_UNSPECIFIED
	}
}

func queryExecuteStreamResultToTableResult(
	ctx context.Context,
	stream Ydb_Query_V1.QueryService_ExecuteQueryClient,
) (_ *transaction, _ result.Result, finalErr error) {
	var (
		t          *transaction
		resultSets []*Ydb.ResultSet
		queryStats *Ydb_TableStats.QueryStats
	)

	for {
		if err := ctx.Err(); err != nil {
			return nil, nil, xerrors.WithStackTrace(err)
		}

		recv, err := stream.Recv()
		if err != nil {
			if xerrors.Is(err, io.EOF) {
				break
			}

			return nil, nil, xerrors.WithStackTrace(err)
		}

		if recv.GetTxMeta() != nil {
			t = &transaction{
				Identifier: tx.ID(recv.GetTxMeta().GetId()),
				control:    table.TxControl(table.WithTxID(recv.GetTxMeta().GetId())),
			}
		}

		if recv.GetExecStats() != nil {
			queryStats = recv.GetExecStats()
		}

		if rs := recv.GetResultSet(); rs != nil {
			if idx := int(recv.GetResultSetIndex()); idx == len(resultSets) {
				resultSets = append(resultSets, recv.GetResultSet())
			} else if idx < len(resultSets) {
				resultSets[idx].Rows = append(resultSets[idx].GetRows(), recv.GetResultSet().GetRows()...)
			} else {
				return nil, nil, xerrors.WithStackTrace(fmt.Errorf("unexpected result set index: %d", idx))
			}
		}
	}

	return t, scanner.NewUnary(
		resultSets,
		queryStats,
		scanner.WithIgnoreTruncated(false),
	), nil
}

func (e queryClientExecutor) execute(
	ctx context.Context,
	txControl *tx.Control,
	executeDataQueryRequest *Ydb_Table.ExecuteDataQueryRequest,
	callOptions ...grpc.CallOption,
) (_ *transaction, _ result.Result, finalErr error) {
	request := &Ydb_Query.ExecuteQueryRequest{}

	request.SessionId = executeDataQueryRequest.GetSessionId()
	request.ExecMode = Ydb_Query.ExecMode_EXEC_MODE_EXECUTE
	request.TxControl = txControl.ToYdbQueryTransactionControl()
	request.Query = &Ydb_Query.ExecuteQueryRequest_QueryContent{
		QueryContent: &Ydb_Query.QueryContent{
			Syntax: Ydb_Query.Syntax_SYNTAX_YQL_V1,
			Text:   executeDataQueryRequest.GetQuery().GetYqlText(),
		},
	}
	request.Parameters = executeDataQueryRequest.GetParameters()
	request.StatsMode = statsModeToStatsMode(executeDataQueryRequest.GetCollectStats())
	request.ConcurrentResultSets = false

	ctx, cancel := xcontext.WithDone(xcontext.ValueOnly(ctx), e.core.Done())
	defer cancel()

	stream, err := e.client.ExecuteQuery(ctx, request, callOptions...)
	if err != nil {
		if status := query.StatusFromErr(err); status != query.StatusUnknown {
			e.core.SetStatus(status)
		}

		return nil, nil, xerrors.WithStackTrace(err)
	}

	return queryExecuteStreamResultToTableResult(ctx, stream)
}

func (e tableClientExecutor) execute(
	ctx context.Context,
	txControl *tx.Control,
	request *Ydb_Table.ExecuteDataQueryRequest,
	callOptions ...grpc.CallOption,
) (*transaction, result.Result, error) {
	request.TxControl = txControl.ToYdbTableTransactionControl()

	r, err := executeDataQuery(ctx, e.client, request, callOptions...)
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	return executeQueryResult(r, request.GetTxControl(), e.ignoreTruncated)
}

var (
	_ dataQueryExecutor = (*tableClientExecutor)(nil)
	_ dataQueryExecutor = (*queryClientExecutor)(nil)
)

// Session represents a single table API session.
//
// session methods are not goroutine safe. Simultaneous execution of requests
// are forbidden within a single session.
//
// Note that after session is no longer needed it should be destroyed by
// Close() call.
type Session struct {
	closeOnce func(ctx context.Context) error
	id        string
	client    Ydb_Table_V1.TableServiceClient
	status    table.SessionStatus
	config    *config.Config
	dataQuery dataQueryExecutor
	lastUsage atomic.Int64
	statusMtx sync.RWMutex
	nodeID    atomic.Uint32
}

func (s *Session) IsAlive() bool {
	return s.Status() == table.SessionReady
}

func (s *Session) LastUsage() time.Time {
	return time.Unix(s.lastUsage.Load(), 0)
}

func nodeID(sessionID string) (uint32, error) {
	u, err := url.Parse(sessionID)
	if err != nil {
		return 0, err
	}
	id, err := strconv.ParseUint(u.Query().Get("node_id"), 10, 32)
	if err != nil {
		return 0, err
	}

	return uint32(id), err
}

func (s *Session) NodeID() uint32 {
	if s == nil {
		return 0
	}
	if id := s.nodeID.Load(); id != 0 {
		return id
	}
	id, err := nodeID(s.id)
	if err != nil {
		return 0
	}
	s.nodeID.Store(id)

	return id
}

func (s *Session) Status() table.SessionStatus {
	if s == nil {
		return table.SessionStatusUnknown
	}
	s.statusMtx.RLock()
	defer s.statusMtx.RUnlock()

	return s.status
}

func (s *Session) SetStatus(status table.SessionStatus) {
	s.statusMtx.Lock()
	defer s.statusMtx.Unlock()
	s.status = status
}

func newSession(ctx context.Context, cc grpc.ClientConnInterface, config *config.Config) (
	s *Session, finalErr error,
) {
	onDone := trace.TableOnSessionNew(config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.newSession"),
	)
	defer func() {
		onDone(s, finalErr)
	}()

	if config.UseQuerySession() {
		return newQuerySession(ctx, cc, config)
	}

	return newTableSession(ctx, cc, config)
}

func newTableSession(
	ctx context.Context, cc grpc.ClientConnInterface, config *config.Config,
) (*Session, error) {
	response, err := Ydb_Table_V1.NewTableServiceClient(cc).CreateSession(ctx,
		&Ydb_Table.CreateSessionRequest{
			OperationParams: operation.Params(
				ctx,
				config.OperationTimeout(),
				config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	var result Ydb_Table.CreateSessionResult
	if err := response.GetOperation().GetResult().UnmarshalTo(&result); err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	s := &Session{
		id:     result.GetSessionId(),
		config: config,
		status: table.SessionReady,
	}

	s.lastUsage.Store(time.Now().Unix())
	s.client = Ydb_Table_V1.NewTableServiceClient(
		conn.WithBeforeFunc(
			conn.WithContextModifier(cc, func(ctx context.Context) context.Context {
				return meta.WithTrailerCallback(balancerContext.WithNodeID(ctx, s.NodeID()), s.checkCloseHint)
			}),
			func() {
				s.lastUsage.Store(time.Now().Unix())
			},
		),
	)
	s.closeOnce = xsync.OnceFunc(closeTableSession(s.client, s.config, s.id))
	s.dataQuery = tableClientExecutor{
		client:          s.client,
		ignoreTruncated: s.config.IgnoreTruncated(),
	}

	return s, nil
}

func closeTableSession(c Ydb_Table_V1.TableServiceClient, cfg *config.Config, id string) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return xerrors.WithStackTrace(err)
		}

		if t := cfg.DeleteTimeout(); t > 0 {
			var cancel context.CancelFunc
			ctx, cancel = xcontext.WithTimeout(ctx, t)
			defer cancel()
		}

		_, err := c.DeleteSession(ctx,
			&Ydb_Table.DeleteSessionRequest{
				SessionId: id,
				OperationParams: operation.Params(ctx,
					cfg.OperationTimeout(),
					cfg.OperationCancelAfter(),
					operation.ModeSync,
				),
			},
		)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}
}

func newQuerySession(
	ctx context.Context, cc grpc.ClientConnInterface, config *config.Config,
) (*Session, error) {
	s := &Session{
		config: config,
		status: table.SessionReady,
	}

	core, err := query.Open(ctx,
		Ydb_Query_V1.NewQueryServiceClient(cc),
		query.WithConn(cc),
		query.OnChangeStatus(func(status query.Status) {
			switch status {
			case query.StatusClosed:
				s.SetStatus(table.SessionClosed)
			case query.StatusClosing:
				s.SetStatus(table.SessionClosing)
			case query.StatusInUse:
				s.SetStatus(table.SessionBusy)
			case query.StatusIdle:
				s.SetStatus(table.SessionReady)
			default:
				s.SetStatus(table.SessionStatusUnknown)
			}
		}),
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	s.id = core.ID()
	s.lastUsage.Store(time.Now().Unix())
	s.client = Ydb_Table_V1.NewTableServiceClient(
		conn.WithBeforeFunc(
			conn.WithContextModifier(cc, func(ctx context.Context) context.Context {
				return meta.WithTrailerCallback(balancerContext.WithNodeID(ctx, s.NodeID()), s.checkCloseHint)
			}),
			func() {
				s.lastUsage.Store(time.Now().Unix())
			},
		),
	)
	s.closeOnce = xsync.OnceFunc(closeQuerySession(core))
	if config.ExecuteDataQueryOverQueryService() {
		s.dataQuery = queryClientExecutor{
			core:   core,
			client: core.Client,
		}
	} else {
		s.dataQuery = tableClientExecutor{
			client:          s.client,
			ignoreTruncated: s.config.IgnoreTruncated(),
		}
	}

	return s, nil
}

func closeQuerySession(core query.Core) func(context.Context) error {
	return func(ctx context.Context) error {
		err := core.Close(ctx)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		return nil
	}
}

func (s *Session) ID() string {
	if s == nil {
		return ""
	}

	return s.id
}

func (s *Session) Close(ctx context.Context) (finalErr error) {
	onDone := trace.TableOnSessionDelete(s.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).Close"),
		s,
	)
	defer func() {
		onDone(finalErr)
		s.SetStatus(table.SessionClosed)
	}()

	if s.closeOnce != nil {
		err := s.closeOnce(ctx)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
	}

	return nil
}

func (s *Session) checkCloseHint(md metadata.MD) {
	for header, values := range md {
		if header != meta.HeaderServerHints {
			continue
		}
		for _, hint := range values {
			if hint == meta.HintSessionClose {
				s.SetStatus(table.SessionClosing)
			}
		}
	}
}

// KeepAlive keeps idle session alive.
func (s *Session) KeepAlive(ctx context.Context) (err error) {
	var (
		result Ydb_Table.KeepAliveResult
		onDone = trace.TableOnSessionKeepAlive(
			s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).KeepAlive"),
			s,
		)
	)
	defer func() {
		onDone(err)
	}()

	resp, err := s.client.KeepAlive(ctx,
		&Ydb_Table.KeepAliveRequest{
			SessionId: s.id,
			OperationParams: operation.Params(
				ctx,
				s.config.OperationTimeout(),
				s.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	err = resp.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	switch result.GetSessionStatus() {
	case Ydb_Table.KeepAliveResult_SESSION_STATUS_READY:
		s.SetStatus(table.SessionReady)
	case Ydb_Table.KeepAliveResult_SESSION_STATUS_BUSY:
		s.SetStatus(table.SessionBusy)
	}

	return nil
}

// CreateTable creates table at given path with given options.
func (s *Session) CreateTable(
	ctx context.Context,
	path string,
	opts ...options.CreateTableOption,
) (err error) {
	request := Ydb_Table.CreateTableRequest{
		SessionId: s.id,
		Path:      path,
		OperationParams: operation.Params(
			ctx,
			s.config.OperationTimeout(),
			s.config.OperationCancelAfter(),
			operation.ModeSync,
		),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyCreateTableOption((*options.CreateTableDesc)(&request))
		}
	}
	_, err = s.client.CreateTable(ctx, &request)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

type describeTableClient interface {
	DescribeTable(
		ctx context.Context, in *Ydb_Table.DescribeTableRequest, opts ...grpc.CallOption,
	) (*Ydb_Table.DescribeTableResponse, error)
}

// DescribeTable describes table at given path.
func DescribeTable(
	ctx context.Context,
	sessionID string,
	client describeTableClient,
	path string,
	opts ...options.DescribeTableOption,
) (desc options.Description, err error) {
	request := describeTableRequest(ctx, sessionID, path, opts)
	response, err := client.DescribeTable(ctx, request)
	if err != nil {
		return desc, xerrors.WithStackTrace(err)
	}

	var result Ydb_Table.DescribeTableResult
	if err = response.GetOperation().GetResult().UnmarshalTo(&result); err != nil {
		return desc, xerrors.WithStackTrace(err)
	}

	desc = options.Description{
		Name:                 result.GetSelf().GetName(),
		PrimaryKey:           result.GetPrimaryKey(),
		Columns:              processColumns(result.GetColumns()),
		KeyRanges:            processKeyRanges(result.GetShardKeyBounds()),
		Stats:                processTableStats(result.GetTableStats()),
		ColumnFamilies:       processColumnFamilies(result.GetColumnFamilies()),
		Attributes:           processAttributes(result.GetAttributes()),
		ReadReplicaSettings:  options.NewReadReplicasSettings(result.GetReadReplicasSettings()),
		StorageSettings:      options.NewStorageSettings(result.GetStorageSettings()),
		KeyBloomFilter:       feature.FromYDB(result.GetKeyBloomFilter()),
		PartitioningSettings: options.NewPartitioningSettings(result.GetPartitioningSettings()),
		Indexes:              processIndexes(result.GetIndexes()),
		TimeToLiveSettings:   NewTimeToLiveSettings(result.GetTtlSettings()),
		Changefeeds:          processChangefeeds(result.GetChangefeeds()),
		Tiering:              result.GetTiering(),
		StoreType:            options.StoreType(result.GetStoreType()),
	}

	return desc, nil
}

// DescribeTable describes table at given path.
func (s *Session) DescribeTable(
	ctx context.Context,
	path string,
	opts ...options.DescribeTableOption,
) (options.Description, error) {
	desc, err := DescribeTable(ctx, s.id, s.client, path, opts...)
	if err != nil {
		return desc, xerrors.WithStackTrace(err)
	}

	return desc, nil
}

func describeTableRequest(
	ctx context.Context,
	sessionID string,
	path string,
	opts []options.DescribeTableOption,
) *Ydb_Table.DescribeTableRequest {
	request := Ydb_Table.DescribeTableRequest{
		SessionId:       sessionID,
		Path:            path,
		OperationParams: operation.Params(ctx, 0, 0, operation.ModeSync),
	}
	for _, opt := range opts {
		if opt != nil {
			opt((*options.DescribeTableDesc)(&request))
		}
	}

	return &request
}

func processColumns(columns []*Ydb_Table.ColumnMeta) []options.Column {
	cs := make([]options.Column, len(columns))
	for i, c := range columns {
		cs[i] = options.Column{
			Name:   c.GetName(),
			Type:   types.TypeFromYDB(c.GetType()),
			Family: c.GetFamily(),
		}
	}

	return cs
}

func processKeyRanges(bounds []*Ydb.TypedValue) []options.KeyRange {
	rs := make([]options.KeyRange, len(bounds)+1)
	var last value.Value
	for i, b := range bounds {
		if last != nil {
			rs[i].From = last
		}

		bound := value.FromYDB(b.GetType(), b.GetValue())
		rs[i].To = bound

		last = bound
	}
	if last != nil {
		i := len(rs) - 1
		rs[i].From = last
	}

	return rs
}

func processTableStats(resStats *Ydb_Table.TableStats) *options.TableStats {
	if resStats == nil {
		return nil
	}

	partStats := make([]options.PartitionStats, len(resStats.GetPartitionStats()))
	for i, v := range resStats.GetPartitionStats() {
		partStats[i].RowsEstimate = v.GetRowsEstimate()
		partStats[i].StoreSize = v.GetStoreSize()
		partStats[i].LeaderNodeID = v.GetLeaderNodeId()
	}

	var creationTime, modificationTime time.Time
	if resStats.GetCreationTime().GetSeconds() != 0 {
		creationTime = time.Unix(resStats.GetCreationTime().GetSeconds(), int64(resStats.GetCreationTime().GetNanos()))
	}
	if resStats.GetModificationTime().GetSeconds() != 0 {
		modificationTime = time.Unix(
			resStats.GetModificationTime().GetSeconds(),
			int64(resStats.GetModificationTime().GetNanos()),
		)
	}

	return &options.TableStats{
		PartitionStats:   partStats,
		RowsEstimate:     resStats.GetRowsEstimate(),
		StoreSize:        resStats.GetStoreSize(),
		Partitions:       resStats.GetPartitions(),
		CreationTime:     creationTime,
		ModificationTime: modificationTime,
	}
}

func processColumnFamilies(families []*Ydb_Table.ColumnFamily) []options.ColumnFamily {
	cf := make([]options.ColumnFamily, len(families))
	for i, c := range families {
		cf[i] = options.NewColumnFamily(c)
	}

	return cf
}

func processAttributes(attrs map[string]string) map[string]string {
	attributes := make(map[string]string, len(attrs))
	for k, v := range attrs {
		attributes[k] = v
	}

	return attributes
}

func processIndexes(indexes []*Ydb_Table.TableIndexDescription) []options.IndexDescription {
	idxs := make([]options.IndexDescription, len(indexes))
	for i, idx := range indexes {
		var typ options.IndexType
		switch idx.GetType().(type) {
		case *Ydb_Table.TableIndexDescription_GlobalAsyncIndex:
			typ = options.IndexTypeGlobalAsync
		case *Ydb_Table.TableIndexDescription_GlobalIndex:
			typ = options.IndexTypeGlobal
		}
		idxs[i] = options.IndexDescription{
			Name:         idx.GetName(),
			IndexColumns: idx.GetIndexColumns(),
			DataColumns:  idx.GetDataColumns(),
			Status:       idx.GetStatus(),
			Type:         typ,
		}
	}

	return idxs
}

func processChangefeeds(changefeeds []*Ydb_Table.ChangefeedDescription) []options.ChangefeedDescription {
	feeds := make([]options.ChangefeedDescription, len(changefeeds))
	for i, proto := range changefeeds {
		feeds[i] = options.NewChangefeedDescription(proto)
	}

	return feeds
}

// DropTable drops table at given path with given options.
func (s *Session) DropTable(
	ctx context.Context,
	path string,
	opts ...options.DropTableOption,
) (err error) {
	request := Ydb_Table.DropTableRequest{
		SessionId: s.id,
		Path:      path,
		OperationParams: operation.Params(
			ctx,
			s.config.OperationTimeout(),
			s.config.OperationCancelAfter(),
			operation.ModeSync,
		),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyDropTableOption((*options.DropTableDesc)(&request))
		}
	}
	_, err = s.client.DropTable(ctx, &request)

	return xerrors.WithStackTrace(err)
}

// AlterTable modifies schema of table at given path with given options.
func (s *Session) AlterTable(
	ctx context.Context,
	path string,
	opts ...options.AlterTableOption,
) (err error) {
	request := Ydb_Table.AlterTableRequest{
		SessionId: s.id,
		Path:      path,
		OperationParams: operation.Params(
			ctx,
			s.config.OperationTimeout(),
			s.config.OperationCancelAfter(),
			operation.ModeSync,
		),
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyAlterTableOption((*options.AlterTableDesc)(&request))
		}
	}
	_, err = s.client.AlterTable(ctx, &request)

	return xerrors.WithStackTrace(err)
}

// CopyTable creates copy of table at given path.
func (s *Session) CopyTable(
	ctx context.Context,
	dst, src string,
	opts ...options.CopyTableOption,
) (err error) {
	request := Ydb_Table.CopyTableRequest{
		SessionId:       s.id,
		SourcePath:      src,
		DestinationPath: dst,
		OperationParams: operation.Params(
			ctx,
			s.config.OperationTimeout(),
			s.config.OperationCancelAfter(),
			operation.ModeSync,
		),
	}
	for _, opt := range opts {
		if opt != nil {
			opt((*options.CopyTableDesc)(&request))
		}
	}
	_, err = s.client.CopyTable(ctx, &request)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func copyTables(
	ctx context.Context,
	sessionID string,
	operationTimeout time.Duration,
	operationCancelAfter time.Duration,
	service interface {
		CopyTables(
			ctx context.Context, in *Ydb_Table.CopyTablesRequest, opts ...grpc.CallOption,
		) (*Ydb_Table.CopyTablesResponse, error)
	},
	opts ...options.CopyTablesOption,
) (err error) {
	request := Ydb_Table.CopyTablesRequest{
		SessionId: sessionID,
		OperationParams: operation.Params(
			ctx,
			operationTimeout,
			operationCancelAfter,
			operation.ModeSync,
		),
	}
	for _, opt := range opts {
		if opt != nil {
			opt((*options.CopyTablesDesc)(&request))
		}
	}
	if len(request.GetTables()) == 0 {
		return xerrors.WithStackTrace(fmt.Errorf("no CopyTablesItem: %w", errParamsRequired))
	}
	_, err = service.CopyTables(ctx, &request)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// CopyTables creates copy of table at given path.
func (s *Session) CopyTables(
	ctx context.Context,
	opts ...options.CopyTablesOption,
) (err error) {
	err = copyTables(ctx, s.id, s.config.OperationTimeout(), s.config.OperationCancelAfter(), s.client, opts...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func renameTables(
	ctx context.Context,
	sessionID string,
	operationTimeout time.Duration,
	operationCancelAfter time.Duration,
	service interface {
		RenameTables(
			ctx context.Context, in *Ydb_Table.RenameTablesRequest, opts ...grpc.CallOption,
		) (*Ydb_Table.RenameTablesResponse, error)
	},
	opts ...options.RenameTablesOption,
) (err error) {
	request := Ydb_Table.RenameTablesRequest{
		SessionId: sessionID,
		OperationParams: operation.Params(
			ctx,
			operationTimeout,
			operationCancelAfter,
			operation.ModeSync,
		),
	}
	for _, opt := range opts {
		if opt != nil {
			opt((*options.RenameTablesDesc)(&request))
		}
	}
	if len(request.GetTables()) == 0 {
		return xerrors.WithStackTrace(fmt.Errorf("no RenameTablesItem: %w", errParamsRequired))
	}
	_, err = service.RenameTables(ctx, &request)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// RenameTables renames tables.
func (s *Session) RenameTables(
	ctx context.Context,
	opts ...options.RenameTablesOption,
) (err error) {
	err = renameTables(ctx, s.id, s.config.OperationTimeout(), s.config.OperationCancelAfter(), s.client, opts...)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// Explain explains data query represented by text.
func (s *Session) Explain(ctx context.Context, sql string) (exp table.DataQueryExplanation, err error) {
	var (
		result   Ydb_Table.ExplainQueryResult
		response *Ydb_Table.ExplainDataQueryResponse
		onDone   = trace.TableOnSessionQueryExplain(
			s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).Explain"),
			s, sql,
		)
	)
	defer func() {
		if err != nil {
			onDone("", "", err)
		} else {
			onDone(exp.AST, exp.AST, nil)
		}
	}()

	response, err = s.client.ExplainDataQuery(ctx,
		&Ydb_Table.ExplainDataQueryRequest{
			SessionId: s.id,
			YqlText:   sql,
			OperationParams: operation.Params(
				ctx,
				s.config.OperationTimeout(),
				s.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return exp, xerrors.WithStackTrace(err)
	}

	err = response.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return exp, xerrors.WithStackTrace(err)
	}

	return table.DataQueryExplanation{
		Explanation: table.Explanation{
			Plan: result.GetQueryPlan(),
		},
		AST: result.GetQueryAst(),
	}, nil
}

// Prepare prepares data query within session s.
func (s *Session) Prepare(ctx context.Context, queryText string) (_ table.Statement, err error) {
	var (
		stmt     *statement
		response *Ydb_Table.PrepareDataQueryResponse
		result   Ydb_Table.PrepareQueryResult
		onDone   = trace.TableOnSessionQueryPrepare(
			s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).Prepare"),
			s, queryText,
		)
	)
	defer func() {
		if err != nil {
			onDone(nil, err)
		} else {
			onDone(stmt.query, nil)
		}
	}()

	response, err = s.client.PrepareDataQuery(ctx,
		&Ydb_Table.PrepareDataQueryRequest{
			SessionId: s.id,
			YqlText:   queryText,
			OperationParams: operation.Params(
				ctx,
				s.config.OperationTimeout(),
				s.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	err = response.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	stmt = &statement{
		session: s,
		query:   queryPrepared(result.GetQueryId(), queryText),
		params:  result.GetParametersTypes(),
	}

	return stmt, nil
}

// Execute executes given data query represented by text.
func (s *Session) Execute(ctx context.Context, txControl *table.TransactionControl, sql string, params *params.Params,
	opts ...options.ExecuteDataQueryOption,
) (
	txr table.Transaction, r result.Result, err error,
) {
	var (
		q       = queryFromText(sql)
		request = options.ExecuteDataQueryDesc{
			ExecuteDataQueryRequest: &Ydb_Table.ExecuteDataQueryRequest{
				SessionId: s.id,
				TxControl: txControl.ToYdbTableTransactionControl(),
				Parameters: func() map[string]*Ydb.TypedValue {
					p, _ := params.ToYDB()

					return p
				}(),
				Query: q.toYDB(),
				QueryCachePolicy: &Ydb_Table.QueryCachePolicy{
					KeepInCache: true,
				},
			},
			IgnoreTruncated: s.config.IgnoreTruncated(),
		}
		callOptions []grpc.CallOption
	)

	request.OperationParams = operation.Params(ctx,
		s.config.OperationTimeout(),
		s.config.OperationCancelAfter(),
		operation.ModeSync,
	)

	for _, opt := range opts {
		if opt != nil {
			callOptions = append(callOptions, opt.ApplyExecuteDataQueryOption(&request)...)
		}
	}

	onDone := trace.TableOnSessionQueryExecute(
		s.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).Execute"),
		s, q, params,
		request.QueryCachePolicy.GetKeepInCache(),
	)
	defer func() {
		onDone(txr, false, r, err)
	}()

	t, r, err := s.dataQuery.execute(ctx, txControl, request.ExecuteDataQueryRequest, callOptions...)
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	if t != nil {
		t.s = s
	}

	return t, r, nil
}

// executeQueryResult returns Transaction and result built from received
// result.
func executeQueryResult(
	res *Ydb_Table.ExecuteQueryResult,
	txControl *Ydb_Table.TransactionControl,
	ignoreTruncated bool,
) (*transaction, result.Result, error) {
	tx := &transaction{
		Identifier: tx.ID(res.GetTxMeta().GetId()),
	}
	if txControl.GetCommitTx() {
		tx.state.Store(txStateCommitted)
	} else {
		tx.state.Store(txStateInitialized)
		tx.control = table.TxControl(table.WithTxID(tx.ID()))
	}

	return tx, scanner.NewUnary(
		res.GetResultSets(),
		res.GetQueryStats(),
		scanner.WithIgnoreTruncated(ignoreTruncated),
	), nil
}

// executeDataQuery executes data query.
func executeDataQuery(
	ctx context.Context, client Ydb_Table_V1.TableServiceClient,
	request *Ydb_Table.ExecuteDataQueryRequest,
	callOptions ...grpc.CallOption,
) (
	_ *Ydb_Table.ExecuteQueryResult,
	err error,
) {
	var (
		result   = &Ydb_Table.ExecuteQueryResult{}
		response *Ydb_Table.ExecuteDataQueryResponse
	)

	response, err = client.ExecuteDataQuery(ctx, request, callOptions...)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	err = response.GetOperation().GetResult().UnmarshalTo(result)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return result, nil
}

// ExecuteSchemeQuery executes scheme query.
func (s *Session) ExecuteSchemeQuery(ctx context.Context, sql string,
	opts ...options.ExecuteSchemeQueryOption,
) (err error) {
	request := Ydb_Table.ExecuteSchemeQueryRequest{
		SessionId: s.id,
		YqlText:   sql,
		OperationParams: operation.Params(
			ctx,
			s.config.OperationTimeout(),
			s.config.OperationCancelAfter(),
			operation.ModeSync,
		),
	}
	for _, opt := range opts {
		if opt != nil {
			opt((*options.ExecuteSchemeQueryDesc)(&request))
		}
	}
	_, err = s.client.ExecuteSchemeQuery(ctx, &request)

	return xerrors.WithStackTrace(err)
}

// DescribeTableOptions describes supported table options.
//
//nolint:funlen
func (s *Session) DescribeTableOptions(ctx context.Context) (
	desc options.TableOptionsDescription,
	err error,
) {
	var (
		response *Ydb_Table.DescribeTableOptionsResponse
		result   Ydb_Table.DescribeTableOptionsResult
	)
	request := Ydb_Table.DescribeTableOptionsRequest{
		OperationParams: operation.Params(
			ctx,
			s.config.OperationTimeout(),
			s.config.OperationCancelAfter(),
			operation.ModeSync,
		),
	}
	response, err = s.client.DescribeTableOptions(ctx, &request)
	if err != nil {
		return desc, xerrors.WithStackTrace(err)
	}

	err = response.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return desc, xerrors.WithStackTrace(err)
	}

	{
		xs := make([]options.TableProfileDescription, len(result.GetTableProfilePresets()))
		for i, p := range result.GetTableProfilePresets() {
			xs[i] = options.TableProfileDescription{
				Name:   p.GetName(),
				Labels: p.GetLabels(),

				DefaultStoragePolicy:      p.GetDefaultStoragePolicy(),
				DefaultCompactionPolicy:   p.GetDefaultCompactionPolicy(),
				DefaultPartitioningPolicy: p.GetDefaultPartitioningPolicy(),
				DefaultExecutionPolicy:    p.GetDefaultExecutionPolicy(),
				DefaultReplicationPolicy:  p.GetDefaultReplicationPolicy(),
				DefaultCachingPolicy:      p.GetDefaultCachingPolicy(),

				AllowedStoragePolicies:      p.GetAllowedStoragePolicies(),
				AllowedCompactionPolicies:   p.GetAllowedCompactionPolicies(),
				AllowedPartitioningPolicies: p.GetAllowedPartitioningPolicies(),
				AllowedExecutionPolicies:    p.GetAllowedExecutionPolicies(),
				AllowedReplicationPolicies:  p.GetAllowedReplicationPolicies(),
				AllowedCachingPolicies:      p.GetAllowedCachingPolicies(),
			}
		}
		desc.TableProfilePresets = xs
	}
	{
		xs := make(
			[]options.StoragePolicyDescription,
			len(result.GetStoragePolicyPresets()),
		)
		for i, p := range result.GetStoragePolicyPresets() {
			xs[i] = options.StoragePolicyDescription{
				Name:   p.GetName(),
				Labels: p.GetLabels(),
			}
		}
		desc.StoragePolicyPresets = xs
	}
	{
		xs := make(
			[]options.CompactionPolicyDescription,
			len(result.GetCompactionPolicyPresets()),
		)
		for i, p := range result.GetCompactionPolicyPresets() {
			xs[i] = options.CompactionPolicyDescription{
				Name:   p.GetName(),
				Labels: p.GetLabels(),
			}
		}
		desc.CompactionPolicyPresets = xs
	}
	{
		xs := make(
			[]options.PartitioningPolicyDescription,
			len(result.GetPartitioningPolicyPresets()),
		)
		for i, p := range result.GetPartitioningPolicyPresets() {
			xs[i] = options.PartitioningPolicyDescription{
				Name:   p.GetName(),
				Labels: p.GetLabels(),
			}
		}
		desc.PartitioningPolicyPresets = xs
	}
	{
		xs := make(
			[]options.ExecutionPolicyDescription,
			len(result.GetExecutionPolicyPresets()),
		)
		for i, p := range result.GetExecutionPolicyPresets() {
			xs[i] = options.ExecutionPolicyDescription{
				Name:   p.GetName(),
				Labels: p.GetLabels(),
			}
		}
		desc.ExecutionPolicyPresets = xs
	}
	{
		xs := make(
			[]options.ReplicationPolicyDescription,
			len(result.GetReplicationPolicyPresets()),
		)
		for i, p := range result.GetReplicationPolicyPresets() {
			xs[i] = options.ReplicationPolicyDescription{
				Name:   p.GetName(),
				Labels: p.GetLabels(),
			}
		}
		desc.ReplicationPolicyPresets = xs
	}
	{
		xs := make(
			[]options.CachingPolicyDescription,
			len(result.GetCachingPolicyPresets()),
		)
		for i, p := range result.GetCachingPolicyPresets() {
			xs[i] = options.CachingPolicyDescription{
				Name:   p.GetName(),
				Labels: p.GetLabels(),
			}
		}
		desc.CachingPolicyPresets = xs
	}

	return desc, nil
}

// StreamReadTable reads table at given path with given options.
//
// Note that given ctx controls the lifetime of the whole read, not only this
// StreamReadTable() call; that is, the time until returned result is closed
// via Close() call or fully drained by sequential NextResultSet() calls.
//
//nolint:funlen
func (s *Session) StreamReadTable(
	ctx context.Context,
	path string,
	opts ...options.ReadTableOption,
) (_ result.StreamResult, err error) {
	var (
		onDone = trace.TableOnSessionQueryStreamRead(s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).StreamReadTable"),
			s,
		)
		request = Ydb_Table.ReadTableRequest{
			SessionId: s.id,
			Path:      path,
		}
		stream Ydb_Table_V1.TableService_StreamReadTableClient
	)
	defer func() {
		onDone(xerrors.HideEOF(err))
	}()

	for _, opt := range opts {
		if opt != nil {
			opt.ApplyReadTableOption((*options.ReadTableDesc)(&request))
		}
	}

	ctx, cancel := xcontext.WithCancel(ctx)

	stream, err = s.client.StreamReadTable(ctx, &request)
	if err != nil {
		cancel()

		return nil, xerrors.WithStackTrace(err)
	}

	return scanner.NewStream(ctx,
		func(ctx context.Context) (
			set *Ydb.ResultSet,
			stats *Ydb_TableStats.QueryStats,
			err error,
		) {
			select {
			case <-ctx.Done():
				return nil, nil, xerrors.WithStackTrace(ctx.Err())
			default:
				var response *Ydb_Table.ReadTableResponse
				response, err = stream.Recv()
				result := response.GetResult()
				if result == nil || err != nil {
					return nil, nil, xerrors.WithStackTrace(err)
				}

				return result.GetResultSet(), nil, nil
			}
		},
		func(err error) error {
			cancel()
			onDone(xerrors.HideEOF(err))

			return err
		},
		scanner.WithIgnoreTruncated(true), // stream read table always returns truncated flag on last result set
	)
}

func (s *Session) ReadRows(
	ctx context.Context,
	path string,
	keys value.Value,
	opts ...options.ReadRowsOption,
) (_ result.Result, err error) {
	var (
		request  = makeReadRowsRequest(s.id, path, keys, opts)
		response *Ydb_Table.ReadRowsResponse
	)

	response, err = s.client.ReadRows(ctx, request)

	return makeReadRowsResponse(response, err, s.config.IgnoreTruncated())
}

// StreamExecuteScanQuery scan-reads table at given path with given options.
//
// Note that given ctx controls the lifetime of the whole read, not only this
// StreamExecuteScanQuery() call; that is, the time until returned result is closed
// via Close() call or fully drained by sequential NextResultSet() calls.
//
//nolint:funlen
func (s *Session) StreamExecuteScanQuery(ctx context.Context, sql string, parameters *params.Params,
	opts ...options.ExecuteScanQueryOption,
) (_ result.StreamResult, err error) {
	var (
		q      = queryFromText(sql)
		onDone = trace.TableOnSessionQueryStreamExecute(
			s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).StreamExecuteScanQuery"),
			s, q, parameters,
		)
		request = Ydb_Table.ExecuteScanQueryRequest{
			Query: q.toYDB(),
			Mode:  Ydb_Table.ExecuteScanQueryRequest_MODE_EXEC, // set default
		}
		stream      Ydb_Table_V1.TableService_StreamExecuteScanQueryClient
		callOptions []grpc.CallOption
	)
	defer func() {
		onDone(xerrors.HideEOF(err))
	}()

	params, err := parameters.ToYDB()
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	request.Parameters = params

	for _, opt := range opts {
		if opt != nil {
			callOptions = append(callOptions, opt.ApplyExecuteScanQueryOption((*options.ExecuteScanQueryDesc)(&request))...)
		}
	}

	ctx, cancel := xcontext.WithCancel(ctx)

	stream, err = s.client.StreamExecuteScanQuery(ctx, &request, callOptions...)
	if err != nil {
		cancel()

		return nil, xerrors.WithStackTrace(err)
	}

	return scanner.NewStream(ctx,
		func(ctx context.Context) (
			set *Ydb.ResultSet,
			stats *Ydb_TableStats.QueryStats,
			err error,
		) {
			select {
			case <-ctx.Done():
				return nil, nil, xerrors.WithStackTrace(ctx.Err())
			default:
				var response *Ydb_Table.ExecuteScanQueryPartialResponse
				response, err = stream.Recv()
				result := response.GetResult()
				if result == nil || err != nil {
					return nil, nil, xerrors.WithStackTrace(err)
				}

				return result.GetResultSet(), result.GetQueryStats(), nil
			}
		},
		func(err error) error {
			cancel()
			onDone(xerrors.HideEOF(err))

			return err
		},
		scanner.WithIgnoreTruncated(s.config.IgnoreTruncated()),
		scanner.WithMarkTruncatedAsRetryable(),
	)
}

// BulkUpsert uploads given list of ydb struct values to the table.
func (s *Session) BulkUpsert(ctx context.Context, table string, rows value.Value,
	opts ...options.BulkUpsertOption,
) (err error) {
	var (
		callOptions []grpc.CallOption
		onDone      = trace.TableOnSessionBulkUpsert(s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).BulkUpsert"), s,
		)
	)
	defer func() {
		onDone(err)
	}()

	for _, opt := range opts {
		if opt != nil {
			callOptions = append(callOptions, opt.ApplyBulkUpsertOption()...)
		}
	}

	_, err = s.client.BulkUpsert(ctx,
		&Ydb_Table.BulkUpsertRequest{
			Table: table,
			Rows:  value.ToYDB(rows),
			OperationParams: operation.Params(
				ctx,
				s.config.OperationTimeout(),
				s.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
		callOptions...,
	)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

// BeginTransaction begins new transaction within given session with given settings.
func (s *Session) BeginTransaction(
	ctx context.Context,
	txSettings *table.TransactionSettings,
) (x table.Transaction, err error) {
	var (
		result   Ydb_Table.BeginTransactionResult
		response *Ydb_Table.BeginTransactionResponse
		onDone   = trace.TableOnTxBegin(
			s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*Session).BeginTransaction"),
			s,
		)
	)
	defer func() {
		onDone(x, err)
	}()

	response, err = s.client.BeginTransaction(ctx,
		&Ydb_Table.BeginTransactionRequest{
			SessionId:  s.id,
			TxSettings: txSettings.ToYdbTableSettings(),
			OperationParams: operation.Params(
				ctx,
				s.config.OperationTimeout(),
				s.config.OperationCancelAfter(),
				operation.ModeSync,
			),
		},
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}
	err = response.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}
	tx := &transaction{
		Identifier: tx.ID(result.GetTxMeta().GetId()),
		s:          s,
		control:    table.TxControl(table.WithTxID(result.GetTxMeta().GetId())),
	}
	tx.state.Store(txStateInitialized)

	return tx, nil
}
