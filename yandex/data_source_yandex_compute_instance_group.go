package yandex

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
)

func dataSourceYandexComputeInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexComputeInstanceGroupRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"instance_group_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_template": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									"cores": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"core_fraction": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"gpus": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},

						"boot_disk": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"disk_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"initialize_params": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"size": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"type": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"image_id": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"snapshot_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},

									"device_name": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"platform_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"metadata": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"network_interface": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"subnet_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},

									"ipv4": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"nat": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"nat_ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"ipv6": {
										Type:     schema.TypeBool,
										Computed: true,
									},

									"ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"ipv6_address": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"security_group_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
									},

									"dns_record": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"dns_zone_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"ttl": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"ptr": {
													Type:     schema.TypeBool,
													Computed: true,
												},
											},
										},
									},

									"ipv6_dns_record": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"dns_zone_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"ttl": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"ptr": {
													Type:     schema.TypeBool,
													Computed: true,
												},
											},
										},
									},

									"nat_dns_record": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"dns_zone_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"ttl": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"ptr": {
													Type:     schema.TypeBool,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},

						"secondary_disk": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"disk_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"initialize_params": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"size": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"type": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"image_id": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"snapshot_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},

									"device_name": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"filesystem": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      hashInstanceGroupFilesystem,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"filesystem_id": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"device_name": {
										Type:     schema.TypeString,
										Computed: true,
									},

									"mode": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"scheduling_policy": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"preemptible": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"network_settings": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},

						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"hostname": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"placement_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"placement_group_id": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},

						"metadata_options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gce_http_endpoint": {
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
									"aws_v1_http_endpoint": {
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
									"gce_http_token": {
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
									"aws_v1_http_token": {
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
								},
							},
						},
					},
				},
			},

			"variables": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
						"auto_scale": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_scale_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"min_zone_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"measurement_duration": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"warmup_duration": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"stabilization_duration": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"initial_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"cpu_utilization_target": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									"custom_rule": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rule_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"metric_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"metric_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"labels": {
													Type:     schema.TypeMap,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Computed: true,
												},
												"target": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
												"folder_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"service": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"test_auto_scale": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_scale_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"min_zone_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"max_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"measurement_duration": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"warmup_duration": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"stabilization_duration": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"initial_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"cpu_utilization_target": {
										Type:     schema.TypeFloat,
										Computed: true,
									},
									"custom_rule": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rule_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"metric_type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"metric_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"labels": {
													Type:     schema.TypeMap,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Computed: true,
												},
												"target": {
													Type:     schema.TypeFloat,
													Computed: true,
												},
												"folder_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"service": {
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
				},
			},

			"deploy_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_unavailable": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_expansion": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_deleting": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_creating": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"startup_duration": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"strategy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"allocation_policy": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"instance_tags_pool": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"tags": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
					},
				},
			},

			"name": {
				Type:     schema.TypeString,
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
			},

			"health_check": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"healthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"unhealthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"tcp_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},

						"http_options": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:     schema.TypeInt,
										Computed: true,
									},

									"path": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			"max_checking_health_duration": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"load_balancer": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_group_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_group_labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"target_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_opening_traffic_duration": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ignore_health_checks": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"application_load_balancer": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_group_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"target_group_labels": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"target_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"max_opening_traffic_duration": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"ignore_health_checks": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instances": {
				Computed: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_changed_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"instance_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"network_interface": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"index": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"mac_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ipv4": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"ipv6": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"ipv6_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"nat": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"nat_ip_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"nat_ip_version": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"instance_tag": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"load_balancer_state": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"application_balancer_state": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status_message": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexComputeInstanceGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := config.Context()

	instanceGroupID := d.Get("instance_group_id").(string)

	if instanceGroupID == "" {
		return fmt.Errorf("instance_group_id should be provided")
	}

	instanceGroup, err := config.sdk.InstanceGroup().InstanceGroup().Get(ctx, &instancegroup.GetInstanceGroupRequest{
		InstanceGroupId: instanceGroupID,
		View:            instancegroup.InstanceGroupView_FULL,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance group %q", d.Get("name").(string)))
	}

	instances, err := config.sdk.InstanceGroup().InstanceGroup().ListInstances(ctx, &instancegroup.ListInstanceGroupInstancesRequest{
		InstanceGroupId: instanceGroupID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Can't read instances for instance group with ID %q", instanceGroupID))
	}

	return flattenInstanceGroupDataSource(d, instanceGroup, instances.GetInstances())
}

func flattenInstanceGroupDataSource(d *schema.ResourceData, instanceGroup *instancegroup.InstanceGroup, instances []*instancegroup.ManagedInstance) error {
	err := flattenInstanceGroup(d, instanceGroup, instances)

	if err != nil {
		return err
	}

	loadBalancerState, err := flattenInstanceGroupLoadBalancerState(instanceGroup)
	if err != nil {
		return err
	}
	if err := d.Set("load_balancer_state", loadBalancerState); err != nil {
		return err
	}

	applicationLoadBalancerState, err := flattenInstanceGroupApplicationLoadBalancerState(instanceGroup)
	if err != nil {
		return err
	}
	if err := d.Set("application_load_balancer", applicationLoadBalancerState); err != nil {
		return err
	}

	d.SetId(instanceGroup.Id)

	return nil
}
