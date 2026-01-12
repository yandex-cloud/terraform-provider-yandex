package yandex

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

// ==============================================================================
//                                 DATABASE
// ==============================================================================

func listPGDatabases(ctx context.Context, config *Config, clusterId string) ([]*postgresql.Database, error) {
	databases := []*postgresql.Database{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().PostgreSQL().Database().List(ctx, &postgresql.ListDatabasesRequest{
			ClusterId: clusterId,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of databases for PostgreSQL Cluster '%q': %s", clusterId, err)
		}

		databases = append(databases, resp.Databases...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return databases, nil
}
