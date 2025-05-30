package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexMDBGreenplumCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed Greenplum cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-greenplum/).\n\n~> Either `cluster_id` or `name` should be specified.\n",

		Read: dataSourceYandexMDBGreenplumClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Greenplum cluster.",
				Computed:    true,
				Optional:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Computed:    true,
				Optional:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
			},
			"environment": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBGreenplumCluster().Schema["environment"].Description,
				Computed:    true,
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Computed:    true,
			},
			"zone": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
			},
			"subnet_id": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBGreenplumCluster().Schema["subnet_id"].Description,
				Computed:    true,
			},
			"assign_public_ip": {
				Type:        schema.TypeBool,
				Description: resourceYandexMDBGreenplumCluster().Schema["assign_public_ip"].Description,
				Computed:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBGreenplumCluster().Schema["version"].Description,
				Computed:    true,
			},
			"master_host_count": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBGreenplumCluster().Schema["master_host_count"].Description,
				Computed:    true,
			},
			"segment_host_count": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBGreenplumCluster().Schema["segment_host_count"].Description,
				Computed:    true,
			},
			"segment_in_host": {
				Type:        schema.TypeInt,
				Description: resourceYandexMDBGreenplumCluster().Schema["segment_in_host"].Description,
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
			"service_account_id": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBGreenplumCluster().Schema["service_account_id"].Description,
				Computed:    true,
			},
			"logging": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"log_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"folder_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"command_center_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"greenplum_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"pooler_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"master_subcluster": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"segment_subcluster": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"master_hosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"segment_hosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"user_name": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBGreenplumCluster().Schema["user_name"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBGreenplumCluster().Schema["health"].Description,
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBGreenplumCluster().Schema["status"].Description,
				Computed:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Computed:    true,
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"day": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"hour": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
			},

			"backup_window_start": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"access": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_lens": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"web_sql": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"data_transfer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"yandex_query": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"pooler_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pooling_mode": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"pool_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"pool_client_idle_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"greenplum_config": {
				Type:             schema.TypeMap,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbGreenplumSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbGreenplumSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cloud_storage": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"master_host_group_ids": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBGreenplumCluster().Schema["master_host_group_ids"].Description,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"segment_host_group_ids": {
				Type:        schema.TypeSet,
				Description: resourceYandexMDBGreenplumCluster().Schema["segment_host_group_ids"].Description,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},
			"pxf_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connection_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"upload_timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_threads": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"pool_allow_core_thread_timeout": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"pool_core_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"pool_queue_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"pool_max_size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"xmx": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"xms": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"background_activities": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"analyze_and_vacuum": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_time": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"analyze_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"vacuum_timeout": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"query_killer_idle": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"max_age": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"ignore_users": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"query_killer_idle_in_transaction": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"max_age": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"ignore_users": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"query_killer_long_running": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enable": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"max_age": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"ignore_users": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexMDBGreenplumClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.GreenplumClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Greenplum Cluster by name: %v", err)
		}

		d.Set("cluster_id", clusterID)
	}

	cluster, err := config.sdk.MDB().Greenplum().Cluster().Get(ctx, &greenplum.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	d.SetId(cluster.Id)
	return resourceYandexMDBGreenplumClusterRead(d, meta)
}
