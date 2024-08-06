package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
)

const (
	yandexComputeInstanceGroupDefaultTimeout = 30 * time.Minute
)

func resourceYandexComputeInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexComputeInstanceGroupCreate,
		Read:   resourceYandexComputeInstanceGroupRead,
		Update: resourceYandexComputeInstanceGroupUpdate,
		Delete: resourceYandexComputeInstanceGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexComputeInstanceGroupDefaultTimeout),
			Update: schema.DefaultTimeout(yandexComputeInstanceGroupDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexComputeInstanceGroupDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"instance_template": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:         schema.TypeFloat,
										Required:     true,
										ValidateFunc: FloatAtLeast(0.0),
									},

									"cores": {
										Type:     schema.TypeInt,
										Required: true,
									},

									"gpus": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},

									"core_fraction": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  100,
									},
								},
							},
						},

						"boot_disk": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"initialize_params": {
										Type:          schema.TypeList,
										Optional:      true,
										MaxItems:      1,
										ConflictsWith: []string{"instance_template.boot_disk.disk_id"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"size": {
													Type:         schema.TypeInt,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},

												"type": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "network-hdd",
												},

												"image_id": {
													Type:          schema.TypeString,
													Optional:      true,
													Computed:      true,
													ConflictsWith: []string{"instance_template.0.boot_disk.initialize_params.snapshot_id"},
												},

												"snapshot_id": {
													Type:          schema.TypeString,
													Optional:      true,
													Computed:      true,
													ConflictsWith: []string{"instance_template.0.boot_disk.initialize_params.image_id"},
												},
											},
										},
									},

									"disk_id": {
										Type:          schema.TypeString,
										Optional:      true,
										ConflictsWith: []string{"instance_template.boot_disk.initialize_params"},
									},

									"mode": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "READ_WRITE",
										ValidateFunc: validation.StringInSlice([]string{"READ_WRITE"}, false),
									},

									"device_name": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},

									"name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},

						"network_interface": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_id": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"subnet_ids": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},

									"ipv4": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
									},

									"nat": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},

									"nat_ip_address": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"ipv6": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},

									"ip_address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},

									"ipv6_address": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},

									"security_group_ids": {
										Type:     schema.TypeSet,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Set:      schema.HashString,
										Optional: true,
									},

									"dns_record": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"dns_zone_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"ttl": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"ptr": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
											},
										},
									},

									"ipv6_dns_record": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"dns_zone_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"ttl": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"ptr": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
											},
										},
									},

									"nat_dns_record": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:     schema.TypeString,
													Required: true,
												},
												"dns_zone_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"ttl": {
													Type:     schema.TypeInt,
													Optional: true,
												},
												"ptr": {
													Type:     schema.TypeBool,
													Optional: true,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},

						"platform_id": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "standard-v1",
						},

						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"metadata": {
							Type:     schema.TypeMap,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
						},

						"labels": {
							Type:     schema.TypeMap,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"secondary_disk": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"initialize_params": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"size": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
													Default:      8,
												},

												"type": {
													Type:     schema.TypeString,
													Optional: true,
													Default:  "network-hdd",
												},

												"image_id": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"snapshot_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},

									"disk_id": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"mode": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "READ_WRITE",
										ValidateFunc: validation.StringInSlice([]string{"READ_ONLY", "READ_WRITE"}, false),
									},

									"device_name": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"name": {
										Type:     schema.TypeString,
										Optional: true,
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
										Required: true,
									},

									"device_name": {
										Type:     schema.TypeString,
										Optional: true,
									},

									"mode": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "READ_WRITE",
										ValidateFunc: validation.StringInSlice([]string{"READ_WRITE", "READ_ONLY"}, false),
									},
								},
							},
						},

						"scheduling_policy": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"preemptible": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},

						"service_account_id": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"network_settings": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},

						"name": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
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
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"scale_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_scale": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"scale_policy.0.auto_scale"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
						"auto_scale": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"scale_policy.0.fixed_scale"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_scale_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "ZONAL",
										ValidateFunc: validation.StringInSlice([]string{"REGIONAL", "ZONAL"}, false),
									},
									"initial_size": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"measurement_duration": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"min_zone_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
									"max_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
									"warmup_duration": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"stabilization_duration": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"cpu_utilization_target": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"custom_rule": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rule_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"UTILIZATION", "WORKLOAD"}, false),
												},
												"metric_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"GAUGE", "COUNTER"}, false),
												},
												"metric_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"target": {
													Type:     schema.TypeFloat,
													Required: true,
												},
												"labels": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"folder_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"service": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"test_auto_scale": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_scale_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "ZONAL",
										ValidateFunc: validation.StringInSlice([]string{"REGIONAL", "ZONAL"}, false),
									},
									"initial_size": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"measurement_duration": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"min_zone_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
									"max_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  0,
									},
									"warmup_duration": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"stabilization_duration": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"cpu_utilization_target": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"custom_rule": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rule_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"UTILIZATION", "WORKLOAD"}, false),
												},
												"metric_type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"GAUGE", "COUNTER"}, false),
												},
												"metric_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												"target": {
													Type:     schema.TypeFloat,
													Required: true,
												},
												"labels": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
												"folder_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"service": {
													Type:     schema.TypeString,
													Optional: true,
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
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_unavailable": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"max_expansion": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"max_deleting": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"max_creating": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"startup_duration": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"proactive", "opportunistic"}, false),
						},
					},
				},
			},

			"allocation_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zones": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"instance_tags_pool": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:     schema.TypeString,
										Required: true,
									},
									"tags": {
										Type:     schema.TypeList,
										Required: true,
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
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"health_check": {
				Type:         schema.TypeList,
				MinItems:     1,
				Optional:     true,
				AtLeastOneOf: []string{},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},

						"healthy_threshold": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  2,
						},

						"unhealthy_threshold": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  2,
						},

						"tcp_options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},

						"http_options": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:     schema.TypeInt,
										Required: true,
									},

									"path": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"max_checking_health_duration": {
				Type:     schema.TypeInt,
				Optional: true,
			},

			"load_balancer": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_description": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
							Optional: true,
							ForceNew: true,
						},
						"ignore_health_checks": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},

			"application_load_balancer": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_name": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_description": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"target_group_labels": {
							Type:     schema.TypeMap,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
							Optional: true,
							ForceNew: true,
						},
						"ignore_health_checks": {
							Type:     schema.TypeBool,
							Optional: true,
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

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceYandexComputeInstanceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateInstanceGroupRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.InstanceGroup().InstanceGroup().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create instance group: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create instance group: %s", err)
	}

	resp, err := op.Response()
	if err != nil {
		return fmt.Errorf("Instance group creation failed: %s", err)
	}

	instanceGroup, ok := resp.(*instancegroup.InstanceGroup)
	if !ok {
		return fmt.Errorf("Create response doesn't contain Instance group")
	}

	d.SetId(instanceGroup.Id)

	return resourceYandexComputeInstanceGroupRead(d, meta)
}

func resourceYandexComputeInstanceGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	instanceGroup, err := config.sdk.InstanceGroup().InstanceGroup().Get(ctx, &instancegroup.GetInstanceGroupRequest{
		InstanceGroupId: d.Id(),
		View:            instancegroup.InstanceGroupView_FULL,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance group %q", d.Id()))
	}

	instances, err := config.sdk.InstanceGroup().InstanceGroup().ListInstances(ctx, &instancegroup.ListInstanceGroupInstancesRequest{
		InstanceGroupId: d.Id(),
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Can't read instances for instance group with ID %q", d.Id()))
	}

	return flattenInstanceGroup(d, instanceGroup, instances.GetInstances())
}

func flattenInstanceGroup(d *schema.ResourceData, instanceGroup *instancegroup.InstanceGroup, instances []*instancegroup.ManagedInstance) error {
	d.Set("created_at", getTimestamp(instanceGroup.CreatedAt))
	d.Set("folder_id", instanceGroup.GetFolderId())
	d.Set("name", instanceGroup.GetName())
	d.Set("description", instanceGroup.GetDescription())
	d.Set("service_account_id", instanceGroup.GetServiceAccountId())
	d.Set("status", instanceGroup.GetStatus().String())
	d.Set("deletion_protection", instanceGroup.GetDeletionProtection())

	if err := d.Set("labels", instanceGroup.GetLabels()); err != nil {
		return err
	}

	template, err := flattenInstanceGroupInstanceTemplate(instanceGroup.GetInstanceTemplate())
	if err != nil {
		return err
	}
	if err := d.Set("instance_template", template); err != nil {
		return err
	}

	variables := flattenInstanceGroupVariable(instanceGroup.GetVariables())
	if err := d.Set("variables", variables); err != nil {
		return err
	}

	scalePolicy, err := flattenInstanceGroupScalePolicy(instanceGroup)
	if err != nil {
		return err
	}
	if err := d.Set("scale_policy", scalePolicy); err != nil {
		return err
	}

	deployPolicy, err := flattenInstanceGroupDeployPolicy(instanceGroup)
	if err != nil {
		return err
	}
	if err := d.Set("deploy_policy", deployPolicy); err != nil {
		return err
	}

	allocationPolicy, err := flattenInstanceGroupAllocationPolicy(instanceGroup)
	if err != nil {
		return err
	}

	if err := d.Set("allocation_policy", allocationPolicy); err != nil {
		return err
	}

	loadBalancerSpec, err := flattenInstanceGroupLoadBalancerSpec(instanceGroup)
	if err != nil {
		return err
	}

	if err := d.Set("load_balancer", loadBalancerSpec); err != nil {
		return err
	}

	applicationLoadBalancerSpec, err := flattenInstanceGroupApplicationLoadBalancerSpec(instanceGroup)
	if err != nil {
		return err
	}

	if err := d.Set("application_load_balancer", applicationLoadBalancerSpec); err != nil {
		return err
	}

	healthChecks, maxDuration, err := flattenInstanceGroupHealthChecks(instanceGroup)
	if err != nil {
		return err
	}

	if maxDuration != nil {
		d.Set("max_checking_health_duration", maxDuration)
	}

	inst, err := flattenInstanceGroupManagedInstances(instances)
	if err != nil {
		return err
	}

	err = d.Set("instances", inst)
	if err != nil {
		return err
	}

	return d.Set("health_check", healthChecks)
}

func resourceYandexComputeInstanceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareUpdateInstanceGroupRequest(d, config)
	if err != nil {
		return err
	}

	err = makeInstanceGroupUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return resourceYandexComputeInstanceGroupRead(d, meta)
}

func resourceYandexComputeInstanceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Instance group %q", d.Id())

	req := &instancegroup.DeleteInstanceGroupRequest{
		InstanceGroupId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.InstanceGroup().InstanceGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Instance group %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Instance group %q", d.Id())
	return nil
}

func prepareCreateInstanceGroupRequest(d *schema.ResourceData, meta *Config) (*instancegroup.CreateInstanceGroupRequest, error) {
	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating instance group: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating instance group: %s", err)
	}

	instanceTemplate, err := expandInstanceGroupInstanceTemplate(d, "instance_template.0", meta)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'instance_template' object of api request: %s", err)
	}

	scalePolicy, err := expandInstanceGroupScalePolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'scale_policy' object of api request: %s", err)
	}

	deployPolicy, err := expandInstanceGroupDeployPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'deploy_policy' object of api request: %s", err)
	}

	allocationPolicy, err := expandInstanceGroupAllocationPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'allocation_policy' object of api request: %s", err)
	}

	healthChecksSpec, err := expandInstanceGroupHealthCheckSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'health_checks_spec' object of api request: %s", err)
	}

	loadBalancerSpec, err := expandInstanceGroupLoadBalancerSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'load_balancer_spec' object of api request: %s", err)
	}

	applicationLoadBalancerSpec, err := expandInstanceGroupApplicationLoadBalancerSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'application_load_balancer' object of api request: %s", err)
	}

	variables, err := expandInstanceGroupVariables(d.Get("variables"))
	if err != nil {
		return nil, fmt.Errorf("Error creating 'variables' object of api request: %s", err)
	}

	deletionProtection := d.Get("deletion_protection")

	req := &instancegroup.CreateInstanceGroupRequest{
		FolderId:                    folderID,
		Name:                        d.Get("name").(string),
		Description:                 d.Get("description").(string),
		Labels:                      labels,
		InstanceTemplate:            instanceTemplate,
		ScalePolicy:                 scalePolicy,
		DeployPolicy:                deployPolicy,
		AllocationPolicy:            allocationPolicy,
		LoadBalancerSpec:            loadBalancerSpec,
		ApplicationLoadBalancerSpec: applicationLoadBalancerSpec,
		HealthChecksSpec:            healthChecksSpec,
		ServiceAccountId:            d.Get("service_account_id").(string),
		Variables:                   variables,
		DeletionProtection:          deletionProtection.(bool),
	}

	return req, nil
}

func prepareUpdateInstanceGroupRequest(d *schema.ResourceData, meta *Config) (*instancegroup.UpdateInstanceGroupRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("Error expanding labels while creating instance: %s", err)
	}

	instanceTemplate, err := expandInstanceGroupInstanceTemplate(d, "instance_template.0", meta)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'instance_template' object of api request: %s", err)
	}

	scalePolicy, err := expandInstanceGroupScalePolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'scale_policy' object of api request: %s", err)
	}

	deployPolicy, err := expandInstanceGroupDeployPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'deploy_policy' object of api request: %s", err)
	}

	allocationPolicy, err := expandInstanceGroupAllocationPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'allocation_policy' object of api request: %s", err)
	}

	healthChecksSpec, err := expandInstanceGroupHealthCheckSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'health_checks_spec' object of api request: %s", err)
	}

	loadBalancerSpec, err := expandInstanceGroupLoadBalancerSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error creating 'load_balancer_spec' object of api request: %s", err)
	}

	variables, err := expandInstanceGroupVariables(d.Get("variables"))
	if err != nil {
		return nil, fmt.Errorf("Error creating 'variables' object of api request: %s", err)
	}

	deletionProtection := d.Get("deletion_protection")

	var updatePath = getStaticUpdatePath()

	var instanceGroupTemplateFieldsMap = map[string]string{
		"instance_template.0.secondary_disk":    "instance_template.secondary_disk_specs",
		"instance_template.0.network_interface": "instance_template.network_interface_specs",
		"instance_template.0.filesystem":        "instance_template.filesystem",
	}

	for field, path := range instanceGroupTemplateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	req := &instancegroup.UpdateInstanceGroupRequest{
		InstanceGroupId:    d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		InstanceTemplate:   instanceTemplate,
		ScalePolicy:        scalePolicy,
		DeployPolicy:       deployPolicy,
		AllocationPolicy:   allocationPolicy,
		LoadBalancerSpec:   loadBalancerSpec,
		HealthChecksSpec:   healthChecksSpec,
		ServiceAccountId:   d.Get("service_account_id").(string),
		UpdateMask:         &field_mask.FieldMask{Paths: updatePath},
		Variables:          variables,
		DeletionProtection: deletionProtection.(bool),
	}

	return req, nil
}

func getStaticUpdatePath() []string {
	return []string{
		"name",
		"description",
		"labels",
		"instance_template.description",
		"instance_template.labels",
		"instance_template.platform_id",
		"instance_template.resources_spec",
		"instance_template.metadata",
		"instance_template.metadata_options",
		"instance_template.boot_disk_spec",
		"instance_template.scheduling_policy",
		"instance_template.placement_policy",
		"instance_template.service_account_id",
		"instance_template.network_settings",
		"instance_template.name",
		"instance_template.hostname",
		"variables",
		"scale_policy",
		"deploy_policy",
		"allocation_policy",
		"load_balancer_spec",
		"health_checks_spec",
		"service_account_id",
		"deletion_protection",
	}
}

func makeInstanceGroupUpdateRequest(req *instancegroup.UpdateInstanceGroupRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.InstanceGroup().InstanceGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Instance group %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Instance group %q: %s", d.Id(), err)
	}

	return nil
}
