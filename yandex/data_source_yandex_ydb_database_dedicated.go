package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexYDBDatabaseDedicated() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Database (dedicated) cluster. For more information, see [the official documentation](https://yandex.cloud/docs/ydb/concepts/serverless_and_dedicated).\n\n~> If `database_id` is not specified `name` and `folder_id` will be used to designate Yandex Database cluster.\n",

		Read: dataSourceYandexYDBDatabaseDedicatedRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"database_id": {
				Type:        schema.TypeString,
				Description: "ID of the Yandex Database cluster.",
				Optional:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Computed:    true,
			},

			"subnet_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["subnet_ids"],
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"resource_preset_id": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseDedicated().Schema["resource_preset_id"].Description,
				Computed:    true,
			},

			"scale_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_scale": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"storage_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_type_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"group_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"location": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"zone": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"location_id": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseDedicated().Schema["location_id"].Description,
				Computed:    true,
			},

			"assign_public_ips": {
				Type:        schema.TypeBool,
				Description: resourceYandexYDBDatabaseDedicated().Schema["assign_public_ips"].Description,
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
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

			"ydb_full_endpoint": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseDedicated().Schema["ydb_full_endpoint"].Description,
				Computed:    true,
			},

			"ydb_api_endpoint": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseDedicated().Schema["ydb_api_endpoint"].Description,
				Computed:    true,
			},

			"database_path": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseDedicated().Schema["database_path"].Description,
				Computed:    true,
			},

			"tls_enabled": {
				Type:        schema.TypeBool,
				Description: resourceYandexYDBDatabaseDedicated().Schema["tls_enabled"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexYDBDatabaseDedicated().Schema["status"].Description,
				Computed:    true,
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
		},
	}
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
