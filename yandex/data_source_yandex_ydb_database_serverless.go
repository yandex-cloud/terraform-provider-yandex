package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexYDBDatabaseServerless() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Database serverless cluster. For more information, see [the official documentation](https://yandex.cloud/docs/ydb/concepts/serverless_and_dedicated).\n\n~> If `database_id` is not specified `name` and `folder_id` will be used to designate Yandex Database serverless cluster.\n",

		Read: dataSourceYandexYDBDatabaseServerlessRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"database_id": {
				Type:        schema.TypeString,
				Description: "ID of the Yandex Database serverless cluster.",
				Optional:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
			},

			"location_id": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseServerless().Schema["location_id"].Description,
				Computed:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"document_api_endpoint": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseServerless().Schema["document_api_endpoint"].Description,
				Computed:    true,
			},

			"ydb_full_endpoint": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseServerless().Schema["ydb_full_endpoint"].Description,
				Computed:    true,
			},

			"ydb_api_endpoint": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseServerless().Schema["ydb_api_endpoint"].Description,
				Computed:    true,
			},

			"database_path": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseServerless().Schema["database_path"].Description,
				Computed:    true,
			},

			"tls_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexYDBDatabaseServerless().Schema["tls_enabled"].Description,
				Computed:    true,
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseServerless().Schema["status"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
			"serverless_database": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"throttling_rcu_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"storage_size_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"enable_throttling_rcu_limit": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"provisioned_rcu_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexYDBDatabaseServerlessRead(d *schema.ResourceData, meta interface{}) error {
	database, err := dataSourceYandexYDBDatabaseRead(d, meta)
	if err != nil {
		return err
	}

	return flattenYandexYDBDatabaseServerless(d, database)
}
