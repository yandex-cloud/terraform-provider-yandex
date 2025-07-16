package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Cloud Managed Kubernetes Cluster. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kubernetes/concepts/#kubernetes-cluster).\n\n~> One of `cluster_id` or `name` should be specified.\n",

		Read: dataSourceYandexKubernetesClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "ID of a specific Kubernetes cluster.",
				Optional:    true,
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
				Computed:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Optional:    true,
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
			"network_id": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["network_id"].Description,
				Computed:    true,
			},
			"service_account_id": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["service_account_id"].Description,
				Computed:    true,
			},
			"node_service_account_id": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["node_service_account_id"].Description,
				Computed:    true,
			},
			"cluster_ipv4_range": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["cluster_ipv4_range"].Description,
				Computed:    true,
			},
			"cluster_ipv6_range": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["cluster_ipv6_range"].Description,
				Computed:    true,
			},
			"node_ipv4_cidr_mask_size": {
				Type:        schema.TypeInt,
				Description: resourceYandexKubernetesCluster().Schema["node_ipv4_cidr_mask_size"].Description,
				Computed:    true,
			},
			"service_ipv4_range": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["service_ipv4_range"].Description,
				Computed:    true,
			},
			"service_ipv6_range": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["service_ipv6_range"].Description,
				Computed:    true,
			},
			"release_channel": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["release_channel"].Description,
				Computed:    true,
			},
			"master": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesCluster().Schema["master"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["version"].Description,
							Computed:    true,
						},
						"public_ip": {
							Type:        schema.TypeBool,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["public_ip"].Description,
							Computed:    true,
						},
						"maintenance_policy": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["maintenance_policy"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_upgrade": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["auto_upgrade"].Description,
										Computed:    true,
									},
									"maintenance_window": {
										Type:        schema.TypeSet,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Description,
										Computed:    true,
										Set:         dayOfWeekHash,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"day": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Elem.(*schema.Resource).Schema["day"].Description,
													Computed:    true,
												},
												"start_time": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Elem.(*schema.Resource).Schema["start_time"].Description,
													Computed:    true,
												},
												"duration": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Elem.(*schema.Resource).Schema["duration"].Description,
													Computed:    true,
												},
											},
										},
									},
								},
							},
						},
						"etcd_cluster_size": {
							Type:        schema.TypeInt,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["etcd_cluster_size"].Description,
							Computed:    true,
						},
						"master_location": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_location"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_location"].Elem.(*schema.Resource).Schema["zone"].Description,
										Computed:    true,
									},
									"subnet_id": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_location"].Elem.(*schema.Resource).Schema["subnet_id"].Description,
										Computed:    true,
									},
								},
							},
						},
						"zonal": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["zonal"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["zonal"].Elem.(*schema.Resource).Schema["zone"].Description,
										Computed:    true,
									},
								},
							},
						},
						"regional": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["regional"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"region": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["regional"].Elem.(*schema.Resource).Schema["region"].Description,
										Computed:    true,
									},
								},
							},
						},
						"security_group_ids": {
							Type:        schema.TypeSet,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["security_group_ids"].Description,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Computed:    true,
						},
						"internal_v4_address": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["internal_v4_address"].Description,
							Computed:    true,
						},
						"external_v4_address": {
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["external_v4_address"].Description,
							Type:        schema.TypeString,
							Computed:    true,
						},
						"external_v6_address": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["external_v6_address"].Description,
							Computed:    true,
						},
						"internal_v4_endpoint": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["internal_v4_endpoint"].Description,
							Computed:    true,
						},
						"external_v4_endpoint": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["external_v4_endpoint"].Description,
							Computed:    true,
						},
						"external_v6_endpoint": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["external_v6_endpoint"].Description,
							Computed:    true,
						},
						"cluster_ca_certificate": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["cluster_ca_certificate"].Description,
							Computed:    true,
						},
						"version_info": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["version_info"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"current_version": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["version_info"].Elem.(*schema.Resource).Schema["current_version"].Description,
										Computed:    true,
									},
									"new_revision_available": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["version_info"].Elem.(*schema.Resource).Schema["new_revision_available"].Description,
										Computed:    true,
									},
									"new_revision_summary": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["version_info"].Elem.(*schema.Resource).Schema["new_revision_summary"].Description,
										Computed:    true,
									},
									"version_deprecated": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["version_info"].Elem.(*schema.Resource).Schema["version_deprecated"].Description,
										Computed:    true,
									},
								},
							},
						},
						"master_logging": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Elem.(*schema.Resource).Schema["enabled"].Description,
										Computed:    true,
									},
									"log_group_id": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Elem.(*schema.Resource).Schema["log_group_id"].Description,
										Computed:    true,
									},
									"folder_id": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Elem.(*schema.Resource).Schema["folder_id"].Description,
										Computed:    true,
									},
									"kube_apiserver_enabled": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Elem.(*schema.Resource).Schema["kube_apiserver_enabled"].Description,
										Computed:    true,
									},
									"cluster_autoscaler_enabled": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Elem.(*schema.Resource).Schema["cluster_autoscaler_enabled"].Description,
										Computed:    true,
									},
									"events_enabled": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Elem.(*schema.Resource).Schema["events_enabled"].Description,
										Computed:    true,
									},
									"audit_enabled": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["master_logging"].Elem.(*schema.Resource).Schema["audit_enabled"].Description,
										Computed:    true,
									},
								},
							},
						},
						"scale_policy": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["scale_policy"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_scale": {
										Type:        schema.TypeList,
										Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["scale_policy"].Elem.(*schema.Resource).Schema["auto_scale"].Description,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"min_resource_preset_id": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesCluster().Schema["master"].Elem.(*schema.Resource).Schema["scale_policy"].Elem.(*schema.Resource).Schema["auto_scale"].Elem.(*schema.Resource).Schema["min_resource_preset_id"].Description,
													Computed:    true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["status"].Description,
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["health"].Description,
				Computed:    true,
			},
			"network_policy_provider": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["network_policy_provider"].Description,
				Computed:    true,
			},
			"kms_provider": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesCluster().Schema["kms_provider"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesCluster().Schema["kms_provider"].Elem.(*schema.Resource).Schema["key_id"].Description,
							Computed:    true,
						},
					},
				},
			},
			"log_group_id": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesCluster().Schema["log_group_id"].Description,
				Computed:    true,
			},
			"network_implementation": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesCluster().Schema["network_implementation"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cilium": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesCluster().Schema["network_implementation"].Elem.(*schema.Resource).Schema["cilium"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"routing_mode": {
										Type:        schema.TypeString,
										Description: "The routing mode of the network interface.",
										Computed:    true,
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
