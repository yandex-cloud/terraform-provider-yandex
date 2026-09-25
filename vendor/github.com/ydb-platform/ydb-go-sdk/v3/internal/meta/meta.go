package meta

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"google.golang.org/grpc/metadata"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/credentials"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/secret"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/version"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var pid = os.Getpid()

func New(
	database string,
	credentials credentials.Credentials,
	trace *trace.Driver,
	opts ...Option,
) *Meta {
	m := &Meta{
		pid:         strconv.Itoa(pid),
		trace:       trace,
		credentials: credentials,
		database:    database,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(m)
		}
	}

	return m
}

type Option func(m *Meta)

func WithApplicationNameOption(applicationName string) Option {
	return func(m *Meta) {
		m.applicationName = applicationName
	}
}

func WithRequestTypeOption(requestType string) Option {
	return func(m *Meta) {
		m.requestsType = requestType
	}
}

func AllowOption(feature string) Option {
	return func(m *Meta) {
		m.capabilities = append(m.capabilities, feature)
	}
}

func ForbidOption(feature string) Option {
	return func(m *Meta) {
		n := 0
		for _, capability := range m.capabilities {
			if capability != feature {
				m.capabilities[n] = capability
				n++
			}
		}
		m.capabilities = m.capabilities[:n]
	}
}

type Meta struct {
	pid             string
	trace           *trace.Driver
	credentials     credentials.Credentials
	database        string
	requestsType    string
	applicationName string
	capabilities    []string
}

func (m *Meta) meta(ctx context.Context) (_ metadata.MD, err error) {
	md, has := metadata.FromOutgoingContext(ctx)
	if !has {
		md = metadata.MD{}
	}

	md.Set(HeaderClientPid, m.pid)

	if len(md.Get(HeaderDatabase)) == 0 {
		md.Set(HeaderDatabase, m.database)
	}

	if len(md.Get(HeaderVersion)) == 0 {
		md.Set(HeaderVersion, version.FullVersion)
	}

	if m.requestsType != "" {
		if len(md.Get(HeaderRequestType)) == 0 {
			md.Set(HeaderRequestType, m.requestsType)
		}
	}

	if m.applicationName != "" {
		md.Append(HeaderApplicationName, m.applicationName)
	}

	if len(m.capabilities) > 0 {
		md.Append(HeaderClientCapabilities, m.capabilities...)
	}

	if m.credentials == nil {
		return md, nil
	}

	var token string
	done := trace.DriverOnGetCredentials(m.trace, &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/meta.(*Meta).meta"),
	)
	defer func() {
		done(secret.Token(token), err)
	}()

	token, err = m.credentials.Token(ctx)
	if err != nil {
		if stringer, ok := m.credentials.(fmt.Stringer); ok {
			return nil, xerrors.WithStackTrace(fmt.Errorf("%w: %s", err, stringer.String()))
		}

		return nil, xerrors.WithStackTrace(err)
	}

	md.Set(HeaderTicket, token)

	return md, nil
}

func (m *Meta) Context(ctx context.Context) (_ context.Context, err error) {
	md, err := m.meta(ctx)
	if err != nil {
		return ctx, xerrors.WithStackTrace(err)
	}

	return metadata.NewOutgoingContext(ctx, md), nil
}
