package yandex

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBMongodbCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed MongoDB cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts).\n\n~> Either `cluster_id` or `name` should be specified.\n",

		ReadContext: dataSourceYandexMDBMongodbClusterRead,
		Schema:      convertToOptional(resourceYandexMDBMongodbCluster().Schema),
	}
}

func convertToOptional(originalSchema map[string]*schema.Schema) map[string]*schema.Schema {
	optionalSchema := map[string]*schema.Schema{}
	for key, value := range originalSchema {
		newItem := *value
		newItem.Required = false
		newItem.ExactlyOneOf = []string{}
		newItem.Optional = true
		newItem.ForceNew = false

		switch newItem.Type {
		case schema.TypeList, schema.TypeSet:
			switch newItem.Elem.(type) {
			case *schema.Resource:
				elem := *newItem.Elem.(*schema.Resource)
				elem.Schema = convertToOptional(elem.Schema)
				newItem.Elem = &elem
			}
		}

		optionalSchema[key] = &newItem
	}
	return optionalSchema
}

func dataSourceYandexMDBMongodbClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// fix to import users and databases
	ctx = context.WithValue(ctx, ReadModeKey, true)
	config := meta.(*Config)

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.MongoDBClusterResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Mongodb Cluster by name: %v", err)
		}

		if err := d.Set("cluster_id", clusterID); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(clusterID)

	return resourceYandexMDBMongodbClusterRead(ctx, d, meta)
}
