package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexKubernetesClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_ipv4_range": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_ipv6_range": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"node_ipv4_cidr_mask_size": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"service_ipv4_range": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_ipv6_range": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"release_channel": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"maintenance_policy": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_upgrade": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"maintenance_window": {
										Type:     schema.TypeSet,
										Computed: true,
										Set:      dayOfWeekHash,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"day": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"start_time": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"duration": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"etcd_cluster_size": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"master_location": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"zonal": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"regional": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"region": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"security_group_ids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Computed: true,
						},
						"internal_v4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_v4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_v6_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"internal_v4_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_v4_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_v6_endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cluster_ca_certificate": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"version_info": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"current_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"new_revision_available": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"new_revision_summary": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"version_deprecated": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"master_logging": {
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
									"kube_apiserver_enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"cluster_autoscaler_enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"events_enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"audit_enabled": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_policy_provider": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_provider": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"log_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_implementation": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cilium": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"routing_mode": {
										Type:     schema.TypeString,
										Computed: true,
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

func dataSourceYandexKubernetesClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.KubernetesClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve Kubernetes cluster by name: %v", err)
		}
	}

	cluster, err := config.sdk.Kubernetes().Cluster().Get(ctx, &k8s.GetClusterRequest{
		ClusterId: clusterID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kubernetes cluster with ID %q", clusterID))
	}

	err = flattenKubernetesClusterAttributes(cluster, d, false)
	if err != nil {
		return fmt.Errorf("failed to fill Kubernetes cluster attributes: %v", err)
	}

	return nil
}
