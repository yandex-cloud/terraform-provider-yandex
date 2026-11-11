package sweep

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	ydbsdk "github.com/yandex-cloud/go-sdk/services/ydb/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_ydb "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/ydb/pluginsdk"
)

func registerResourceYandexYdbDatabaseDedicated(add func(string, *resource.Sweeper)) {
	add("yandex_ydb_database_dedicated", &resource.Sweeper{
		Name: "yandex_ydb_database_dedicated",
		F:    sweepAllYDBDatabaseDedicated,
	})
}

func sweepAllYDBDatabaseDedicated(_ string) error {
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
		// skip serverless, it's swept in a separate sweeper
		if _, ok := c.DatabaseType.(*ydb.Database_ServerlessDatabase); !ok {
			if !sweepYDBDatabaseDedicated(conf, c.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep YDB dedicated database %q", c.Id))
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepYDBDatabaseDedicated(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepYDBDatabaseDedicatedOnce, conf, "YDB dedicated database", id)
}

func sweepYDBDatabaseDedicatedOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_ydb.YandexYDBDedicatedDefaultTimeout)
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

func checkAndUnsetYDBDeletionProtection(
	conf *providerconfig.Config,
	ctx context.Context,
	id string,
) error {
	client := ydbsdk.NewDatabaseClient(conf.SDK)

	db, err := client.Get(ctx, &ydb.GetDatabaseRequest{
		DatabaseId: id,
	})
	if err != nil {
		return fmt.Errorf("failed to get database with id %q: %s", id, err)
	}

	if !db.DeletionProtection {
		return nil
	}

	return unsetYDBDatabaseDeletionProtection(conf, ctx, db)
}

func unsetYDBDatabaseDeletionProtection(
	conf *providerconfig.Config,
	ctx context.Context,
	db *ydb.Database,
) error {
	req := &ydb.UpdateDatabaseRequest{
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{
				"deletion_protection",
			},
		},
		DatabaseId: db.Id,
	}
	req.DeletionProtection = false

	client := ydbsdk.NewDatabaseClient(conf.SDK)

	op, err := client.Update(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to unset deletion_protection for database %q: %s", db.Id, err)
	}

	_, err = op.Wait(ctx)
	return err
}
