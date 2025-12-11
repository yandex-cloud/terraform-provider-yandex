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
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexComputeInstanceGroupDefaultTimeout = 30 * time.Minute
)

func resourceYandexComputeInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Description: "An Instance group resource. For more information, see [the official documentation](https://yandex.cloud/docs/compute/concepts/instance-groups/).",

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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["service_account_id"],
				Required:    true,
			},

			"instance_template": {
				Type:        schema.TypeList,
				Description: "The template for creating new instances.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:        schema.TypeList,
							Description: "Compute resource specifications for the instance.",
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:         schema.TypeFloat,
										Description:  "The memory size in GB.",
										Required:     true,
										ValidateFunc: FloatAtLeast(0.0),
									},

									"cores": {
										Type:        schema.TypeInt,
										Description: "The number of CPU cores for the instance.",
										Required:    true,
									},

									"gpus": {
										Type:        schema.TypeInt,
										Description: "If provided, specifies the number of GPU devices for the instance.",
										Optional:    true,
										ForceNew:    true,
									},

									"core_fraction": {
										Type:        schema.TypeInt,
										Description: "If provided, specifies baseline core performance as a percent.",
										Optional:    true,
										Default:     100,
									},
								},
							},
						},

						"boot_disk": {
							Type:        schema.TypeList,
							Description: "Boot disk specifications for the instance.",
							Required:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"initialize_params": {
										Type:          schema.TypeList,
										Description:   "Parameters for creating a disk alongside the instance.\n\n~> `image_id` or `snapshot_id` must be specified.\n",
										Optional:      true,
										MaxItems:      1,
										ConflictsWith: []string{"instance_template.boot_disk.disk_id"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description": {
													Type:        schema.TypeString,
													Description: "A description of the boot disk.",
													Optional:    true,
												},

												"size": {
													Type:         schema.TypeInt,
													Description:  "The size of the disk in GB.",
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},

												"type": {
													Type:        schema.TypeString,
													Description: "The disk type.",
													Optional:    true,
													Default:     "network-hdd",
												},

												"image_id": {
													Type:          schema.TypeString,
													Description:   "The disk image to initialize this disk from.",
													Optional:      true,
													Computed:      true,
													ConflictsWith: []string{"instance_template.0.boot_disk.initialize_params.snapshot_id"},
												},

												"snapshot_id": {
													Type:          schema.TypeString,
													Description:   "The snapshot to initialize this disk from.",
													Optional:      true,
													Computed:      true,
													ConflictsWith: []string{"instance_template.0.boot_disk.initialize_params.image_id"},
												},
											},
										},
									},

									"disk_id": {
										Type:          schema.TypeString,
										Description:   "The ID of the existing disk (such as those managed by yandex_compute_disk) to attach as a boot disk.",
										Optional:      true,
										ConflictsWith: []string{"instance_template.boot_disk.initialize_params"},
									},

									"mode": {
										Type:         schema.TypeString,
										Description:  "The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.",
										Optional:     true,
										Default:      "READ_WRITE",
										ValidateFunc: validation.StringInSlice([]string{"READ_WRITE"}, false),
									},

									"device_name": {
										Type:        schema.TypeString,
										Description: "This value can be used to reference the device under `/dev/disk/by-id/`.",
										Optional:    true,
										Computed:    true,
									},

									"name": {
										Type:        schema.TypeString,
										Description: "When set can be later used to change DiskSpec of actual disk.",
										Optional:    true,
									},
								},
							},
						},

						"network_interface": {
							Type:        schema.TypeList,
							Description: "Network specifications for the instance. This can be used multiple times for adding multiple interfaces.",
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"network_id": {
										Type:        schema.TypeString,
										Description: "The ID of the network.",
										Optional:    true,
									},

									"subnet_ids": {
										Type:        schema.TypeSet,
										Description: "The ID of the subnets to attach this interface to.",
										Optional:    true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},

									"ipv4": {
										Type:        schema.TypeBool,
										Description: "Allocate an IPv4 address for the interface. The default value is `true`.",
										Optional:    true,
										Default:     true,
									},

									"nat": {
										Type:        schema.TypeBool,
										Description: "Flag for using NAT.",
										Optional:    true,
										Computed:    true,
									},

									"nat_ip_address": {
										Type:        schema.TypeString,
										Description: "A public address that can be used to access the internet over NAT. Use `variables` to set.",
										Optional:    true,
									},

									"ipv6": {
										Type:        schema.TypeBool,
										Description: "If `true`, allocate an IPv6 address for the interface. The address will be automatically assigned from the specified subnet.",
										Optional:    true,
										Computed:    true,
									},

									"ip_address": {
										Type:        schema.TypeString,
										Description: "Manual set static IP address.",
										Optional:    true,
										Computed:    true,
									},

									"ipv6_address": {
										Type:        schema.TypeString,
										Description: "Manual set static IPv6 address.",
										Optional:    true,
										Computed:    true,
									},

									"security_group_ids": {
										Type:        schema.TypeSet,
										Description: "Security group (SG) `IDs` for network interface.",
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
										Optional:    true,
									},

									"dns_record": {
										Type:        schema.TypeList,
										Description: "List of DNS records.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:        schema.TypeString,
													Description: "DNS record FQDN (must have dot at the end).",
													Required:    true,
												},
												"dns_zone_id": {
													Type:        schema.TypeString,
													Description: "DNS zone id (if not set, private zone used).",
													Optional:    true,
												},
												"ttl": {
													Type:        schema.TypeInt,
													Description: "DNS record TTL.",
													Optional:    true,
												},
												"ptr": {
													Type:        schema.TypeBool,
													Description: "When set to `true`, also create PTR DNS record.",
													Optional:    true,
													Computed:    true,
												},
											},
										},
									},

									"ipv6_dns_record": {
										Type:        schema.TypeList,
										Description: "List of IPv6 DNS records.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:        schema.TypeString,
													Description: "DNS record FQDN (must have dot at the end).",
													Required:    true,
												},
												"dns_zone_id": {
													Type:        schema.TypeString,
													Description: "DNS zone id (if not set, private zone used).",
													Optional:    true,
												},
												"ttl": {
													Type:        schema.TypeInt,
													Description: "DNS record TTL.",
													Optional:    true,
												},
												"ptr": {
													Type:        schema.TypeBool,
													Description: "When set to `true`, also create PTR DNS record.",
													Optional:    true,
													Computed:    true,
												},
											},
										},
									},

									"nat_dns_record": {
										Type:        schema.TypeList,
										Description: "List of NAT DNS records.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:        schema.TypeString,
													Description: "DNS record FQDN (must have dot at the end).",
													Required:    true,
												},
												"dns_zone_id": {
													Type:        schema.TypeString,
													Description: "DNS zone id (if not set, private zone used).",
													Optional:    true,
												},
												"ttl": {
													Type:        schema.TypeInt,
													Description: "DNS record TTL.",
													Optional:    true,
												},
												"ptr": {
													Type:        schema.TypeBool,
													Description: "When set to `true`, also create PTR DNS record.",
													Optional:    true,
													Computed:    true,
												},
											},
										},
									},
								},
							},
						},

						"platform_id": {
							Type:        schema.TypeString,
							Description: "The ID of the hardware platform configuration for the instance.",
							Optional:    true,
							Default:     "standard-v1",
						},

						"description": {
							Type:        schema.TypeString,
							Description: "A description of the instance.",
							Optional:    true,
						},

						"metadata": {
							Type:        schema.TypeMap,
							Description: "A set of metadata key/value pairs to make available from within the instance.",
							Optional:    true,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},

						"labels": {
							Type:        schema.TypeMap,
							Description: "A set of key/value label pairs to assign to the instance.",
							Optional:    true,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},

						"secondary_disk": {
							Type:        schema.TypeList,
							Description: "A list of disks to attach to the instance.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"initialize_params": {
										Type:        schema.TypeList,
										Description: "Parameters used for creating a disk alongside the instance.\n\n~> `image_id` or `snapshot_id` must be specified.\n",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"description": {
													Type:        schema.TypeString,
													Description: "A description of the boot disk.",
													Optional:    true,
												},

												"size": {
													Type:         schema.TypeInt,
													Description:  "The size of the disk in GB.",
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
													Default:      8,
												},

												"type": {
													Type:        schema.TypeString,
													Description: "The disk type.",
													Optional:    true,
													Default:     "network-hdd",
												},

												"image_id": {
													Type:        schema.TypeString,
													Description: "The disk image to initialize this disk from.",
													Optional:    true,
												},

												"snapshot_id": {
													Type:        schema.TypeString,
													Description: "The snapshot to initialize this disk from.",
													Optional:    true,
												},
											},
										},
									},

									"disk_id": {
										Type:        schema.TypeString,
										Description: "ID of the existing disk. To set use variables.",
										Optional:    true,
									},

									"mode": {
										Type:         schema.TypeString,
										Description:  "The access mode to the disk resource. By default a disk is attached in `READ_WRITE` mode.",
										Optional:     true,
										Default:      "READ_WRITE",
										ValidateFunc: validation.StringInSlice([]string{"READ_ONLY", "READ_WRITE"}, false),
									},

									"device_name": {
										Type:        schema.TypeString,
										Description: "This value can be used to reference the device under `/dev/disk/by-id/`.",
										Optional:    true,
									},

									"name": {
										Type:        schema.TypeString,
										Description: "When set can be later used to change DiskSpec of actual disk.",
										Optional:    true,
									},
								},
							},
						},

						"filesystem": {
							Type:        schema.TypeSet,
							Description: "List of filesystems to attach to the instance.",
							Optional:    true,
							Set:         hashInstanceGroupFilesystem,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"filesystem_id": {
										Type:        schema.TypeString,
										Description: "ID of the filesystem that should be attached.",
										Required:    true,
									},

									"device_name": {
										Type:        schema.TypeString,
										Description: "Name of the device representing the filesystem on the instance.",
										Optional:    true,
									},

									"mode": {
										Type:         schema.TypeString,
										Description:  "Mode of access to the filesystem that should be attached. By default, filesystem is attached in `READ_WRITE` mode.",
										Optional:     true,
										Default:      "READ_WRITE",
										ValidateFunc: validation.StringInSlice([]string{"READ_WRITE", "READ_ONLY"}, false),
									},
								},
							},
						},

						"scheduling_policy": {
							Type:        schema.TypeList,
							Description: "The scheduling policy configuration.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"preemptible": {
										Type:        schema.TypeBool,
										Description: "Specifies if the instance is preemptible. Defaults to `false`.",
										Optional:    true,
										Default:     false,
									},
								},
							},
						},

						"service_account_id": {
							Type:        schema.TypeString,
							Description: "The ID of the service account authorized for this instance.",
							Optional:    true,
						},

						"network_settings": {
							Type:        schema.TypeList,
							Description: "Network acceleration type for instance.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Description: "Network acceleration type. By default a network is in `STANDARD` mode.",
										Optional:    true,
									},
								},
							},
						},

						"name": {
							Type:        schema.TypeString,
							Description: "Name template of the instance.\nIn order to be unique it must contain at least one of instance unique placeholders:*`{instance.short_id}`\n* `{instance.index}`\n* combination of `{instance.zone_id}` and`{instance.index_in_zone}`.\nExample: `my-instance-{instance.index}`.\nIf not set, default name is used: `{instance_group.id}-{instance.short_id}`. It may also contain another placeholders, see `metadata` doc for full list.",
							Optional:    true,
						},

						"hostname": {
							Type:        schema.TypeString,
							Description: "Hostname template for the instance. This field is used to generate the FQDN value of instance. The `hostname` must be unique within the network and region. If not specified, the hostname will be equal to `id` of the instance and FQDN will be `<id>.auto.internal`. Otherwise FQDN will be `<hostname>.<region_id>.internal`.\nIn order to be unique it must contain at least on of instance unique placeholders:\n* `{instance.short_id}`\n* {instance.index}\n* combination of `{instance.zone_id}` and `{instance.index_in_zone}`\nExample: `my-instance-{instance.index}`. If hostname is not set, `name` value will be used. It may also contain another placeholders, see `metadata` doc for full list.",
							Optional:    true,
						},

						"placement_policy": {
							Type:        schema.TypeList,
							Description: "The placement policy configuration.",
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"placement_group_id": {
										Type:        schema.TypeString,
										Description: "Specifies the id of the Placement Group to assign to the instances.",
										Required:    true,
									},
								},
							},
						},

						"metadata_options": {
							Type:        schema.TypeList,
							Description: "Options allow user to configure access to managed instances metadata",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gce_http_endpoint": {
										Type:         schema.TypeInt,
										Description:  "Enables access to GCE flavored metadata. Possible values: `0`, `1` for `enabled` and `2` for `disabled`.",
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
									"aws_v1_http_endpoint": {
										Type:         schema.TypeInt,
										Description:  "Enables access to AWS flavored metadata (IMDSv1). Possible values: `0`, `1` for `enabled` and `2` for `disabled`.",
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
									"gce_http_token": {
										Type:         schema.TypeInt,
										Description:  "Enables access to IAM credentials with GCE flavored metadata. Possible values: `0`, `1` for `enabled` and `2` for `disabled`.",
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
									"aws_v1_http_token": {
										Type:         schema.TypeInt,
										Description:  "Enables access to IAM credentials with AWS flavored metadata (IMDSv1). Possible values: `0`, `1` for `enabled` and `2` for `disabled`.",
										ValidateFunc: validation.IntBetween(0, 2),
										Optional:     true,
										Computed:     true,
									},
								},
							},
						},

						"reserved_instance_pool_id": {
							Type:        schema.TypeString,
							Description: "ID of the reserved instance pool that the instance should belong to.",
							Optional:    true,
						},
					},
				},
			},

			"variables": {
				Type:        schema.TypeMap,
				Description: "A set of key/value variables pairs to assign to the instance group.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},

			"scale_policy": {
				Type:        schema.TypeList,
				Description: "The scaling policy of the instance group.\n\n~> Either `fixed_scale` or `auto_scale` must be specified.\n",
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_scale": {
							Type:          schema.TypeList,
							Description:   "The fixed scaling policy of the instance group.",
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"scale_policy.0.auto_scale"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeInt,
										Description: "The number of instances in the instance group.",
										Required:    true,
									},
								},
							},
						},
						"auto_scale": {
							Type:          schema.TypeList,
							Description:   "The auto scaling policy of the instance group.",
							Optional:      true,
							MaxItems:      1,
							ConflictsWith: []string{"scale_policy.0.fixed_scale"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_scale_type": {
										Type:         schema.TypeString,
										Description:  "Autoscale type, can be `ZONAL` or `REGIONAL`. By default `ZONAL` type is used.",
										Optional:     true,
										Default:      "ZONAL",
										ValidateFunc: validation.StringInSlice([]string{"REGIONAL", "ZONAL"}, false),
									},
									"initial_size": {
										Type:        schema.TypeInt,
										Description: "The initial number of instances in the instance group.",
										Required:    true,
									},
									"measurement_duration": {
										Type:        schema.TypeInt,
										Description: "The amount of time, in seconds, that metrics are averaged for. If the average value at the end of the interval is higher than the `cpu_utilization_target`, the instance group will increase the number of virtual machines in the group.",
										Required:    true,
									},
									"min_zone_size": {
										Type:        schema.TypeInt,
										Description: "The minimum number of virtual machines in a single availability zone.",
										Optional:    true,
										Default:     0,
									},
									"max_size": {
										Type:        schema.TypeInt,
										Description: "The maximum number of virtual machines in the group.",
										Optional:    true,
										Default:     0,
									},
									"warmup_duration": {
										Type:        schema.TypeInt,
										Description: "The warm-up time of the virtual machine, in seconds. During this time, traffic is fed to the virtual machine, but load metrics are not taken into account.",
										Optional:    true,
										Computed:    true,
									},
									"stabilization_duration": {
										Type:        schema.TypeInt,
										Description: "The minimum time interval, in seconds, to monitor the load before an instance group can reduce the number of virtual machines in the group. During this time, the group will not decrease even if the average load falls below the value of `cpu_utilization_target`.",
										Optional:    true,
										Computed:    true,
									},
									"cpu_utilization_target": {
										Type:        schema.TypeFloat,
										Description: "Target CPU load level.",
										Optional:    true,
									},
									"custom_rule": {
										Type:        schema.TypeList,
										Description: "A list of custom rules.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rule_type": {
													Type:         schema.TypeString,
													Description:  "The metric rule type (UTILIZATION, WORKLOAD). UTILIZATION for metrics describing resource utilization per VM instance. WORKLOAD for metrics describing total workload on all VM instances.",
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"UTILIZATION", "WORKLOAD"}, false),
												},
												"metric_type": {
													Type:         schema.TypeString,
													Description:  "Type of metric, can be `GAUGE` or `COUNTER`. `GAUGE` metric reflects the value at particular time point. `COUNTER` metric exhibits a monotonous growth over time.",
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"GAUGE", "COUNTER"}, false),
												},
												"metric_name": {
													Type:        schema.TypeString,
													Description: "Name of the metric in Monitoring.",
													Required:    true,
												},
												"target": {
													Type:        schema.TypeFloat,
													Description: "Target metric value by which Instance Groups calculates the number of required VM instances.",
													Required:    true,
												},
												"labels": {
													Type:        schema.TypeMap,
													Description: "Metrics [labels](https://yandex.cloud/en/docs/monitoring/concepts/data-model#label) from Monitoring.",
													Optional:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
												},
												"folder_id": {
													Type:        schema.TypeString,
													Description: "If specified, sets the folder id to fetch metrics from. By default, it is the ID of the folder the group belongs to.",
													Optional:    true,
												},
												"service": {
													Type:        schema.TypeString,
													Description: "If specified, sets the service name to fetch metrics. The default value is `custom`. You can use a label to specify service metrics, e.g., `service` with the `compute` value for Compute Cloud.",
													Optional:    true,
												},
											},
										},
									},
								},
							},
						},
						"test_auto_scale": {
							Type:        schema.TypeList,
							Description: "The test auto scaling policy of the instance group. Use it to test how the auto scale works.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"auto_scale_type": {
										Type:         schema.TypeString,
										Description:  "Autoscale type, can be `ZONAL` or `REGIONAL`. By default `ZONAL` type is used.",
										Optional:     true,
										Default:      "ZONAL",
										ValidateFunc: validation.StringInSlice([]string{"REGIONAL", "ZONAL"}, false),
									},
									"initial_size": {
										Type:        schema.TypeInt,
										Description: "The initial number of instances in the instance group.",
										Required:    true,
									},
									"measurement_duration": {
										Type:        schema.TypeInt,
										Description: "The amount of time, in seconds, that metrics are averaged for. If the average value at the end of the interval is higher than the `cpu_utilization_target`, the instance group will increase the number of virtual machines in the group.",
										Required:    true,
									},
									"min_zone_size": {
										Type:        schema.TypeInt,
										Description: "The minimum number of virtual machines in a single availability zone.",
										Optional:    true,
										Default:     0,
									},
									"max_size": {
										Type:        schema.TypeInt,
										Description: "The maximum number of virtual machines in the group.",
										Optional:    true,
										Default:     0,
									},
									"warmup_duration": {
										Type:        schema.TypeInt,
										Description: "The warm-up time of the virtual machine, in seconds. During this time, traffic is fed to the virtual machine, but load metrics are not taken into account.",
										Optional:    true,
										Computed:    true,
									},
									"stabilization_duration": {
										Type:        schema.TypeInt,
										Description: "The minimum time interval, in seconds, to monitor the load before an instance group can reduce the number of virtual machines in the group. During this time, the group will not decrease even if the average load falls below the value of `cpu_utilization_target`.",
										Optional:    true,
										Computed:    true,
									},
									"cpu_utilization_target": {
										Type:        schema.TypeFloat,
										Description: "Target CPU load level.",
										Optional:    true,
									},
									"custom_rule": {
										Type:        schema.TypeList,
										Description: "A list of custom rules.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"rule_type": {
													Type:         schema.TypeString,
													Description:  "Rule type: `UTILIZATION` - This type means that the metric applies to one instance. First, Instance Groups calculates the average metric value for each instance, then averages the values for instances in one availability zone. This type of metric must have the `instance_id` label. `WORKLOAD` - This type means that the metric applies to instances in one availability zone. This type of metric must have the `zone_id` label.",
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"UTILIZATION", "WORKLOAD"}, false),
												},
												"metric_type": {
													Type:         schema.TypeString,
													Description:  "Metric type, `GAUGE` or `COUNTER`.",
													Required:     true,
													ValidateFunc: validation.StringInSlice([]string{"GAUGE", "COUNTER"}, false),
												},
												"metric_name": {
													Type:        schema.TypeString,
													Description: "The name of metric.",
													Required:    true,
												},
												"target": {
													Type:        schema.TypeFloat,
													Description: "Target metric value level.",
													Required:    true,
												},
												"labels": {
													Type:        schema.TypeMap,
													Description: "A map of labels of metric.",
													Optional:    true,
													Elem:        &schema.Schema{Type: schema.TypeString},
												},
												"folder_id": {
													Type:        schema.TypeString,
													Description: "Folder ID of custom metric in Yandex Monitoring that should be used for scaling.",
													Optional:    true,
												},
												"service": {
													Type:        schema.TypeString,
													Description: "Service of custom metric in Yandex Monitoring that should be used for scaling.",
													Optional:    true,
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
				Type:        schema.TypeList,
				Description: "The deployment policy of the instance group.",
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_unavailable": {
							Type:        schema.TypeInt,
							Description: "The maximum number of running instances that can be taken offline (stopped or deleted) at the same time during the update process.",
							Required:    true,
						},
						"max_expansion": {
							Type:        schema.TypeInt,
							Description: "The maximum number of instances that can be temporarily allocated above the group's target size during the update process.",
							Required:    true,
						},
						"max_deleting": {
							Type:        schema.TypeInt,
							Description: "The maximum number of instances that can be deleted at the same time.",
							Optional:    true,
							Default:     0,
						},
						"max_creating": {
							Type:        schema.TypeInt,
							Description: "The maximum number of instances that can be created at the same time.",
							Optional:    true,
							Default:     0,
						},
						"startup_duration": {
							Type:        schema.TypeInt,
							Description: "The amount of time in seconds to allow for an instance to start. Instance will be considered up and running (and start receiving traffic) only after the startup_duration has elapsed and all health checks are passed.",
							Optional:    true,
							Default:     0,
						},
						"strategy": {
							Type:         schema.TypeString,
							Description:  "Affects the lifecycle of the instance during deployment. If set to `proactive` (default), Instance Groups can forcefully stop a running instance. If `opportunistic`, Instance Groups does not stop a running instance. Instead, it will wait until the instance stops itself or becomes unhealthy.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"proactive", "opportunistic"}, false),
						},
					},
				},
			},

			"allocation_policy": {
				Type:        schema.TypeList,
				Description: "The allocation policy of the instance group by zone and region.",
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zones": {
							Type:        schema.TypeSet,
							Description: "A list of [availability zones](https://yandex.cloud/docs/overview/concepts/geo-scope).",
							Required:    true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"instance_tags_pool": {
							Type:        schema.TypeList,
							Description: "Array of availability zone IDs with list of instance tags.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:        schema.TypeString,
										Description: "Availability zone.",
										Required:    true,
									},
									"tags": {
										Type:        schema.TypeList,
										Description: "List of tags for instances in zone.",
										Required:    true,
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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Optional:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"health_check": {
				Type:         schema.TypeList,
				Description:  "Health check specifications.",
				MinItems:     1,
				Optional:     true,
				AtLeastOneOf: []string{},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Type:        schema.TypeInt,
							Description: "The interval to wait between health checks in seconds.",
							Optional:    true,
						},

						"timeout": {
							Type:        schema.TypeInt,
							Description: "The length of time to wait for a response before the health check times out in seconds.",
							Optional:    true,
						},

						"healthy_threshold": {
							Type:        schema.TypeInt,
							Description: "The number of successful health checks before the managed instance is declared healthy.",
							Optional:    true,
							Default:     2,
						},

						"unhealthy_threshold": {
							Type:        schema.TypeInt,
							Description: "The number of failed health checks before the managed instance is declared unhealthy.",
							Optional:    true,
							Default:     2,
						},

						"tcp_options": {
							Type:        schema.TypeList,
							Description: "TCP check options.",
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:        schema.TypeInt,
										Description: "The port used for TCP health checks.",
										Required:    true,
									},
								},
							},
						},

						"http_options": {
							Type:        schema.TypeList,
							Description: "HTTP check options.",
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"port": {
										Type:        schema.TypeInt,
										Description: "The port used for HTTP health checks.",
										Required:    true,
									},

									"path": {
										Type:        schema.TypeString,
										Description: "The URL path used for health check requests.",
										Required:    true,
									},
								},
							},
						},
					},
				},
			},

			"max_checking_health_duration": {
				Type:        schema.TypeInt,
				Description: "Timeout for waiting for the VM to become healthy. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.",
				Optional:    true,
			},

			"load_balancer": {
				Type:        schema.TypeList,
				Description: "Load balancing specifications.",
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_name": {
							Type:        schema.TypeString,
							Description: "The name of the target group.",
							Optional:    true,
							ForceNew:    true,
						},
						"target_group_description": {
							Type:        schema.TypeString,
							Description: "A description of the target group.",
							Optional:    true,
							ForceNew:    true,
						},
						"target_group_labels": {
							Type:        schema.TypeMap,
							Description: "A set of key/value label pairs.",
							Optional:    true,
							ForceNew:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
						"target_group_id": {
							Type:        schema.TypeString,
							Description: "The ID of the target group.",
							Computed:    true,
						},
						"status_message": {
							Type:        schema.TypeString,
							Description: "The status message of the target group.",
							Computed:    true,
						},
						"max_opening_traffic_duration": {
							Type:        schema.TypeInt,
							Description: "Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.",
							Optional:    true,
							ForceNew:    true,
						},
						"ignore_health_checks": {
							Type:        schema.TypeBool,
							Description: "Do not wait load balancer health checks.",
							Optional:    true,
						},
					},
				},
			},

			"application_load_balancer": {
				Type:        schema.TypeList,
				Description: "Application Load balancing (L7) specifications.",
				MaxItems:    1,
				Optional:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_group_name": {
							Type:        schema.TypeString,
							Description: "The name of the target group.",
							Optional:    true,
							ForceNew:    true,
						},
						"target_group_description": {
							Type:        schema.TypeString,
							Description: "A description of the target group.",
							Optional:    true,
							ForceNew:    true,
						},
						"target_group_labels": {
							Type:        schema.TypeMap,
							Description: "A set of key/value label pairs.",
							Optional:    true,
							ForceNew:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
						"target_group_id": {
							Type:        schema.TypeString,
							Description: "The ID of the target group.",
							Computed:    true,
						},
						"status_message": {
							Type:        schema.TypeString,
							Description: "The status message of the instance.",
							Computed:    true,
						},
						"max_opening_traffic_duration": {
							Type:        schema.TypeInt,
							Description: "Timeout for waiting for the VM to be checked by the load balancer. If the timeout is exceeded, the VM will be turned off based on the deployment policy. Specified in seconds.",
							Optional:    true,
							ForceNew:    true,
						},
						"ignore_health_checks": {
							Type:        schema.TypeBool,
							Description: "Do not wait load balancer health checks.",
							Optional:    true,
						},
					},
				},
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"instances": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "Instances block.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:        schema.TypeString,
							Description: "Status of the managed instance.",
							Computed:    true,
						},
						"status_changed_at": {
							Type:        schema.TypeString,
							Description: "The timestamp in [RFC3339](https://www.ietf.org/rfc/rfc3339.txt) text format when the status of the managed instance was last changed.",
							Computed:    true,
						},
						"instance_id": {
							Type:        schema.TypeString,
							Description: "The ID of the instance.",
							Computed:    true,
						},
						"fqdn": {
							Type:        schema.TypeString,
							Description: "The Fully Qualified Domain Name.",
							Computed:    true,
						},
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the managed instance.",
							Computed:    true,
						},
						"status_message": {
							Type:        schema.TypeString,
							Description: "The status message of the instance.",
							Computed:    true,
						},
						"zone_id": {
							Type:        schema.TypeString,
							Description: "The ID of the availability zone where the instance resides.",
							Computed:    true,
						},
						"network_interface": {
							Type:        schema.TypeList,
							Description: "An array with the network interfaces attached to the managed instance.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"index": {
										Type:        schema.TypeInt,
										Description: "The index of the network interface as generated by the server.",
										Computed:    true,
									},
									"mac_address": {
										Type:        schema.TypeString,
										Description: "The MAC address assigned to the network interface.",
										Computed:    true,
									},
									"ipv4": {
										Type:        schema.TypeBool,
										Description: "`True` if IPv4 address allocated for the network interface.",
										Computed:    true,
									},
									"ip_address": {
										Type:        schema.TypeString,
										Description: "The private IP address to assign to the instance. If empty, the address is automatically assigned from the specified subnet.",
										Computed:    true,
									},
									"ipv6": {
										Type:        schema.TypeBool,
										Description: "If `true`, allocate an IPv6 address for the interface. The address will be automatically assigned from the specified subnet.",
										Computed:    true,
									},
									"ipv6_address": {
										Type:        schema.TypeString,
										Description: "The private IPv6 address to assign to the instance.",
										Computed:    true,
									},
									"subnet_id": {
										Type:        schema.TypeString,
										Description: "The ID of the subnet to attach this interface to. The subnet must reside in the same zone where this instance was created.",
										Computed:    true,
									},
									"nat": {
										Type:        schema.TypeBool,
										Description: "The instance's public address for accessing the internet over NAT.",
										Computed:    true,
									},
									"nat_ip_address": {
										Type:        schema.TypeString,
										Description: "The public IP address of the instance.",
										Computed:    true,
									},
									"nat_ip_version": {
										Type:        schema.TypeString,
										Description: "The IP version for the public address.",
										Computed:    true,
									},
								},
							},
						},
						"instance_tag": {
							Type:        schema.TypeString,
							Description: "Managed instance tag.",
							Computed:    true,
						},
					},
				},
			},

			"status": {
				Type:        schema.TypeString,
				Description: "The status of the instance.",
				Computed:    true,
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Default:     false,
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
		"instance_template.0.filesystem":        "instance_template.filesystem_specs",
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
		"instance_template.reserved_instance_pool_id",
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
