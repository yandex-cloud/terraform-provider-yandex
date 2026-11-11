package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	ydbsdk "github.com/yandex-cloud/go-sdk/services/ydb/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_ydb "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/ydb/pluginsdk"
)

func registerResourceYandexYdbDatabaseServerless(add func(string, *resource.Sweeper)) {
	add("yandex_ydb_database_serverless", &resource.Sweeper{
		Name: "yandex_ydb_database_serverless",
		F:    sweepAllYDBDatabaseServerless,
	})
}

func sweepAllYDBDatabaseServerless(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := ydbsdk.NewDatabaseClient(conf.SDK)

	resp, err := client.List(conf.Context(), &ydb.ListDatabasesRequest{
		FolderId: conf.FolderID,
		PageSize: 1000,
	})
	if err != nil {
		return fmt.Errorf("error getting YDB databases: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Databases {
		// only serverless, other types are swept in a separate sweeper
		if _, ok := c.DatabaseType.(*ydb.Database_ServerlessDatabase); ok {
			if !sweepYDBServerlessDatabase(conf, c.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep YDB serverless database %q", c.Id))
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepYDBServerlessDatabase(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepYDBServerlessDatabaseOnce, conf, "YDB serverless database", id)
}

func sweepYDBServerlessDatabaseOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_ydb.YandexYDBServerlessDefaultTimeout)
	defer cancel()

	err := checkAndUnsetYDBDeletionProtection(conf, ctx, id)
	if err != nil {
		return err
	}

	client := ydbsdk.NewDatabaseClient(conf.SDK)

	op, err := client.Delete(ctx, &ydb.DeleteDatabaseRequest{
		DatabaseId: id,
	})
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
