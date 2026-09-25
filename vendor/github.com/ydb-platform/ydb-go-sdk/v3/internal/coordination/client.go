package coordination

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Coordination_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Coordination"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
	"google.golang.org/grpc"

	"github.com/ydb-platform/ydb-go-sdk/v3/coordination"
	"github.com/ydb-platform/ydb-go-sdk/v3/coordination/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/scheme"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

//go:generate mockgen -destination grpc_client_mock_test.go --typed -package coordination -write_package_comment=false github.com/ydb-platform/ydb-go-genproto/Ydb_Coordination_V1 CoordinationServiceClient,CoordinationService_SessionClient

var errNilClient = xerrors.Wrap(errors.New("coordination client is not initialized"))

type Client struct {
	config config.Config
	client Ydb_Coordination_V1.CoordinationServiceClient

	mutex    sync.Mutex // guards the fields below
	sessions map[*session]struct{}
}

func New(ctx context.Context, cc grpc.ClientConnInterface, config config.Config) *Client {
	onDone := trace.CoordinationOnNew(config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination.New"),
	)
	defer onDone()

	return &Client{
		config:   config,
		client:   Ydb_Coordination_V1.NewCoordinationServiceClient(cc),
		sessions: make(map[*session]struct{}),
	}
}

func operationParams(
	ctx context.Context,
	config interface {
		OperationTimeout() time.Duration
		OperationCancelAfter() time.Duration
	},
	mode operation.Mode,
) *Ydb_Operations.OperationParams {
	return operation.Params(
		ctx,
		config.OperationTimeout(),
		config.OperationCancelAfter(),
		mode,
	)
}

func createNodeRequest(
	path string, config coordination.NodeConfig, operationParams *Ydb_Operations.OperationParams,
) *Ydb_Coordination.CreateNodeRequest {
	return &Ydb_Coordination.CreateNodeRequest{
		Path: path,
		Config: &Ydb_Coordination.Config{
			Path:                     config.Path,
			SelfCheckPeriodMillis:    config.SelfCheckPeriodMillis,
			SessionGracePeriodMillis: config.SessionGracePeriodMillis,
			ReadConsistencyMode:      config.ReadConsistencyMode.To(),
			AttachConsistencyMode:    config.AttachConsistencyMode.To(),
			RateLimiterCountersMode:  config.RatelimiterCountersMode.To(),
		},
		OperationParams: operationParams,
	}
}

func createNode(
	ctx context.Context, client Ydb_Coordination_V1.CoordinationServiceClient, request *Ydb_Coordination.CreateNodeRequest,
) error {
	_, err := client.CreateNode(ctx, request)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (c *Client) CreateNode(ctx context.Context, path string, config coordination.NodeConfig) (finalErr error) {
	if c == nil {
		return xerrors.WithStackTrace(errNilClient)
	}

	onDone := trace.CoordinationOnCreateNode(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination.(*Client).CreateNode"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()

	request := createNodeRequest(path, config, operationParams(ctx, &c.config, operation.ModeSync))

	if !c.config.AutoRetry() {
		return createNode(ctx, c.client, request)
	}

	return retry.Retry(ctx,
		func(ctx context.Context) error {
			return createNode(ctx, c.client, request)
		},
		retry.WithStackTrace(),
		retry.WithIdempotent(true),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)
}

func (c *Client) AlterNode(ctx context.Context, path string, config coordination.NodeConfig) (finalErr error) {
	if c == nil {
		return xerrors.WithStackTrace(errNilClient)
	}

	onDone := trace.CoordinationOnAlterNode(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination.(*Client).AlterNode"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()

	request := alterNodeRequest(path, config, operationParams(ctx, &c.config, operation.ModeSync))

	call := func(ctx context.Context) error {
		return alterNode(ctx, c.client, request)
	}
	if !c.config.AutoRetry() {
		return xerrors.WithStackTrace(call(ctx))
	}

	return retry.Retry(ctx,
		func(ctx context.Context) (err error) {
			return alterNode(ctx, c.client, request)
		},
		retry.WithStackTrace(),
		retry.WithIdempotent(true),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)
}

func alterNodeRequest(
	path string, config coordination.NodeConfig, operationParams *Ydb_Operations.OperationParams,
) *Ydb_Coordination.AlterNodeRequest {
	return &Ydb_Coordination.AlterNodeRequest{
		Path: path,
		Config: &Ydb_Coordination.Config{
			Path:                     config.Path,
			SelfCheckPeriodMillis:    config.SelfCheckPeriodMillis,
			SessionGracePeriodMillis: config.SessionGracePeriodMillis,
			ReadConsistencyMode:      config.ReadConsistencyMode.To(),
			AttachConsistencyMode:    config.AttachConsistencyMode.To(),
			RateLimiterCountersMode:  config.RatelimiterCountersMode.To(),
		},
		OperationParams: operationParams,
	}
}

func alterNode(
	ctx context.Context, client Ydb_Coordination_V1.CoordinationServiceClient, request *Ydb_Coordination.AlterNodeRequest,
) error {
	_, err := client.AlterNode(ctx, request)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (c *Client) DropNode(ctx context.Context, path string) (finalErr error) {
	if c == nil {
		return xerrors.WithStackTrace(errNilClient)
	}

	onDone := trace.CoordinationOnDropNode(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination.(*Client).DropNode"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()

	request := dropNodeRequest(path, operationParams(ctx, &c.config, operation.ModeSync))

	call := func(ctx context.Context) error {
		return dropNode(ctx, c.client, request)
	}
	if !c.config.AutoRetry() {
		return xerrors.WithStackTrace(call(ctx))
	}

	return retry.Retry(ctx,
		func(ctx context.Context) (err error) {
			return dropNode(ctx, c.client, request)
		},
		retry.WithStackTrace(),
		retry.WithIdempotent(true),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)
}

func dropNodeRequest(path string, operationParams *Ydb_Operations.OperationParams) *Ydb_Coordination.DropNodeRequest {
	return &Ydb_Coordination.DropNodeRequest{
		Path:            path,
		OperationParams: operationParams,
	}
}

func dropNode(
	ctx context.Context, client Ydb_Coordination_V1.CoordinationServiceClient, request *Ydb_Coordination.DropNodeRequest,
) error {
	_, err := client.DropNode(ctx, request)
	if err != nil {
		return xerrors.WithStackTrace(err)
	}

	return nil
}

func (c *Client) DescribeNode(
	ctx context.Context,
	path string,
) (
	entry *scheme.Entry,
	config *coordination.NodeConfig,
	finalErr error,
) {
	if c == nil {
		return nil, nil, xerrors.WithStackTrace(errNilClient)
	}

	onDone := trace.CoordinationOnDescribeNode(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination.(*Client).DescribeNode"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()

	request := describeNodeRequest(path, operationParams(ctx, &c.config, operation.ModeSync))

	if !c.config.AutoRetry() {
		return describeNode(ctx, c.client, request)
	}

	err := retry.Retry(ctx,
		func(ctx context.Context) (err error) {
			entry, config, err = describeNode(ctx, c.client, request)
			if err != nil {
				return xerrors.WithStackTrace(err)
			}

			return nil
		},
		retry.WithStackTrace(),
		retry.WithIdempotent(true),
		retry.WithTrace(c.config.TraceRetry()),
		retry.WithBudget(c.config.RetryBudget()),
	)
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	return entry, config, nil
}

func describeNodeRequest(
	path string, operationParams *Ydb_Operations.OperationParams,
) *Ydb_Coordination.DescribeNodeRequest {
	return &Ydb_Coordination.DescribeNodeRequest{
		Path:            path,
		OperationParams: operationParams,
	}
}

// DescribeNode describes a coordination node
func describeNode(
	ctx context.Context,
	client Ydb_Coordination_V1.CoordinationServiceClient,
	request *Ydb_Coordination.DescribeNodeRequest,
) (
	_ *scheme.Entry,
	_ *coordination.NodeConfig,
	err error,
) {
	var (
		response *Ydb_Coordination.DescribeNodeResponse
		result   Ydb_Coordination.DescribeNodeResult
	)
	response, err = client.DescribeNode(ctx, request)
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	err = response.GetOperation().GetResult().UnmarshalTo(&result)
	if err != nil {
		return nil, nil, xerrors.WithStackTrace(err)
	}

	return scheme.InnerConvertEntry(result.GetSelf()), &coordination.NodeConfig{
		Path:                     result.GetConfig().GetPath(),
		SelfCheckPeriodMillis:    result.GetConfig().GetSelfCheckPeriodMillis(),
		SessionGracePeriodMillis: result.GetConfig().GetSessionGracePeriodMillis(),
		ReadConsistencyMode:      consistencyMode(result.GetConfig().GetReadConsistencyMode()),
		AttachConsistencyMode:    consistencyMode(result.GetConfig().GetAttachConsistencyMode()),
		RatelimiterCountersMode:  ratelimiterCountersMode(result.GetConfig().GetRateLimiterCountersMode()),
	}, nil
}

func applyToSession(t *trace.Coordination, opts ...options.SessionOption) sessionOption {
	c := defaultCreateSessionConfig()

	for _, o := range opts {
		if o != nil {
			o(c)
		}
	}

	return func(s *session) {
		s.trace = t
		s.description = c.Description
		s.sessionTimeout = c.SessionTimeout
		s.sessionStartTimeout = c.SessionStartTimeout
		s.sessionStopTimeout = c.SessionStopTimeout
		s.sessionKeepAliveTimeout = c.SessionKeepAliveTimeout
		s.sessionReconnectDelay = c.SessionReconnectDelay
	}
}

func (c *Client) sessionCreated(s *session) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.sessions[s] = struct{}{}
}

func (c *Client) sessionClosed(s *session) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.sessions, s)
}

func (c *Client) closeSessions(ctx context.Context) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for s := range c.sessions {
		s.Close(ctx)
	}
}

func defaultCreateSessionConfig() *options.CreateSessionOptions {
	return &options.CreateSessionOptions{
		Description:             "YDB Go SDK",
		SessionTimeout:          time.Second * 5, //nolint:mnd
		SessionStartTimeout:     time.Second * 1,
		SessionStopTimeout:      time.Second * 1,
		SessionKeepAliveTimeout: time.Second * 10,       //nolint:mnd
		SessionReconnectDelay:   time.Millisecond * 500, //nolint:mnd
	}
}

func (c *Client) Session(
	ctx context.Context,
	path string,
	opts ...options.SessionOption,
) (_ coordination.Session, finalErr error) {
	if c == nil {
		return nil, xerrors.WithStackTrace(errNilClient)
	}

	onDone := trace.CoordinationOnSession(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination.(*Client).Session"),
		path,
	)
	defer func() {
		onDone(finalErr)
	}()

	s, err := createSession(ctx, c.client, path,
		applyToSession(c.config.Trace(), opts...),
		func(s *session) {
			s.onCreate = append(s.onCreate, func(s *session) {
				c.sessionCreated(s)
			})
			s.onClose = append(s.onClose, func(s *session) {
				c.sessionClosed(s)
			})
		},
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	return s, nil
}

func (c *Client) Close(ctx context.Context) (finalErr error) {
	if c == nil {
		return xerrors.WithStackTrace(errNilClient)
	}

	onDone := trace.CoordinationOnClose(c.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination.(*Client).Close"),
	)
	defer func() {
		onDone(finalErr)
	}()

	c.closeSessions(ctx)

	return nil
}

func consistencyMode(t Ydb_Coordination.ConsistencyMode) coordination.ConsistencyMode {
	switch t {
	case Ydb_Coordination.ConsistencyMode_CONSISTENCY_MODE_STRICT:
		return coordination.ConsistencyModeStrict
	case Ydb_Coordination.ConsistencyMode_CONSISTENCY_MODE_RELAXED:
		return coordination.ConsistencyModeRelaxed
	default:
		return coordination.ConsistencyModeUnset
	}
}

func ratelimiterCountersMode(t Ydb_Coordination.RateLimiterCountersMode) coordination.RatelimiterCountersMode {
	switch t {
	case Ydb_Coordination.RateLimiterCountersMode_RATE_LIMITER_COUNTERS_MODE_AGGREGATED:
		return coordination.RatelimiterCountersModeAggregated
	case Ydb_Coordination.RateLimiterCountersMode_RATE_LIMITER_COUNTERS_MODE_DETAILED:
		return coordination.RatelimiterCountersModeDetailed
	default:
		return coordination.RatelimiterCountersModeUnset
	}
}
