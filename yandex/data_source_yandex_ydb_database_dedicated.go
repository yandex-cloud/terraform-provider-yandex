package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexYDBDatabaseDedicated() *schema.Resource {
	dataSource := convertResourceToDataSource(resourceYandexYDBDatabaseDedicated())

	dataSource.Read = dataSourceYandexYDBDatabaseDedicatedRead

	dataSource.Description = "Get information about a Yandex Database (dedicated) cluster. For more information, see [the official documentation](https://yandex.cloud/docs/ydb/concepts/serverless_and_dedicated).\n\n~> If `database_id` is not specified `name` and `folder_id` will be used to designate Yandex Database cluster.\n"

	dataSource.Schema["name"].Optional = true
	dataSource.Schema["name"].Computed = false

	dataSource.Schema["folder_id"].Optional = true
	dataSource.Schema["folder_id"].Computed = false

	dataSource.Schema["deletion_protection"].Optional = true

	dataSource.Schema["database_id"] = &schema.Schema{
		Type:        schema.TypeString,
		Description: "ID of the Yandex Database cluster.",
		Optional:    true,
	}

	delete(dataSource.Schema, "sleep_after")

	return dataSource
}

func dataSourceYandexYDBDatabaseDedicatedRead(d *schema.ResourceData, meta interface{}) error {
	database, err := dataSourceYandexYDBDatabaseRead(d, meta)
	if err != nil {
		return err
	}

	return flattenYandexYDBDatabaseDedicated(d, database)
}

func dataSourceYandexYDBDatabaseRead(d *schema.ResourceData, meta interface{}) (*ydb.Database, error) {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	err := checkOneOf(d, "database_id", "name")
	if err != nil {
		return nil, err
	}

	databaseID := d.Get("database_id").(string)
	_, tgNameOk := d.GetOk("name")

	if tgNameOk {
		databaseID, err = resolveObjectID(ctx, config, d, sdkresolvers.YDBDatabaseResolver)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve data source Yandex Database by name: %v", err)
		}
	}

	req := ydb.GetDatabaseRequest{
		DatabaseId: databaseID,
	}

	database, err := config.sdk.YDB().Database().Get(ctx, &req)
	if err != nil {
		return nil, handleNotFoundError(err, d, fmt.Sprintf("Yandex Database %q", d.Id()))
	}

	d.SetId(database.Id)
	d.Set("database_id", databaseID)

	return database, nil
}
