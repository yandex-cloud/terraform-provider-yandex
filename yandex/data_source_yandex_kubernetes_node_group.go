package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

func dataSourceYandexKubernetesNodeGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about a Yandex Kubernetes Node Group. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kubernetes/concepts/#node-group).\n\n~> One of `node_group_id` or `name` should be specified.\n",

		Read: dataSourceYandexKubernetesNodeGroupRead,
		Schema: map[string]*schema.Schema{
			"node_group_id": {
				Type:        schema.TypeString,
				Description: "ID of a specific Kubernetes node group.",
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
			"cluster_id": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesNodeGroup().Schema["cluster_id"].Description,
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
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
			"status": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesNodeGroup().Schema["status"].Description,
				Computed:    true,
			},
			"instance_template": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_runtime": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["container_runtime"].Description,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["container_runtime"].Elem.(*schema.Resource).Schema["type"].Description,
										Required:    true,
									},
								},
							},
						},
						"resources": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["resources"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:        schema.TypeFloat,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["resources"].Elem.(*schema.Resource).Schema["memory"].Description,
										Computed:    true,
									},
									"cores": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["resources"].Elem.(*schema.Resource).Schema["cores"].Description,
										Computed:    true,
									},
									"core_fraction": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["resources"].Elem.(*schema.Resource).Schema["core_fraction"].Description,
										Computed:    true,
									},
									"gpus": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["resources"].Elem.(*schema.Resource).Schema["gpus"].Description,
										Computed:    true,
									},
								},
							},
						},
						"boot_disk": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["boot_disk"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["boot_disk"].Elem.(*schema.Resource).Schema["size"].Description,
										Computed:    true,
									},
									"type": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["boot_disk"].Elem.(*schema.Resource).Schema["type"].Description,
										Computed:    true,
									},
								},
							},
						},
						"platform_id": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["platform_id"].Description,
							Computed:    true,
						},
						"nat": {
							Type:        schema.TypeBool,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["nat"].Description,
							Computed:    true,
						},
						"network_interface": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_ids": {
										Type:        schema.TypeSet,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["subnet_ids"].Description,
										Computed:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"ipv4": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv4"].Description,
										Computed:    true,
									},
									"ipv6": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv6"].Description,
										Computed:    true,
									},
									"nat": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["nat"].Description,
										Computed:    true,
									},
									"security_group_ids": {
										Type:        schema.TypeSet,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["security_group_ids"].Description,
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
										Computed:    true,
									},
									"ipv4_dns_records": {
										Type:        schema.TypeList,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv4_dns_records"].Description,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv4_dns_records"].Elem.(*schema.Resource).Schema["fqdn"].Description,
													Computed:    true,
												},
												"dns_zone_id": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv4_dns_records"].Elem.(*schema.Resource).Schema["dns_zone_id"].Description,
													Computed:    true,
												},
												"ttl": {
													Type:        schema.TypeInt,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv4_dns_records"].Elem.(*schema.Resource).Schema["ttl"].Description,
													Computed:    true,
												},
												"ptr": {
													Type:        schema.TypeBool,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv4_dns_records"].Elem.(*schema.Resource).Schema["ptr"].Description,
													Computed:    true,
												},
											},
										},
									},
									"ipv6_dns_records": {
										Type:        schema.TypeList,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv6_dns_records"].Description,
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv6_dns_records"].Elem.(*schema.Resource).Schema["fqdn"].Description,
													Computed:    true,
												},
												"dns_zone_id": {
													Type:        schema.TypeString,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv6_dns_records"].Elem.(*schema.Resource).Schema["dns_zone_id"].Description,
													Computed:    true,
												},
												"ttl": {
													Type:        schema.TypeInt,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv6_dns_records"].Elem.(*schema.Resource).Schema["ttl"].Description,
													Computed:    true,
												},
												"ptr": {
													Type:        schema.TypeBool,
													Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_interface"].Elem.(*schema.Resource).Schema["ipv6_dns_records"].Elem.(*schema.Resource).Schema["ptr"].Description,
													Computed:    true,
												},
											},
										},
									},
								},
							},
						},
						"network_acceleration_type": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["network_acceleration_type"].Description,
							Computed:    true,
						},
						"metadata": {
							Type:        schema.TypeMap,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["metadata"].Description,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
						"scheduling_policy": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["scheduling_policy"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"preemptible": {
										Type:        schema.TypeBool,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["scheduling_policy"].Elem.(*schema.Resource).Schema["preemptible"].Description,
										Computed:    true,
									},
								},
							},
						},
						"placement_policy": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["placement_policy"].Description,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"placement_group_id": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["placement_policy"].Elem.(*schema.Resource).Schema["placement_group_id"].Description,
										Required:    true,
									},
								},
							},
						},
						"name": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["name"].Description,
							Computed:    true,
						},
						"labels": {
							Type:        schema.TypeMap,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["labels"].Description,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
						"container_network": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["container_network"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pod_mtu": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["container_network"].Elem.(*schema.Resource).Schema["pod_mtu"].Description,
										Computed:    true,
									},
								},
							},
						},
						"gpu_settings": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["gpu_settings"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gpu_cluster_id": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["gpu_settings"].Elem.(*schema.Resource).Schema["gpu_cluster_id"].Description,
										Computed:    true,
									},
									"gpu_environment": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["instance_template"].Elem.(*schema.Resource).Schema["gpu_settings"].Elem.(*schema.Resource).Schema["gpu_environment"].Description,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"scale_policy": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["scale_policy"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_scale": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["scale_policy"].Elem.(*schema.Resource).Schema["fixed_scale"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["scale_policy"].Elem.(*schema.Resource).Schema["fixed_scale"].Elem.(*schema.Resource).Schema["size"].Description,
										Computed:    true,
									},
								},
							},
						},
						"auto_scale": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["scale_policy"].Elem.(*schema.Resource).Schema["auto_scale"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["scale_policy"].Elem.(*schema.Resource).Schema["auto_scale"].Elem.(*schema.Resource).Schema["min"].Description,
										Computed:    true,
									},
									"max": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["scale_policy"].Elem.(*schema.Resource).Schema["auto_scale"].Elem.(*schema.Resource).Schema["max"].Description,
										Computed:    true,
									},
									"initial": {
										Type:        schema.TypeInt,
										Description: resourceYandexKubernetesNodeGroup().Schema["scale_policy"].Elem.(*schema.Resource).Schema["auto_scale"].Elem.(*schema.Resource).Schema["initial"].Description,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"allocation_policy": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["allocation_policy"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:        schema.TypeList,
							Description: resourceYandexKubernetesNodeGroup().Schema["allocation_policy"].Elem.(*schema.Resource).Schema["location"].Description,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["allocation_policy"].Elem.(*schema.Resource).Schema["location"].Elem.(*schema.Resource).Schema["zone"].Description,
										Computed:    true,
									},
									"subnet_id": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["allocation_policy"].Elem.(*schema.Resource).Schema["location"].Elem.(*schema.Resource).Schema["subnet_id"].Description,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"instance_group_id": {
				Type:        schema.TypeString,
				Description: resourceYandexKubernetesNodeGroup().Schema["instance_group_id"].Description,
				Computed:    true,
			},
			"version_info": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["version_info"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"current_version": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesNodeGroup().Schema["version_info"].Elem.(*schema.Resource).Schema["current_version"].Description,
							Computed:    true,
						},
						"new_revision_available": {
							Type:        schema.TypeBool,
							Description: resourceYandexKubernetesNodeGroup().Schema["version_info"].Elem.(*schema.Resource).Schema["new_revision_available"].Description,
							Computed:    true,
						},
						"new_revision_summary": {
							Type:        schema.TypeString,
							Description: resourceYandexKubernetesNodeGroup().Schema["version_info"].Elem.(*schema.Resource).Schema["new_revision_summary"].Description,
							Computed:    true,
						},
						"version_deprecated": {
							Type:        schema.TypeBool,
							Description: resourceYandexKubernetesNodeGroup().Schema["version_info"].Elem.(*schema.Resource).Schema["version_deprecated"].Description,
							Computed:    true,
						},
					},
				},
			},
			"maintenance_policy": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["maintenance_policy"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_upgrade": {
							Type:        schema.TypeBool,
							Description: resourceYandexKubernetesNodeGroup().Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["auto_upgrade"].Description,
							Computed:    true,
						},
						"auto_repair": {
							Type:        schema.TypeBool,
							Description: resourceYandexKubernetesNodeGroup().Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["auto_repair"].Description,
							Computed:    true,
						},
						"maintenance_window": {
							Type:        schema.TypeSet,
							Description: resourceYandexKubernetesNodeGroup().Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Description,
							Computed:    true,
							Set:         dayOfWeekHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Elem.(*schema.Resource).Schema["day"].Description,
										Computed:    true,
									},
									"start_time": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Elem.(*schema.Resource).Schema["start_time"].Description,
										Computed:    true,
									},
									"duration": {
										Type:        schema.TypeString,
										Description: resourceYandexKubernetesNodeGroup().Schema["maintenance_policy"].Elem.(*schema.Resource).Schema["maintenance_window"].Elem.(*schema.Resource).Schema["duration"].Description,
										Computed:    true,
									},
								},
							},
						},
					},
				},
			},
			"allowed_unsafe_sysctls": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["allowed_unsafe_sysctls"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"node_labels": {
				Type:        schema.TypeMap,
				Description: resourceYandexKubernetesNodeGroup().Schema["node_labels"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"node_taints": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["node_taints"].Description,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"deploy_policy": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["deploy_policy"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_expansion": {
							Type:        schema.TypeInt,
							Description: resourceYandexKubernetesNodeGroup().Schema["deploy_policy"].Elem.(*schema.Resource).Schema["max_expansion"].Description,
							Computed:    true,
						},
						"max_unavailable": {
							Type:        schema.TypeInt,
							Description: resourceYandexKubernetesNodeGroup().Schema["deploy_policy"].Elem.(*schema.Resource).Schema["max_unavailable"].Description,
							Computed:    true,
						},
					},
				},
			},
			"workload_identity_federation": {
				Type:        schema.TypeList,
				Description: resourceYandexKubernetesNodeGroup().Schema["workload_identity_federation"].Description,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: resourceYandexKubernetesNodeGroup().Schema["workload_identity_federation"].Elem.(*schema.Resource).Schema["enabled"].Description,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceYandexKubernetesNodeGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "node_group_id", "name")
	if err != nil {
		return err
	}

	nodeGroupID := d.Get("node_group_id").(string)
	_, nodeGroupNameOk := d.GetOk("name")

	if nodeGroupNameOk {
		nodeGroupID, err = resolveObjectID(ctx, config, d, sdkresolvers.KubernetesNodeGroupResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source node-group by name: %v", err)
		}
	}

	ng, err := config.sdk.Kubernetes().NodeGroup().Get(ctx, &k8s.GetNodeGroupRequest{
		NodeGroupId: nodeGroupID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kubernetes node-group with ID %q", nodeGroupID))
	}

	err = flattenNodeGroupSchemaData(ng, d)
	if err != nil {
		return fmt.Errorf("failed to fill Kubernetes node-group shema: %v", err)
	}

	d.Set("node_group_id", ng.Id)
	return nil
}
