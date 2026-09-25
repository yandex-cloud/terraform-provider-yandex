package table

import (
	"context"

	ydb "github.com/ydb-platform/ydb-go-sdk/v3"

	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
)

type ClientParams struct {
	DatabaseEndpoint string
	AuthCreds        auth.YdbCredentials
}

func CreateDBConnection(ctx context.Context, params ClientParams) (*ydb.Driver, error) {
	var opts []ydb.Option
	switch {
	case params.AuthCreds.Token != "":
		opts = append(opts, ydb.WithAccessTokenCredentials(params.AuthCreds.Token))
	case params.AuthCreds.User != "":
		opts = append(opts, ydb.WithStaticCredentials(params.AuthCreds.User, params.AuthCreds.Password))
	}

	db, err := ydb.Open(ctx, params.DatabaseEndpoint, opts...)
	if err != nil {
		return nil, err
	}
	return db, nil
}
