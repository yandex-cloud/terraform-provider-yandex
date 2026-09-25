package ydb

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/ydb-platform/ydb-go-sdk/v3/balancers"
	"github.com/ydb-platform/ydb-go-sdk/v3/credentials"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/bind"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/dsn"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql"
)

const tablePathPrefixTransformer = "table_path_prefix"

var dsnParsers = []func(dsn string) (opts []Option, _ error){
	func(dsn string) ([]Option, error) {
		opts, err := parseConnectionString(dsn)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		return opts, nil
	},
}

// RegisterDsnParser registers DSN parser for ydb.Open and sql.Open driver constructors
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func RegisterDsnParser(parser func(dsn string) (opts []Option, _ error)) (registrationID int) {
	dsnParsers = append(dsnParsers, parser)

	return len(dsnParsers) - 1
}

// UnregisterDsnParser unregisters DSN parser by key
//
// Experimental: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#experimental
func UnregisterDsnParser(registrationID int) {
	dsnParsers[registrationID] = nil
}

var stringToType = map[string]QueryMode{
	"data":      DataQueryMode,
	"scan":      ScanQueryMode,
	"scheme":    SchemeQueryMode,
	"scripting": ScriptingQueryMode,
	"query":     QueryExecuteQueryMode,
}

func queryModeFromString(s string) QueryMode {
	if t, ok := stringToType[s]; ok {
		return t
	}

	return unknownQueryMode
}

//nolint:funlen
func parseConnectionString(dataSourceName string) (opts []Option, _ error) {
	info, err := dsn.Parse(dataSourceName)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}
	opts = append(opts, With(info.Options...))
	if token := info.Params.Get("token"); token != "" {
		opts = append(opts, WithCredentials(credentials.NewAccessTokenCredentials(token)))
	}
	if balancer := info.Params.Get("go_balancer"); balancer != "" {
		opts = append(opts, WithBalancer(balancers.FromConfig(balancer)))
	} else if balancer := info.Params.Get("balancer"); balancer != "" {
		opts = append(opts, WithBalancer(balancers.FromConfig(balancer)))
	}
	if queryMode := info.Params.Get("go_query_mode"); queryMode != "" {
		switch mode := queryModeFromString(queryMode); mode {
		case QueryExecuteQueryMode:
			opts = append(opts, withConnectorOptions(xsql.WithQueryService(true)))
		case unknownQueryMode:
			return nil, xerrors.WithStackTrace(fmt.Errorf("unknown query mode: %s", queryMode))
		default:
			opts = append(opts, withConnectorOptions(xsql.WithDefaultQueryMode(modeToMode(mode))))
		}
	} else if queryMode := info.Params.Get("query_mode"); queryMode != "" {
		switch mode := queryModeFromString(queryMode); mode {
		case QueryExecuteQueryMode:
			opts = append(opts, withConnectorOptions(xsql.WithQueryService(true)))
		case unknownQueryMode:
			return nil, xerrors.WithStackTrace(fmt.Errorf("unknown query mode: %s", queryMode))
		default:
			opts = append(opts, withConnectorOptions(xsql.WithDefaultQueryMode(modeToMode(mode))))
		}
	}
	if fakeTx := info.Params.Get("go_fake_tx"); fakeTx != "" {
		for _, queryMode := range strings.Split(fakeTx, ",") {
			switch mode := queryModeFromString(queryMode); mode {
			case unknownQueryMode:
				return nil, xerrors.WithStackTrace(fmt.Errorf("unknown query mode: %s", queryMode))
			default:
				opts = append(opts, withConnectorOptions(WithFakeTx(mode)))
			}
		}
	}
	if info.Params.Has("go_query_bind") {
		var binders []xsql.Option
		queryTransformers := strings.Split(info.Params.Get("go_query_bind"), ",")
		for _, transformer := range queryTransformers {
			switch transformer {
			case "declare":
				binders = append(binders, xsql.WithQueryBind(bind.AutoDeclare{}))
			case "positional":
				binders = append(binders, xsql.WithQueryBind(bind.PositionalArgs{}))
			case "numeric":
				binders = append(binders, xsql.WithQueryBind(bind.NumericArgs{}))
			case "wide_time_types":
				binders = append(binders, xsql.WithQueryBind(bind.WideTimeTypes(true)))
			default:
				if strings.HasPrefix(transformer, tablePathPrefixTransformer) {
					prefix, err := extractTablePathPrefixFromBinderName(transformer)
					if err != nil {
						return nil, xerrors.WithStackTrace(err)
					}
					binders = append(binders, xsql.WithQueryBind(bind.TablePathPrefix(prefix)))
				} else {
					return nil, xerrors.WithStackTrace(
						fmt.Errorf("unknown query rewriter: %s", transformer),
					)
				}
			}
		}
		opts = append(opts, withConnectorOptions(binders...))
	}

	return opts, nil
}

var (
	tablePathPrefixRe       = regexp.MustCompile(tablePathPrefixTransformer + "\\((.*)\\)")
	errWrongTablePathPrefix = errors.New("wrong '" + tablePathPrefixTransformer + "' query transformer")
)

func extractTablePathPrefixFromBinderName(binderName string) (string, error) {
	ss := tablePathPrefixRe.FindAllStringSubmatch(binderName, -1)
	if len(ss) != 1 || len(ss[0]) != 2 || ss[0][1] == "" {
		return "", xerrors.WithStackTrace(fmt.Errorf("%w: %s", errWrongTablePathPrefix, binderName))
	}

	return ss[0][1], nil
}
