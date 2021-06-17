package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBElasticsearchCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBElasticsearchClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
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

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"environment": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"config": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
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
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resources": {
										Type:     schema.TypeList,
										MaxItems: 1,
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
										MaxItems: 1,
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
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
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

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"host": {
				Type:     schema.TypeSet,
				MinItems: 1,
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

	d.SetId(clusterID)
	return resourceYandexMDBElasticsearchClusterRead(d, meta)
}
