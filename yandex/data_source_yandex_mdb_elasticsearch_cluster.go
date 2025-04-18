package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/elasticsearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBElasticsearchCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Managed Elasticsearch cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-elasticsearch/concepts).\n\n~> Either `cluster_id` or `name` should be specified.\n",
		Read:        dataSourceYandexMDBElasticsearchClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Elasticsearch cluster.",
				Computed:    true,
				Optional:    true,
			},

			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Computed:    true,
				Optional:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Computed:    true,
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Computed:    true,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
				Computed:    true,
			},

			"environment": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBElasticsearchCluster().Schema["environment"].Description,
				Computed:    true,
			},

			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Computed:    true,
			},

			"service_account_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["service_account_id"],
				Optional:    true,
				Computed:    true,
			},

			"config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"edition": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"admin_password": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"data_node": {
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
												"disk_size": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"disk_type_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},

						"master_node": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
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
												"disk_size": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"disk_type_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						}, // masternode

						"plugins": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},

			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
			},

			"health": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBElasticsearchCluster().Schema["health"].Description,
				Computed:    true,
			},

			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexMDBElasticsearchCluster().Schema["status"].Description,
				Computed:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"host": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      elasticsearchHostFQDNHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Computed:    true,
				Optional:    true,
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
		},
	}
}

func dataSourceYandexMDBElasticsearchClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.ElasticSearchClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Elasticsearch Cluster by name: %v", err)
		}

		d.Set("cluster_id", clusterID)
	}

	cluster, err := config.sdk.MDB().ElasticSearch().Cluster().Get(ctx, &elasticsearch.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	mw := flattenElasticsearchMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	d.SetId(clusterID)
	return resourceYandexMDBElasticsearchClusterRead(d, meta)
}
