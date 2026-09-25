package ydb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"
	balancerConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/balancer/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/certificates"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/conn"
	coordinationConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/coordination/config"
	discoveryConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/discovery/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/dsn"
	queryConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/query/config"
	ratelimiterConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/ratelimiter/config"
	schemeConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/scheme/config"
	scriptingConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/scripting/config"
	tableConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/table/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql"
	"github.com/ydb-platform/ydb-go-sdk/v3/log"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

// Option contains configuration values for Driver
type Option func(ctx context.Context, d *Driver) error

func WithStaticCredentials(user, password string) Option {
	return func(ctx context.Context, d *Driver) error {
		d.userInfo = &dsn.UserInfo{
			User:     user,
			Password: password,
		}

		return nil
	}
}

func WithStaticCredentialsLogin(login string) Option {
	return func(ctx context.Context, d *Driver) error {
		if d.userInfo == nil {
			d.userInfo = &dsn.UserInfo{
				User: login,
			}
		} else {
			d.userInfo.User = login
		}

		return nil
	}
}

func WithStaticCredentialsPassword(password string) Option {
	return func(ctx context.Context, d *Driver) error {
		if d.userInfo == nil {
			d.userInfo = &dsn.UserInfo{
				Password: password,
			}
		} else {
			d.userInfo.Password = password
		}

		return nil
	}
}

// WithNodeAddressMutator applies mutator for node addresses from discovery.ListEndpoints response
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithNodeAddressMutator(mutator func(address string) string) Option {
	return func(ctx context.Context, d *Driver) error {
		d.discoveryOptions = append(d.discoveryOptions, discoveryConfig.WithAddressMutator(mutator))

		return nil
	}
}

func WithAccessTokenCredentials(accessToken string) Option {
	return WithCredentials(
		credentials.NewAccessTokenCredentials(
			accessToken,
			credentials.WithSourceInfo(
				"ydb.WithAccessTokenCredentials(accessToken)", // hide access token for logs
			),
		),
	)
}

// WithOauth2TokenExchangeCredentials adds credentials that exchange token using
// OAuth 2.0 token exchange protocol:
// https://www.rfc-editor.org/rfc/rfc8693
func WithOauth2TokenExchangeCredentials(
	opts ...credentials.Oauth2TokenExchangeCredentialsOption,
) Option {
	opts = append(opts, credentials.WithSourceInfo("ydb.WithOauth2TokenExchangeCredentials(opts)"))

	return WithCreateCredentialsFunc(func(context.Context) (credentials.Credentials, error) {
		return credentials.NewOauth2TokenExchangeCredentials(opts...)
	})
}

/*
WithOauth2TokenExchangeCredentialsFile adds credentials that exchange token using
OAuth 2.0 token exchange protocol:
https://www.rfc-editor.org/rfc/rfc8693
Config file must be a valid json file

Fields of json file

	grant-type:           [string] Grant type option (default: "urn:ietf:params:oauth:grant-type:token-exchange")
	res:                  [string | list of strings] Resource option (optional)
	aud:                  [string | list of strings] Audience option for token exchange request (optional)
	scope:                [string | list of strings] Scope option (optional)
	requested-token-type: [string] Requested token type option (default: "urn:ietf:params:oauth:token-type:access_token")
	subject-credentials:  [creds_json] Subject credentials options (optional)
	actor-credentials:    [creds_json] Actor credentials options (optional)
	token-endpoint:       [string] Token endpoint

Fields of creds_json (JWT):

	type:                 [string] Token source type. Set JWT
	alg:                  [string] Algorithm for JWT signature.
								   Supported algorithms can be listed
								   with GetSupportedOauth2TokenExchangeJwtAlgorithms()
	private-key:          [string] (Private) key in PEM format (RSA, EC) or Base64 format (HMAC) for JWT signature
	kid:                  [string] Key id JWT standard claim (optional)
	iss:                  [string] Issuer JWT standard claim (optional)
	sub:                  [string] Subject JWT standard claim (optional)
	aud:                  [string | list of strings] Audience JWT standard claim (optional)
	jti:                  [string] JWT ID JWT standard claim (optional)
	ttl:                  [string] Token TTL (default: 1h)

Fields of creds_json (FIXED):

	type:                 [string] Token source type. Set FIXED
	token:                [string] Token value
	token-type:           [string] Token type value. It will become
								   subject_token_type/actor_token_type parameter
								   in token exchange request (https://www.rfc-editor.org/rfc/rfc8693)
*/
func WithOauth2TokenExchangeCredentialsFile(
	configFilePath string,
	opts ...credentials.Oauth2TokenExchangeCredentialsOption,
) Option {
	srcInfo := credentials.WithSourceInfo(fmt.Sprintf("ydb.WithOauth2TokenExchangeCredentialsFile(%s)", configFilePath))
	opts = append(opts, srcInfo)

	return WithCreateCredentialsFunc(func(context.Context) (credentials.Credentials, error) {
		return credentials.NewOauth2TokenExchangeCredentialsFile(configFilePath, opts...)
	})
}

// WithApplicationName add provided application name to all api requests
func WithApplicationName(applicationName string) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithApplicationName(applicationName))

		return nil
	}
}

// WithUserAgent add provided user agent value to all api requests
//
// Deprecated: use WithApplicationName instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithUserAgent(userAgent string) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithApplicationName(userAgent))

		return nil
	}
}

func WithRequestsType(requestsType string) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithRequestsType(requestsType))

		return nil
	}
}

// WithConnectionString accept Driver string like
//
//	grpc[s]://{endpoint}/{database}[?param=value]
//
// Warning: WithConnectionString will be removed at next major release
//
// (Driver string will be required string param of ydb.Open)
func WithConnectionString(connectionString string) Option {
	return func(ctx context.Context, d *Driver) error {
		if connectionString == "" {
			return nil
		}
		info, err := dsn.Parse(connectionString)
		if err != nil {
			return xerrors.WithStackTrace(
				fmt.Errorf("parse connection string '%s' failed: %w", connectionString, err),
			)
		}
		d.options = append(d.options, info.Options...)
		d.userInfo = info.UserInfo

		return nil
	}
}

// WithConnectionTTL defines duration for parking idle connections
func WithConnectionTTL(ttl time.Duration) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithConnectionTTL(ttl))

		return nil
	}
}

// WithEndpoint defines endpoint option
//
// Warning: use ydb.Open with required Driver string parameter instead
//
// For making Driver string from endpoint+database+secure - use sugar.DSN()
//
// Deprecated: use dsn parameter in Open method
func WithEndpoint(endpoint string) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithEndpoint(endpoint))

		return nil
	}
}

// WithDatabase defines database option
//
// Warning: use ydb.Open with required Driver string parameter instead
//
// For making Driver string from endpoint+database+secure - use sugar.DSN()
//
// Deprecated: use dsn parameter in Open method
func WithDatabase(database string) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithDatabase(database))

		return nil
	}
}

// WithDisableSessionBalancer returns an Option that disables session balancing.
func WithDisableSessionBalancer() Option {
	return func(ctx context.Context, d *Driver) error {
		d.queryOptions = append(d.queryOptions, queryConfig.WithDisableSessionBalancer())
		d.tableOptions = append(d.tableOptions, tableConfig.WithDisableSessionBalancer())

		// implicit disable session balancer for SQL driver;
		// rule of thumb: if user disables session balancer anywhere, we disable it everywhere
		d.databaseSQLOptions = append(d.databaseSQLOptions, WithDisableServerBalancer())

		return nil
	}
}

// WithSecure defines secure option
//
// Warning: use ydb.Open with required Driver string parameter instead
//
// For making Driver string from endpoint+database+secure - use sugar.DSN()
func WithSecure(secure bool) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithSecure(secure))

		return nil
	}
}

// WithInsecure defines secure option.
//
// Warning: WithInsecure lost current TLS config.
func WithInsecure() Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithSecure(false))

		return nil
	}
}

// WithMinTLSVersion set minimum TLS version acceptable for connections
func WithMinTLSVersion(minVersion uint16) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithMinTLSVersion(minVersion))

		return nil
	}
}

// WithTLSSInsecureSkipVerify applies InsecureSkipVerify flag to TLS config
func WithTLSSInsecureSkipVerify() Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithTLSSInsecureSkipVerify())

		return nil
	}
}

// WithLogger add enables logging for selected tracing events.
//
// See trace package documentation for details.
func WithLogger(l log.Logger, details trace.Detailer, opts ...log.Option) Option {
	return func(ctx context.Context, d *Driver) error {
		d.logger = l
		d.loggerOpts = opts
		d.loggerDetails = details

		return nil
	}
}

// WithAnonymousCredentials force to make requests withou authentication.
func WithAnonymousCredentials() Option {
	return WithCredentials(
		credentials.NewAnonymousCredentials(credentials.WithSourceInfo("ydb.WithAnonymousCredentials()")),
	)
}

// WithCreateCredentialsFunc add callback funcion to provide requests credentials
func WithCreateCredentialsFunc(createCredentials func(ctx context.Context) (credentials.Credentials, error)) Option {
	return func(ctx context.Context, d *Driver) error {
		creds, err := createCredentials(ctx)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		d.options = append(d.options, config.WithCredentials(creds))

		return nil
	}
}

// WithCredentials in conjunction with Driver.With function prohibit reuse of conn pool.
// Thus, Driver.With will effectively create totally separate Driver.
func WithCredentials(c credentials.Credentials) Option {
	return WithCreateCredentialsFunc(func(context.Context) (credentials.Credentials, error) {
		return c, nil
	})
}

func WithBalancer(balancer *balancerConfig.Config) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithBalancer(balancer))

		return nil
	}
}

// WithDialTimeout sets timeout for establishing new Driver to cluster
//
// Default dial timeout is config.DefaultDialTimeout
func WithDialTimeout(timeout time.Duration) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithDialTimeout(timeout))

		return nil
	}
}

// WithGrpcMaxMessageSize set max size of message on grpc level
// use the option for the driver known custom limit.
// Driver can't read the limit from direct grpc option
func WithGrpcMaxMessageSize(sizeBytes int) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithGrpcMaxMessageSize(sizeBytes))

		return nil
	}
}

// With collects additional configuration options.
//
// This option does not replace collected option, instead it will append provided options.
func With(options ...config.Option) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, options...)

		return nil
	}
}

// MergeOptions concatentaes provided options to one cumulative value.
func MergeOptions(opts ...Option) Option {
	return func(ctx context.Context, d *Driver) error {
		for _, opt := range opts {
			if opt != nil {
				if err := opt(ctx, d); err != nil {
					return xerrors.WithStackTrace(err)
				}
			}
		}

		return nil
	}
}

// WithDiscoveryInterval sets interval between cluster discovery calls.
func WithDiscoveryInterval(discoveryInterval time.Duration) Option {
	return func(ctx context.Context, d *Driver) error {
		d.discoveryOptions = append(d.discoveryOptions, discoveryConfig.WithInterval(discoveryInterval))

		return nil
	}
}

// WithRetryBudget sets retry budget for all calls of all retryers.
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithRetryBudget(b budget.Budget) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithRetryBudget(b))

		return nil
	}
}

// WithTraceDriver appends trace.Driver into driver traces
func WithTraceDriver(t trace.Driver, opts ...trace.DriverComposeOption) Option { //nolint:gocritic
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithTrace(t, opts...))

		return nil
	}
}

// WithTraceRetry appends trace.Retry into retry traces
func WithTraceRetry(t trace.Retry, opts ...trace.RetryComposeOption) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options,
			config.WithTraceRetry(&t, append(
				[]trace.RetryComposeOption{
					trace.WithRetryPanicCallback(d.panicCallback),
				},
				opts...,
			)...),
		)

		return nil
	}
}

// WithCertificate appends certificate to TLS config root certificates
func WithCertificate(cert *x509.Certificate) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithCertificate(cert))

		return nil
	}
}

// WithCertificatesFromFile appends certificates by filepath to TLS config root certificates
func WithCertificatesFromFile(caFile string, opts ...certificates.FromFileOption) Option {
	if len(caFile) > 0 && caFile[0] == '~' {
		if home, err := os.UserHomeDir(); err == nil {
			caFile = filepath.Join(home, caFile[1:])
		}
	}
	if file, err := filepath.Abs(caFile); err == nil {
		caFile = file
	}
	if file, err := filepath.EvalSymlinks(caFile); err == nil {
		caFile = file
	}

	return func(ctx context.Context, d *Driver) error {
		certs, err := certificates.FromFile(caFile, opts...)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		for _, cert := range certs {
			if err := WithCertificate(cert)(ctx, d); err != nil {
				return xerrors.WithStackTrace(err)
			}
		}

		return nil
	}
}

// WithTLSConfig replaces older TLS config
//
// Warning: all early TLS config changes (such as WithCertificate, WithCertificatesFromFile, WithCertificatesFromPem,
// WithMinTLSVersion, WithTLSSInsecureSkipVerify) will be lost
func WithTLSConfig(tlsConfig *tls.Config) Option {
	return func(ctx context.Context, d *Driver) error {
		d.options = append(d.options, config.WithTLSConfig(tlsConfig))

		return nil
	}
}

// WithCertificatesFromPem appends certificates from pem-encoded data to TLS config root certificates
func WithCertificatesFromPem(bytes []byte, opts ...certificates.FromPemOption) Option {
	return func(ctx context.Context, d *Driver) error {
		certs, err := certificates.FromPem(bytes, opts...)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}
		for _, cert := range certs {
			_ = WithCertificate(cert)(ctx, d)
		}

		return nil
	}
}

// WithTableConfigOption collects additional configuration options for table.Client.
// This option does not replace collected option, instead it will appen provided options.
func WithTableConfigOption(option tableConfig.Option) Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, option)

		return nil
	}
}

// WithQueryConfigOption collects additional configuration options for query.Client.
// This option does not replace collected option, instead it will appen provided options.
func WithQueryConfigOption(option queryConfig.Option) Option {
	return func(ctx context.Context, d *Driver) error {
		d.queryOptions = append(d.queryOptions, option)

		return nil
	}
}

// WithSessionPoolSizeLimit set max size of internal sessions pool in table.Client
func WithSessionPoolSizeLimit(sizeLimit int) Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, tableConfig.WithSizeLimit(sizeLimit))
		d.queryOptions = append(d.queryOptions, queryConfig.WithPoolLimit(sizeLimit))

		return nil
	}
}

// WithSessionPoolSessionUsageLimit set pool session max usage:
// - if argument type is uint64 - WithSessionPoolSessionUsageLimit limits max usage count of pool session
// - if argument type is time.Duration - WithSessionPoolSessionUsageLimit limits max time to live of pool session
func WithSessionPoolSessionUsageLimit[T interface{ uint64 | time.Duration }](limit T) Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, tableConfig.WithSessionPoolSessionUsageLimit(limit))
		d.queryOptions = append(d.queryOptions, queryConfig.WithSessionPoolSessionUsageLimit(limit))

		return nil
	}
}

// WithLazyTx enables lazy transactions in query service client
//
// Lazy transaction means that begin call will be noop and first execute creates interactive transaction with given
// transaction settings
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func WithLazyTx(lazyTx bool) Option {
	return func(ctx context.Context, d *Driver) error {
		d.queryOptions = append(d.queryOptions, queryConfig.WithLazyTx(lazyTx))

		return nil
	}
}

// WithSessionPoolIdleThreshold defines interval for idle sessions
func WithSessionPoolIdleThreshold(idleThreshold time.Duration) Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, tableConfig.WithIdleThreshold(idleThreshold))
		d.databaseSQLOptions = append(d.databaseSQLOptions,
			xsql.WithIdleThreshold(idleThreshold),
		)

		return nil
	}
}

// WithExecuteDataQueryOverQueryClient option enables execution data queries from legacy table service client over
// query client API. Using this option you can execute queries from legacy table service client through
// `table.Session.Execute` using internal query client API without limitation of 1000 rows in response.
// Be careful: an OOM problem may happen because bigger result requires more memory
func WithExecuteDataQueryOverQueryClient(b bool) Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, tableConfig.ExecuteDataQueryOverQueryService(b))

		return nil
	}
}

// WithSessionPoolSessionIdleTimeToLive limits maximum time to live of idle session
// If idleTimeToLive is less than or equal to zero then sessions will not be closed by idle
func WithSessionPoolSessionIdleTimeToLive(idleThreshold time.Duration) Option {
	return func(ctx context.Context, d *Driver) error {
		d.queryOptions = append(d.queryOptions, queryConfig.WithSessionIdleTimeToLive(idleThreshold))

		return nil
	}
}

// WithSessionPoolCreateSessionTimeout set timeout for new session creation process in table.Client
func WithSessionPoolCreateSessionTimeout(createSessionTimeout time.Duration) Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, tableConfig.WithCreateSessionTimeout(createSessionTimeout))
		d.queryOptions = append(d.queryOptions, queryConfig.WithSessionCreateTimeout(createSessionTimeout))

		return nil
	}
}

// WithSessionPoolDeleteTimeout set timeout to gracefully close deleting session in table.Client
func WithSessionPoolDeleteTimeout(deleteTimeout time.Duration) Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, tableConfig.WithDeleteTimeout(deleteTimeout))
		d.queryOptions = append(d.queryOptions, queryConfig.WithSessionDeleteTimeout(deleteTimeout))

		return nil
	}
}

// WithSessionPoolKeepAliveMinSize set minimum sessions should be keeped alive in table.Client
//
// Deprecated: use WithApplicationName instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithSessionPoolKeepAliveMinSize(keepAliveMinSize int) Option {
	return func(ctx context.Context, d *Driver) error { return nil }
}

// WithSessionPoolKeepAliveTimeout set timeout of keep alive requests for session in table.Client
//
// Deprecated: use WithApplicationName instead.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func WithSessionPoolKeepAliveTimeout(keepAliveTimeout time.Duration) Option {
	return func(ctx context.Context, d *Driver) error { return nil }
}

// WithIgnoreTruncated disables errors on truncated flag
func WithIgnoreTruncated() Option {
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(d.tableOptions, tableConfig.WithIgnoreTruncated())

		return nil
	}
}

// WithPanicCallback specified behavior on panic
// Warning: WithPanicCallback must be defined on start of all options
// (before `WithTrace{Driver,Table,Scheme,Scripting,Coordination,Ratelimiter}` and other options)
// If not defined - panic would not intercept with driver
func WithPanicCallback(panicCallback func(e interface{})) Option {
	return func(ctx context.Context, d *Driver) error {
		d.panicCallback = panicCallback
		d.options = append(d.options, config.WithPanicCallback(panicCallback))

		return nil
	}
}

// WithSharedBalancer sets balancer from parent driver to child driver
func WithSharedBalancer(parent *Driver) Option {
	return func(ctx context.Context, d *Driver) error {
		d.metaBalancer.balancer = parent.metaBalancer.balancer
		d.metaBalancer.close = func(ctx context.Context) error { return nil }

		return nil
	}
}

// WithTraceTable appends trace.Table into table traces
func WithTraceTable(t trace.Table, opts ...trace.TableComposeOption) Option { //nolint:gocritic
	return func(ctx context.Context, d *Driver) error {
		d.tableOptions = append(
			d.tableOptions,
			tableConfig.WithTrace(
				&t,
				append(
					[]trace.TableComposeOption{
						trace.WithTablePanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithTraceQuery appends trace.Query into query traces
func WithTraceQuery(t trace.Query, opts ...trace.QueryComposeOption) Option { //nolint:gocritic
	return func(ctx context.Context, d *Driver) error {
		d.queryOptions = append(d.queryOptions,
			queryConfig.WithTrace(&t,
				append(
					[]trace.QueryComposeOption{
						trace.WithQueryPanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithTraceScripting scripting trace option
func WithTraceScripting(t trace.Scripting, opts ...trace.ScriptingComposeOption) Option {
	return func(ctx context.Context, d *Driver) error {
		d.scriptingOptions = append(d.scriptingOptions,
			scriptingConfig.WithTrace(
				t,
				append(
					[]trace.ScriptingComposeOption{
						trace.WithScriptingPanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithTraceScheme returns scheme trace option
func WithTraceScheme(t trace.Scheme, opts ...trace.SchemeComposeOption) Option {
	return func(ctx context.Context, d *Driver) error {
		d.schemeOptions = append(d.schemeOptions,
			schemeConfig.WithTrace(
				t,
				append(
					[]trace.SchemeComposeOption{
						trace.WithSchemePanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithTraceCoordination returns coordination trace option
func WithTraceCoordination(t trace.Coordination, opts ...trace.CoordinationComposeOption) Option { //nolint:gocritic
	return func(ctx context.Context, d *Driver) error {
		d.coordinationOptions = append(d.coordinationOptions,
			coordinationConfig.WithTrace(
				&t,
				append(
					[]trace.CoordinationComposeOption{
						trace.WithCoordinationPanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithTraceRatelimiter returns ratelimiter trace option
func WithTraceRatelimiter(t trace.Ratelimiter, opts ...trace.RatelimiterComposeOption) Option {
	return func(ctx context.Context, d *Driver) error {
		d.ratelimiterOptions = append(d.ratelimiterOptions,
			ratelimiterConfig.WithTrace(
				t,
				append(
					[]trace.RatelimiterComposeOption{
						trace.WithRatelimiterPanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithRatelimiterOptions returns reatelimiter option
func WithRatelimiterOptions(opts ...ratelimiterConfig.Option) Option {
	return func(ctx context.Context, d *Driver) error {
		d.ratelimiterOptions = append(d.ratelimiterOptions, opts...)

		return nil
	}
}

// WithTraceDiscovery adds configured discovery tracer to Driver
func WithTraceDiscovery(t trace.Discovery, opts ...trace.DiscoveryComposeOption) Option {
	return func(ctx context.Context, d *Driver) error {
		d.discoveryOptions = append(d.discoveryOptions,
			discoveryConfig.WithTrace(
				t,
				append(
					[]trace.DiscoveryComposeOption{
						trace.WithDiscoveryPanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithTraceTopic adds configured discovery tracer to Driver
func WithTraceTopic(t trace.Topic, opts ...trace.TopicComposeOption) Option { //nolint:gocritic
	return func(ctx context.Context, d *Driver) error {
		d.topicOptions = append(d.topicOptions,
			topicoptions.WithTrace(
				t,
				append(
					[]trace.TopicComposeOption{
						trace.WithTopicPanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// WithTraceDatabaseSQL adds configured discovery tracer to Driver
func WithTraceDatabaseSQL(t trace.DatabaseSQL, opts ...trace.DatabaseSQLComposeOption) Option { //nolint:gocritic
	return func(ctx context.Context, d *Driver) error {
		d.databaseSQLOptions = append(d.databaseSQLOptions,
			xsql.WithTrace(
				&t,
				append(
					[]trace.DatabaseSQLComposeOption{
						trace.WithDatabaseSQLPanicCallback(d.panicCallback),
					},
					opts...,
				)...,
			),
		)

		return nil
	}
}

// Private technical options for correct copies processing

func withOnClose(onClose func(c *Driver)) Option {
	return func(ctx context.Context, d *Driver) error {
		d.onClose = append(d.onClose, onClose)

		return nil
	}
}

func withConnPool(pool *conn.Pool) Option {
	return func(ctx context.Context, d *Driver) error {
		d.pool = pool

		return pool.Take(ctx)
	}
}
