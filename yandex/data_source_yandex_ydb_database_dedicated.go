package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexYDBDatabaseDedicated() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexYDBDatabaseDedicatedRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"database_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"subnet_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"resource_preset_id": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:     schema.TypeString,
				Computed: true,
			},

			"assign_public_ips": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"ydb_full_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ydb_api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_path": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
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
