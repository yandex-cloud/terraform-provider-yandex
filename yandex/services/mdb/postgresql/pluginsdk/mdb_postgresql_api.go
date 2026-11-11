package pluginsdk

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	postgresqlsdk "github.com/yandex-cloud/go-sdk/services/mdb/postgresql/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func listPGDatabases(ctx context.Context, config *providerconfig.Config, clusterId string) ([]*postgresql.Database, error) {
	databases := []*postgresql.Database{}
	pageToken := ""

	for {
		resp, err := postgresqlsdk.NewDatabaseClient(config.SDK).List(ctx, &postgresql.ListDatabasesRequest{
			ClusterId: clusterId,
			PageSize:  sdkcommon.DefaultMDBPageSize,
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
