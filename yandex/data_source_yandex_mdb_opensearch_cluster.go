package yandex

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBOpenSearchCluster() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceYandexMDBOpenSearchClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"environment": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"config": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"admin_password": {
							Type:      schema.TypeString,
							Optional:  true,
							Computed:  true,
							Sensitive: true,
						},

						"opensearch": {
							Type:     schema.TypeList,
							Required: true,
							//Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"node_groups": {
										Type: schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"resources": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 1,
													MinItems: 1,
													Elem:     openSearchResourcesSchema(),
												},

												"hosts_count": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"zone_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Optional: true,
													Computed: true,
												},

												"subnet_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Optional: true,
													Computed: true,
												},

												"assign_public_ip": {
													Type:     schema.TypeBool,
													Computed: true,
													Optional: true,
												},

												"roles": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      openSearchRoleHash,
													Optional: true,
													Computed: true,
												},
											},
										},
										Set: openSearchNodeGroupDeepHash,
										//Computed: true,
										Required: true,
									},

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

						"dashboards": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"node_groups": {
										Type: schema.TypeSet,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"name": {
													Type:     schema.TypeString,
													Required: true,
												},

												"resources": {
													Type:     schema.TypeSet,
													Required: true,
													MaxItems: 1,
													Elem:     openSearchResourcesSchema(),
												},

												"hosts_count": {
													Type:     schema.TypeInt,
													Required: true,
												},

												"zone_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Optional: true,
													Computed: true,
												},

												"subnet_ids": {
													Type:     schema.TypeSet,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
													Optional: true,
													Computed: true,
												},

												"assign_public_ip": {
													Type:     schema.TypeBool,
													Computed: true,
													Optional: true,
												},
											},
										},
										//Set: dashboardsNodeGroupDeepHash,

										//Set:      dashboardsNodeGroupNameHash,
										//Optional: false,
										//Computed: true,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"hosts": {
				Type:     schema.TypeSet,
				Computed: true,
				Set:      opensearchHostFQDNHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"roles": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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

			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
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

func dataSourceYandexMDBOpenSearchClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.OpenSearchClusterResolver)
		if err != nil {
			return diag.Errorf("failed to resolve data source Opensearch Cluster by name: %v", err)
		}

		d.Set("cluster_id", clusterID)
	}

	cluster, err := config.sdk.MDB().OpenSearch().Cluster().Get(ctx, &opensearch.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string))))
	}

	mw := flattenOpenSearchMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(clusterID)
	if err := resourceYandexMDBOpenSearchClusterReadEx(ctx, d, meta, "DataSource"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
