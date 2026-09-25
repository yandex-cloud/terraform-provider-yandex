package dsn

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	insecureSchema = "grpc"
	secureSchema   = "grpcs"
	reScheme       = regexp.MustCompile(`^\w+://`)
	databaseParam  = "database"
)

type UserInfo struct {
	User     string
	Password string
}

type parsedInfo struct {
	UserInfo *UserInfo
	Options  []config.Option
	Params   url.Values
}

func Parse(dsn string) (info parsedInfo, err error) {
	uri, err := url.ParseRequestURI(func() string {
		if reScheme.MatchString(dsn) {
			return dsn
		}

		return secureSchema + "://" + dsn
	}())
	if err != nil {
		return info, xerrors.WithStackTrace(err)
	}
	if port := uri.Port(); port == "" {
		return info, xerrors.WithStackTrace(fmt.Errorf("bad connection string '%s': port required", dsn))
	}
	info.Options = append(info.Options,
		config.WithSecure(uri.Scheme != insecureSchema),
		config.WithEndpoint(uri.Host),
	)
	if uri.Path != "" {
		info.Options = append(info.Options, config.WithDatabase(uri.Path))
	}
	if uri.User != nil {
		password, _ := uri.User.Password()
		info.UserInfo = &UserInfo{
			User:     uri.User.Username(),
			Password: password,
		}
	}
	info.Params = uri.Query()
	if database, has := info.Params[databaseParam]; has && len(database) > 0 {
		info.Options = append(info.Options,
			config.WithDatabase(database[0]),
		)
		delete(info.Params, databaseParam)
	}

	return info, nil
}
