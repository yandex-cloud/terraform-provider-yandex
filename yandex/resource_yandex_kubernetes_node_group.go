package yandex

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexKubernetesNodeGroupReadTimeout   = 10 * time.Minute
	yandexKubernetesNodeGroupCreateTimeout = 60 * time.Minute
	yandexKubernetesNodeGroupUpdateTimeout = 60 * time.Minute
	yandexKubernetesNodeGroupDeleteTimeout = 20 * time.Minute
)

func resourceYandexKubernetesNodeGroup() *schema.Resource {
	return &schema.Resource{
		Description: "Creates a Yandex Managed Kubernetes Cluster Node Group. For more information, see [the official documentation](https://yandex.cloud/docs/managed-kubernetes/concepts/#node-group).",

		Create: resourceYandexKubernetesNodeGroupCreate,
		Read:   resourceYandexKubernetesNodeGroupRead,
		Update: resourceYandexKubernetesNodeGroupUpdate,
		Delete: resourceYandexKubernetesNodeGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexKubernetesNodeGroupCreateTimeout),
			Read:   schema.DefaultTimeout(yandexKubernetesNodeGroupReadTimeout),
			Update: schema.DefaultTimeout(yandexKubernetesNodeGroupUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexKubernetesNodeGroupDeleteTimeout),
		},
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the Kubernetes cluster that this node group belongs to.",
				Required:    true,
				ForceNew:    true,
			},
			"instance_template": {
				Type:        schema.TypeList,
				Description: "Template used to create compute instances in this Kubernetes node group.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"container_runtime": {
							Type:        schema.TypeList,
							Description: "Container runtime configuration.",
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Description:  "Type of container runtime. Values: `docker`, `containerd`.",
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"containerd", "docker"}, true),
									},
								},
							},
						},
						"resources": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:         schema.TypeFloat,
										Description:  "The memory size allocated to the instance.",
										Optional:     true,
										Computed:     true,
										ValidateFunc: FloatGreater(0.0),
									},
									"cores": {
										Type:         schema.TypeInt,
										Description:  "Number of CPU cores allocated to the instance.",
										Optional:     true,
										Computed:     true,
										ValidateFunc: IntGreater(0),
									},
									"core_fraction": {
										Type:        schema.TypeInt,
										Description: "Baseline core performance as a percent.",
										Optional:    true,
										Computed:    true,
									},
									"gpus": {
										Type:        schema.TypeInt,
										Description: "Number of GPU cores allocated to the instance.",
										Optional:    true,
										Default:     0,
									},
								},
							},
						},
						"boot_disk": {
							Type:        schema.TypeList,
							Description: "The specifications for boot disks that will be attached to the instance.",
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeInt,
										Description: "The size of the disk in GB. Allowed minimal size: 64 GB.",
										Optional:    true,
										Computed:    true,
									},
									"type": {
										Type:        schema.TypeString,
										Description: "The disk type.",
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"platform_id": {
							Type:        schema.TypeString,
							Description: "The ID of the hardware platform configuration for the node group compute instances.",
							Optional:    true,
							Computed:    true,
						},
						"nat": {
							Type:        schema.TypeBool,
							Description: "Enables NAT for node group compute instances.",
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
							Deprecated:  fieldDeprecatedForAnother("nat", "nat under network_interface"),
						},
						"network_interface": {
							Type:        schema.TypeList,
							Description: "An array with the network interfaces that will be attached to the instance.",
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"subnet_ids": {
										Type:        schema.TypeSet,
										Description: "The IDs of the subnets.",
										Required:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"ipv4": {
										Type:        schema.TypeBool,
										Description: "Allocate an IPv4 address for the interface. The default value is `true`.",
										Optional:    true,
										Default:     true,
									},
									"ipv6": {
										Type:        schema.TypeBool,
										Description: "If true, allocate an IPv6 address for the interface. The address will be automatically assigned from the specified subnet.",
										Optional:    true,
										Computed:    true,
									},
									"nat": {
										Type:        schema.TypeBool,
										Description: "A public address that can be used to access the internet over NAT.",
										Optional:    true,
										Computed:    true,
									},
									"security_group_ids": {
										Type:        schema.TypeSet,
										Description: "Security group IDs for network interface.",
										Elem:        &schema.Schema{Type: schema.TypeString},
										Set:         schema.HashString,
										Optional:    true,
									},
									"ipv4_dns_records": {
										Type:        schema.TypeList,
										Description: "List of configurations for creating ipv4 DNS records.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:        schema.TypeString,
													Description: "DNS record FQDN.",
													Required:    true,
												},
												"dns_zone_id": {
													Type:        schema.TypeString,
													Description: "DNS zone ID (if not set, private zone is used).",
													Optional:    true,
												},
												"ttl": {
													Type:        schema.TypeInt,
													Description: "DNS record TTL (in seconds).",
													Optional:    true,
												},
												"ptr": {
													Type:        schema.TypeBool,
													Description: "When set to `true`, also create a PTR DNS record.",
													Optional:    true,
												},
											},
										},
									},
									"ipv6_dns_records": {
										Type:        schema.TypeList,
										Description: "List of configurations for creating ipv6 DNS records.",
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"fqdn": {
													Type:        schema.TypeString,
													Description: "DNS record FQDN.",
													Required:    true,
												},
												"dns_zone_id": {
													Type:        schema.TypeString,
													Description: "DNS zone ID (if not set, private zone is used).",
													Optional:    true,
												},
												"ttl": {
													Type:        schema.TypeInt,
													Description: "DNS record TTL (in seconds).",
													Optional:    true,
												},
												"ptr": {
													Type:        schema.TypeBool,
													Description: "When set to `true`, also create a PTR DNS record.",
													Optional:    true,
												},
											},
										},
									},
								},
							},
						},
						"network_acceleration_type": {
							Type:         schema.TypeString,
							Description:  "Type of network acceleration. Values: `standard`, `software_accelerated`.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"standard", "software_accelerated"}, false),
						},
						"metadata": {
							Type:        schema.TypeMap,
							Description: "The set of metadata `key:value` pairs assigned to this instance template. This includes custom metadata and predefined keys. **Note**: key `user-data` won't be provided into instances. It reserved for internal activity in `kubernetes_node_group` resource.",
							Optional:    true,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
						"scheduling_policy": {
							Type:        schema.TypeList,
							Description: "The scheduling policy for the instances in node group.",
							MaxItems:    1,
							Optional:    true,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"preemptible": {
										Type:        schema.TypeBool,
										Description: "Specifies if the instance is preemptible. Defaults to `false`.",
										Optional:    true,
										Computed:    true,
									},
								},
							},
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
						"name": {
							Type:        schema.TypeString,
							Description: "Name template of the instance. In order to be unique it must contain at least one of instance unique placeholders:\n* `{instance.short_id}\n* `{instance.index}`\n* combination of `{instance.zone_id}` and `{instance.index_in_zone}`\n\nExample: `my-instance-{instance.index}`.\nIf not set, default is used: `{instance_group.id}-{instance.short_id}`. It may also contain another placeholders, see [Compute Instance group metadata doc](https://yandex.cloud/docs/compute/instancegroup/api-ref/grpc/InstanceGroup) for full list.",
							Optional:    true,
						},
						"labels": {
							Type:        schema.TypeMap,
							Description: "Labels that will be assigned to compute nodes (instances), created by the Node Group.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
						"container_network": {
							Type:        schema.TypeList,
							Description: "Container network configuration.",
							Computed:    true,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pod_mtu": {
										Type:        schema.TypeInt,
										Description: "MTU for pods.",
										Optional:    true,
										Computed:    true,
										ForceNew:    true,
									},
								},
							},
						},
						"gpu_settings": {
							Type:        schema.TypeList,
							Description: "GPU settings.",
							Computed:    true,
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"gpu_cluster_id": {
										Type:        schema.TypeString,
										Description: "GPU cluster id.",
										Optional:    true,
										ForceNew:    true,
									},
									"gpu_environment": {
										Type:         schema.TypeString,
										Description:  "GPU environment. Values: `runc`, `runc_drivers_cuda`.",
										Optional:     true,
										Computed:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"runc", "runc_drivers_cuda"}, false),
									},
								},
							},
						},
					},
				},
			},
			"scale_policy": {
				Type:        schema.TypeList,
				Description: "Scale policy of the node group.",
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_scale": {
							Type:          schema.TypeList,
							Description:   "Scale policy for a fixed scale node group.",
							MaxItems:      1,
							Optional:      true,
							ConflictsWith: []string{"scale_policy.0.auto_scale"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:        schema.TypeInt,
										Description: "The number of instances in the node group.",
										Optional:    true,
										Computed:    true,
									},
								},
							},
						},
						"auto_scale": {
							Type:          schema.TypeList,
							Description:   "Scale policy for an autoscaled node group.",
							MaxItems:      1,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"scale_policy.0.fixed_scale"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min": {
										Type:         schema.TypeInt,
										Description:  "Minimum number of instances in the node group.",
										Required:     true,
										ForceNew:     false,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"max": {
										Type:         schema.TypeInt,
										Description:  "Maximum number of instances in the node group.",
										Required:     true,
										ForceNew:     false,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"initial": {
										Type:         schema.TypeInt,
										Description:  "Initial number of instances in the node group.",
										Required:     true,
										ForceNew:     false,
										ValidateFunc: validation.IntAtLeast(0),
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
				Computed:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"version": {
				Type:        schema.TypeString,
				Description: "Version of Kubernetes that will be used for Kubernetes node group.",
				Optional:    true,
				Computed:    true,
			},
			"allocation_policy": {
				Type:        schema.TypeList,
				Description: "This argument specify subnets (zones), that will be used by node group compute instances.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:        schema.TypeList,
							Description: "Repeated field, that specify subnets (zones), that will be used by node group compute instances. Subnet specified by `subnet_id` should be allocated in zone specified by 'zone' argument.",
							Optional:    true,
							Computed:    true,
							ForceNew:    false,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:        schema.TypeString,
										Description: "ID of the availability zone where for one compute instance in node group.",
										Optional:    true,
										Computed:    true,
										ForceNew:    false,
									},
									"subnet_id": {
										Type:        schema.TypeString,
										Description: "ID of the subnet, that will be used by one compute instance in node group.",
										Optional:    true,
										Computed:    true,
										ForceNew:    true,
										Deprecated:  fieldDeprecatedForAnother("subnet_id", "subnet_ids under network_interface"),
									},
								},
							},
						},
					},
				},
			},
			"allowed_unsafe_sysctls": {
				Type:        schema.TypeList,
				Description: "A list of allowed unsafe `sysctl` parameters for this node group. For more details see [documentation](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster).",
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"node_labels": {
				Type:        schema.TypeMap,
				Description: "A set of key/value label pairs, that are assigned to all the nodes of this Kubernetes node group.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
			},
			"node_taints": {
				Type:        schema.TypeList,
				Description: "A list of Kubernetes taints, that are applied to all the nodes of this Kubernetes node group.",
				Optional:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"maintenance_policy": {
				Type:        schema.TypeList,
				Description: "Maintenance policy for this Kubernetes node group. If policy is omitted, automatic revision upgrades are enabled and could happen at any time. Revision upgrades are performed only within the same minor version, e.g. `1.29`. Minor version upgrades (e.g. `1.29`->`1.30`) should be performed manually.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_upgrade": {
							Type:        schema.TypeBool,
							Description: "Flag specifies if node group can be upgraded automatically. When omitted, default value is `true`.",
							Required:    true,
						},
						"auto_repair": {
							Type:        schema.TypeBool,
							Description: "Flag that specifies if node group can be repaired automatically. When omitted, default value is `true`.",
							Required:    true,
						},
						"maintenance_window": {
							Type:        schema.TypeSet,
							Description: "Set of day intervals, when maintenance is allowed for this node group. When omitted, it defaults to any time.\n\nTo specify time of day interval, for all days, one element should be provided, with two fields set, `start_time` and `duration`.\n\nTo allow maintenance only on specific days of week, please provide list of elements, with all fields set. Only one time interval is allowed for each day of week. Please see `my_node_group` config example.\n",
							Optional:    true,
							Set:         dayOfWeekHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"day": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateFunc:     validateParsableValue(parseDayOfWeek),
										DiffSuppressFunc: shouldSuppressDiffForDayOfWeek,
									},
									"start_time": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateFunc:     validateParsableValue(parseDayTime),
										DiffSuppressFunc: shouldSuppressDiffForTimeOfDay,
									},
									"duration": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateFunc:     validateParsableValue(parseDuration),
										DiffSuppressFunc: shouldSuppressDiffForTimeDuration,
									},
								},
							},
						},
					},
				},
			},
			"deploy_policy": {
				Type:        schema.TypeList,
				Description: "Deploy policy of the node group.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_expansion": {
							Type:        schema.TypeInt,
							Description: "The maximum number of instances that can be temporarily allocated above the group's target size during the update.",
							Required:    true,
						},
						"max_unavailable": {
							Type:        schema.TypeInt,
							Description: "The maximum number of running instances that can be taken offline during update.",
							Required:    true,
						},
					},
				},
			},
			"instance_group_id": {
				Type:        schema.TypeString,
				Description: "ID of instance group that is used to manage this Kubernetes node group.",
				Computed:    true,
			},
			"version_info": {
				Type:        schema.TypeList,
				Description: "Information about Kubernetes node group version.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"current_version": {
							Type:        schema.TypeString,
							Description: "Current Kubernetes version, major.minor (e.g. `1.30`).",
							Computed:    true,
						},
						"new_revision_available": {
							Type:        schema.TypeBool,
							Description: "True/false flag. Newer revisions may include Kubernetes patches (e.g `1.30.1` -> `1.30.2`) as well as some internal component updates - new features or bug fixes in yandex-specific components either on the master or nodes.",
							Computed:    true,
						},
						"new_revision_summary": {
							Type:        schema.TypeString,
							Description: "Human readable description of the changes to be applied when updating to the latest revision. Empty if new_revision_available is false.",
							Computed:    true,
						},
						"version_deprecated": {
							Type:        schema.TypeBool,
							Description: "True/false flag. The current version is on the deprecation schedule, component (master or node group) should be upgraded.",
							Computed:    true,
						},
					},
				},
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the Kubernetes node group.",
				Computed:    true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
		},
	}
}

func resourceYandexKubernetesNodeGroupCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateNodeGroupRequest(d)
	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Kubernetes().NodeGroup().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Kubernetes node group: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while get Kubernetes node group create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*k8s.CreateNodeGroupMetadata)
	if !ok {
		return fmt.Errorf("could not get Instance ID from create operation metadata")
	}

	d.SetId(md.GetNodeGroupId())

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to create Kubernetes node group: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Kubernetes node group creation failed: %s", err)
	}

	return resourceYandexKubernetesNodeGroupRead(d, meta)
}

func resourceYandexKubernetesNodeGroupRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ngID := d.Id()

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	ng, err := config.sdk.Kubernetes().NodeGroup().Get(ctx, &k8s.GetNodeGroupRequest{
		NodeGroupId: ngID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kubernetes node group with ID %q", ngID))
	}

	// resource only parameter
	d.Set("version", ng.GetVersionInfo().GetCurrentVersion())

	return flattenNodeGroupSchemaData(ng, d)
}

func prepareCreateNodeGroupRequest(d *schema.ResourceData) (*k8s.CreateNodeGroupRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while creating Kubernetes node group: %s", err)
	}

	tpl, err := getNodeGroupTemplate(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group template while creating Kubernetes node group: %s", err)
	}

	mp, err := getNodeGroupMaintenancePolicy(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group maintenance policy while creating Kubernetes node group: %s", err)
	}

	sp, err := getNodeGroupScalePolicy(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group scale policy for a Kubernetes node group creation: %s", err)
	}

	dp, err := getNodeGroupDeployPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group deploy policy while creating Kubernetes node group: %s", err)
	}

	sysctls := getNodeGroupAllowedUnsafeSysctls(d)
	nodeLabels := getNodeGroupNodeLabels(d)

	nodeTaints, err := getNodeGroupNodeTaints(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node taints for a Kubernetes node group creation: %s", err)
	}

	req := &k8s.CreateNodeGroupRequest{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Labels:               labels,
		ClusterId:            d.Get("cluster_id").(string),
		NodeTemplate:         tpl,
		ScalePolicy:          sp,
		AllocationPolicy:     getNodeGroupAllocationPolicy(d),
		Version:              d.Get("version").(string),
		MaintenancePolicy:    mp,
		AllowedUnsafeSysctls: sysctls,
		NodeLabels:           nodeLabels,
		NodeTaints:           nodeTaints,
		DeployPolicy:         dp,
	}

	return req, nil
}

func getNodeGroupMaintenancePolicy(d *schema.ResourceData) (*k8s.NodeGroupMaintenancePolicy, error) {
	if _, ok := d.GetOk("maintenance_policy"); !ok {
		return nil, nil
	}

	mp := &k8s.NodeGroupMaintenancePolicy{
		AutoUpgrade: d.Get("maintenance_policy.0.auto_upgrade").(bool),
		AutoRepair:  d.Get("maintenance_policy.0.auto_repair").(bool),
	}

	if mw, ok := d.GetOk("maintenance_policy.0.maintenance_window"); ok {
		var err error
		if mp.MaintenanceWindow, err = expandMaintenanceWindow(mw.(*schema.Set).List()); err != nil {
			return nil, err
		}
	}

	return mp, nil
}

func getNodeGroupAllocationPolicy(d *schema.ResourceData) *k8s.NodeGroupAllocationPolicy {
	return &k8s.NodeGroupAllocationPolicy{
		Locations: getNodeGroupAllocationPolicyLocationsFromConfig(d),
	}
}

// getNodeGroupAllocationPolicyLocationsFromConfig returns  AllocationPolicy Locations from config. It will NOT get values from state even if they absent in a config.
func getNodeGroupAllocationPolicyLocationsFromConfig(d *schema.ResourceData) []*k8s.NodeGroupLocation {
	var locations []*k8s.NodeGroupLocation

	ap := d.GetRawConfig().GetAttr("allocation_policy")
	if ap.IsNull() {
		return locations
	}

	aps := ap.AsValueSlice()
	if len(aps) < 1 {
		return locations
	}

	locs, ok := aps[0].AsValueMap()["location"]
	if !ok || locs.IsNull() {
		return locations
	}

	for _, locationAttr := range locs.AsValueSlice() {
		locationSpec := &k8s.NodeGroupLocation{}

		if zone, ok := locationAttr.AsValueMap()["zone"]; ok && !zone.IsNull() {
			locationSpec.ZoneId = zone.AsString()
		}

		if subnet, ok := locationAttr.AsValueMap()["subnet_id"]; ok && !subnet.IsNull() {
			locationSpec.SubnetId = subnet.AsString()
		}

		locations = append(locations, locationSpec)
	}
	return locations

}

func getNodeGroupScalePolicy(d *schema.ResourceData) (*k8s.ScalePolicy, error) {
	_, okFixed := d.GetOk("scale_policy.0.fixed_scale")
	_, okAuto := d.GetOk("scale_policy.0.auto_scale")
	switch {
	case !okFixed && !okAuto:
		return nil, fmt.Errorf("no scale policy has been specified for a node group")
	case okFixed && okAuto:
		return nil, fmt.Errorf("scale policy should be exactly one of fixed scale or auto scale")
	case okFixed:
		if size, ok := d.GetOk("scale_policy.0.fixed_scale.0.size"); ok {
			return &k8s.ScalePolicy{
				ScaleType: &k8s.ScalePolicy_FixedScale_{
					FixedScale: &k8s.ScalePolicy_FixedScale{
						Size: int64(size.(int)),
					},
				},
			}, nil
		}
		return nil, fmt.Errorf("no size has been specified for a node group with a fixed scale policy")
	default: // okAuto
		return &k8s.ScalePolicy{
			ScaleType: &k8s.ScalePolicy_AutoScale_{
				AutoScale: &k8s.ScalePolicy_AutoScale{
					MinSize:     int64(d.Get("scale_policy.0.auto_scale.0.min").(int)),
					MaxSize:     int64(d.Get("scale_policy.0.auto_scale.0.max").(int)),
					InitialSize: int64(d.Get("scale_policy.0.auto_scale.0.initial").(int)),
				},
			},
		}, nil
	}
}

func getNodeGroupDeployPolicy(d *schema.ResourceData) (*k8s.DeployPolicy, error) {
	if _, ok := d.GetOk("deploy_policy"); !ok {
		return nil, nil
	}

	dp := &k8s.DeployPolicy{
		MaxExpansion:   int64(d.Get("deploy_policy.0.max_expansion").(int)),
		MaxUnavailable: int64(d.Get("deploy_policy.0.max_unavailable").(int)),
	}
	return dp, nil
}

func getNodeGroupAllowedUnsafeSysctls(d *schema.ResourceData) []string {
	obj := d.Get("allowed_unsafe_sysctls")
	if obj == nil {
		return nil
	}
	var sysctls []string
	for _, s := range obj.([]interface{}) {
		sysctls = append(sysctls, s.(string))
	}
	return sysctls
}

func getNodeGroupNodeLabels(d *schema.ResourceData) map[string]string {
	obj := d.Get("node_labels")
	if obj == nil {
		return nil
	}
	m := map[string]string{}
	for k, v := range obj.(map[string]interface{}) {
		m[k] = v.(string)
	}
	return m
}

func getNodeGroupNodeTaints(d *schema.ResourceData) ([]*k8s.Taint, error) {
	obj := d.Get("node_taints")
	if obj == nil {
		return nil, nil
	}
	var taints []*k8s.Taint
	for _, v := range obj.([]interface{}) {
		taint, err := parseTaint(v.(string))
		if err != nil {
			return nil, err
		}
		taints = append(taints, taint)
	}
	return taints, nil
}

// parseTaint parses a taint from a string, whose form must be
// '<key>=<value>:<effect>'.
func parseTaint(st string) (*k8s.Taint, error) {
	parts := strings.Split(st, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid taint spec: %v", st)
	}
	effect, err := toAPI(parts[1])
	if err != nil {
		return nil, err
	}

	partsKV := strings.Split(parts[0], "=")
	if len(partsKV) != 2 {
		return nil, fmt.Errorf("invalid taint spec: %v", st)
	}

	return &k8s.Taint{
		Key:    partsKV[0],
		Value:  partsKV[1],
		Effect: effect,
	}, nil
}

func toAPI(effect string) (k8s.Taint_Effect, error) {
	switch effect {
	case "NoSchedule":
		return k8s.Taint_NO_SCHEDULE, nil
	case "PreferNoSchedule":
		return k8s.Taint_PREFER_NO_SCHEDULE, nil
	case "NoExecute":
		return k8s.Taint_NO_EXECUTE, nil
	default:
		supported := []string{
			"NoSchedule",
			"PreferNoSchedule",
			"NoExecute",
		}
		return 0, fmt.Errorf("invalid taint effect: %v, supported taint effects %s", effect, strings.Join(supported, ", "))
	}
}

func getNodeGroupTemplate(d *schema.ResourceData) (*k8s.NodeTemplate, error) {
	h := schemaHelper(d, "instance_template.0.")

	metadata, err := expandLabels(h.Get("metadata"))
	if err != nil {
		return nil, fmt.Errorf("error expanding metadata while creating Kubernetes node group: %s", err)
	}

	labels, err := expandLabels(h.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding template labels while creating Kubernetes node group: %s", err)
	}

	ns, err := getNodeGroupTemplateNetworkSettings(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding metadata while creating Kubernetes node group: %s", err)
	}

	crs, err := getNodeGroupContainerRuntimeSettings(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding metadata while creating Kubernetes node group: %s", err)
	}

	cns, err := getNodeGroupContainerNetworkSettings(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding container network while creating Kubernetes node group: %s", err)
	}
	gpuSettings, err := getNodeGroupGPUSettings(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding gpu_settings while creating Kubernetes node group: %s", err)
	}
	tpl := &k8s.NodeTemplate{
		PlatformId:               h.GetString("platform_id"),
		ResourcesSpec:            getNodeGroupResourceSpec(d),
		BootDiskSpec:             getNodeGroupBootDiskSpec(d),
		Metadata:                 metadata,
		V4AddressSpec:            getNodeGroupAddressSpec(d),
		SchedulingPolicy:         getNodeGroupTemplateSchedulingPolicy(d),
		NetworkInterfaceSpecs:    getNodeGroupNetworkInterfaceSpecs(d),
		PlacementPolicy:          getNodeGroupTemplatePlacementPolicy(d),
		NetworkSettings:          ns,
		ContainerRuntimeSettings: crs,
		ContainerNetworkSettings: cns,
		GpuSettings:              gpuSettings,
		Name:                     h.GetString("name"),
		Labels:                   labels,
	}

	return tpl, nil
}

func getNodeGroupTemplateSchedulingPolicy(d *schema.ResourceData) *k8s.SchedulingPolicy {
	if preemptible, ok := d.GetOk("instance_template.0.scheduling_policy.0.preemptible"); ok {
		return &k8s.SchedulingPolicy{
			Preemptible: preemptible.(bool),
		}
	}

	return nil
}

func getNodeGroupTemplatePlacementPolicy(d *schema.ResourceData) *k8s.PlacementPolicy {
	if placementGroupId, ok := d.GetOk("instance_template.0.placement_policy.0.placement_group_id"); ok {
		return &k8s.PlacementPolicy{
			PlacementGroupId: placementGroupId.(string),
		}
	}

	return nil
}

func getNodeGroupTemplateNetworkSettings(d *schema.ResourceData) (*k8s.NodeTemplate_NetworkSettings, error) {
	if v, ok := d.GetOk("instance_template.0.network_acceleration_type"); ok {
		typeVal, ok := k8s.NodeTemplate_NetworkSettings_Type_value[strings.ToUpper(v.(string))]
		if !ok {
			return nil, fmt.Errorf("value for 'network_acceleration_type' should be 'standard' or 'software_accelerated'', not '%s'", v)
		}
		return &k8s.NodeTemplate_NetworkSettings{
			Type: k8s.NodeTemplate_NetworkSettings_Type(typeVal),
		}, nil
	}
	return nil, nil
}

func getNodeGroupNetworkInterfaceSpecs(d *schema.ResourceData) []*k8s.NetworkInterfaceSpec {
	var nifs []*k8s.NetworkInterfaceSpec
	h := schemaHelper(d, "instance_template.0.network_interface.")
	nifCount := h.GetInt("#")
	for i := 0; i < nifCount; i++ {
		nif := h.Get(fmt.Sprintf("%d", i)).(map[string]interface{})
		nifSpec := &k8s.NetworkInterfaceSpec{}

		if securityGroups, ok := nif["security_group_ids"]; ok {
			nifSpec.SecurityGroupIds = expandSecurityGroupIds(securityGroups)
		}

		if ipv4, ok := nif["ipv4"]; ok && ipv4.(bool) {
			nifSpec.PrimaryV4AddressSpec = &k8s.NodeAddressSpec{}
		}
		if ipv6, ok := nif["ipv6"]; ok && ipv6.(bool) {
			nifSpec.PrimaryV6AddressSpec = &k8s.NodeAddressSpec{}
		}

		if nat, ok := nif["nat"]; ok && nat.(bool) {
			nifSpec.PrimaryV4AddressSpec = &k8s.NodeAddressSpec{
				OneToOneNatSpec: &k8s.OneToOneNatSpec{
					IpVersion: k8s.IpVersion_IPV4,
				},
			}
		}

		if subnets, ok := nif["subnet_ids"]; ok {
			nifSpec.SubnetIds = expandSubnetIds(subnets)
		}

		if rec, ok := nif["ipv4_dns_records"]; ok {
			if nifSpec.PrimaryV4AddressSpec != nil {
				nifSpec.PrimaryV4AddressSpec.DnsRecordSpecs = expandK8SNodeGroupDNSRecords(rec.([]interface{}))
			}
		}
		if rec, ok := nif["ipv6_dns_records"]; ok {
			if nifSpec.PrimaryV6AddressSpec != nil {
				nifSpec.PrimaryV6AddressSpec.DnsRecordSpecs = expandK8SNodeGroupDNSRecords(rec.([]interface{}))
			}
		}

		nifs = append(nifs, nifSpec)
	}
	return nifs
}

func getNodeGroupAddressSpec(d *schema.ResourceData) *k8s.NodeAddressSpec {
	if nat, ok := d.GetOk("instance_template.0.nat"); ok && nat.(bool) {
		return &k8s.NodeAddressSpec{
			OneToOneNatSpec: &k8s.OneToOneNatSpec{
				IpVersion: k8s.IpVersion_IPV4,
			},
		}
	}

	return nil
}

func getNodeGroupBootDiskSpec(d *schema.ResourceData) *k8s.DiskSpec {
	h := schemaHelper(d, "instance_template.0.boot_disk.0.")
	spec := &k8s.DiskSpec{
		DiskTypeId: h.GetString("type"),
		DiskSize:   toBytes(h.GetInt("size")),
	}
	return spec
}

func getNodeGroupResourceSpec(d *schema.ResourceData) *k8s.ResourcesSpec {
	h := schemaHelper(d, "instance_template.0.resources.0.")
	spec := &k8s.ResourcesSpec{
		Memory:       toBytesFromFloat(h.Get("memory").(float64)),
		Cores:        int64(h.GetInt("cores")),
		CoreFraction: int64(h.GetInt("core_fraction")),
		Gpus:         int64(h.GetInt("gpus")),
	}
	return spec
}

func getNodeGroupContainerRuntimeSettings(d *schema.ResourceData) (*k8s.NodeTemplate_ContainerRuntimeSettings, error) {
	if v, ok := d.GetOk("instance_template.0.container_runtime.0.type"); ok {
		typeVal, ok := k8s.NodeTemplate_ContainerRuntimeSettings_Type_value[strings.ToUpper(v.(string))]
		if !ok {
			return nil, fmt.Errorf("value for 'type' should be 'containerd' or 'docker', not '%s'", v)
		}
		return &k8s.NodeTemplate_ContainerRuntimeSettings{
			Type: k8s.NodeTemplate_ContainerRuntimeSettings_Type(typeVal),
		}, nil
	}
	return nil, nil
}

func getNodeGroupContainerNetworkSettings(d *schema.ResourceData) (*k8s.NodeTemplate_ContainerNetworkSettings, error) {
	if _, ok := d.GetOk("instance_template.0.container_network"); !ok {
		return nil, nil
	}
	cns := &k8s.NodeTemplate_ContainerNetworkSettings{}
	if podMTU, ok := d.GetOk("instance_template.0.container_network.0.pod_mtu"); ok {
		cns.SetPodMtu(int64(podMTU.(int)))
	}
	return cns, nil
}

func getNodeGroupGPUSettings(d *schema.ResourceData) (*k8s.GpuSettings, error) {
	if _, ok := d.GetOk("instance_template.0.gpu_settings"); !ok {
		return nil, nil
	}
	gs := &k8s.GpuSettings{}
	if gpuClusterID, ok := d.GetOk("instance_template.0.gpu_settings.0.gpu_cluster_id"); ok {
		gs.SetGpuClusterId(gpuClusterID.(string))
	}
	if gpuEnvironment, ok := d.GetOk("instance_template.0.gpu_settings.0.gpu_environment"); ok {
		typeVal, ok := k8s.GpuSettings_GpuEnvironment_value[strings.ToUpper(gpuEnvironment.(string))]
		if !ok {
			return nil, fmt.Errorf("value for 'gpu_environment' should be 'runc' or 'runc_drivers_cuda'', not '%s'", gpuEnvironment)
		}
		gs.SetGpuEnvironment(k8s.GpuSettings_GpuEnvironment(typeVal))
	}
	return gs, nil
}

func flattenNodeGroupSchemaData(ng *k8s.NodeGroup, d *schema.ResourceData) error {
	d.Set("cluster_id", ng.ClusterId)
	d.Set("created_at", getTimestamp(ng.CreatedAt))
	d.Set("name", ng.Name)
	d.Set("description", ng.Description)
	d.Set("status", strings.ToLower(ng.Status.String()))
	d.Set("instance_group_id", ng.GetInstanceGroupId())

	if err := d.Set("labels", ng.Labels); err != nil {
		return err
	}

	tpl := flattenKubernetesNodeGroupTemplate(ng.GetNodeTemplate())
	if err := d.Set("instance_template", tpl); err != nil {
		return err
	}

	scalePolicy := flattenKubernetesNodeScalePolicy(ng.GetScalePolicy())
	if err := d.Set("scale_policy", scalePolicy); err != nil {
		return err
	}

	allocationPolicy := flattenKubernetesNodeGroupAllocationPolicy(ng.GetAllocationPolicy())
	if err := d.Set("allocation_policy", allocationPolicy); err != nil {
		return err
	}

	versionInfo := flattenKubernetesNodeGroupVersionInfo(ng.GetVersionInfo())
	if err := d.Set("version_info", versionInfo); err != nil {
		return err
	}

	maintenancePolicy, err := flattenKubernetesNodeGroupMaintenancePolicy(ng.GetMaintenancePolicy())
	if err != nil {
		return err
	}

	if err := d.Set("maintenance_policy", maintenancePolicy); err != nil {
		return err
	}

	deployPolicy, err := flattenKubernetesNodeGroupDeployPolicy(ng.GetDeployPolicy())
	if err != nil {
		return err
	}

	if err := d.Set("deploy_policy", deployPolicy); err != nil {
		return err
	}

	if err := d.Set("allowed_unsafe_sysctls", ng.AllowedUnsafeSysctls); err != nil {
		return err
	}

	if err := d.Set("node_labels", ng.NodeLabels); err != nil {
		return err
	}

	taints := flattenKubernetesNodeGroupTaints(ng.NodeTaints)
	if err := d.Set("node_taints", taints); err != nil {
		return err
	}

	d.SetId(ng.Id)
	return nil
}

func flattenKubernetesNodeGroupTaints(taints []*k8s.Taint) interface{} {
	var values []interface{}
	for _, t := range taints {
		var effect string
		switch t.GetEffect() {
		case k8s.Taint_NO_SCHEDULE:
			effect = "NoSchedule"
		case k8s.Taint_PREFER_NO_SCHEDULE:
			effect = "PreferNoSchedule"
		case k8s.Taint_NO_EXECUTE:
			effect = "NoExecute"
		}
		values = append(values, fmt.Sprintf("%s=%s:%s", t.GetKey(), t.GetValue(), effect))
	}
	return values
}

var nodeGroupUpdateFieldsMap = map[string]string{
	"name":                                                      "name",
	"description":                                               "description",
	"labels":                                                    "labels",
	"node_labels":                                               "node_labels",
	"instance_template.0.platform_id":                           "node_template.platform_id",
	"instance_template.0.metadata":                              "node_template.metadata",
	"instance_template.0.resources.0.memory":                    "node_template.resources_spec.memory",
	"instance_template.0.resources.0.cores":                     "node_template.resources_spec.cores",
	"instance_template.0.resources.0.gpus":                      "node_template.resources_spec.gpus",
	"instance_template.0.resources.0.core_fraction":             "node_template.resources_spec.core_fraction",
	"instance_template.0.boot_disk.0.type":                      "node_template.boot_disk_spec.disk_type_id",
	"instance_template.0.boot_disk.0.size":                      "node_template.boot_disk_spec.disk_size",
	"instance_template.0.scheduling_policy.0.preemptible":       "node_template.scheduling_policy.preemptible",
	"instance_template.0.placement_policy.0.placement_group_id": "node_template.placement_policy.placement_group_id",
	"instance_template.0.network_interface":                     "node_template.network_interface_specs",
	"instance_template.0.network_acceleration_type":             "node_template.network_settings",
	"instance_template.0.container_runtime.0.type":              "node_template.container_runtime_settings.type",
	"instance_template.0.name":                                  "node_template.name",
	"instance_template.0.labels":                                "node_template.labels",
	"allocation_policy":                                         "allocation_policy.locations",
	"scale_policy.0.fixed_scale.0.size":                         "scale_policy.fixed_scale.size",
	"scale_policy.0.auto_scale.0.min":                           "scale_policy.auto_scale.min_size",
	"scale_policy.0.auto_scale.0.max":                           "scale_policy.auto_scale.max_size",
	"scale_policy.0.auto_scale.0.initial":                       "scale_policy.auto_scale.initial_size",
	"version":                                                   "version",
	"maintenance_policy":                                        "maintenance_policy",
	"deploy_policy.0.max_expansion":                             "deploy_policy.max_expansion",
	"deploy_policy.0.max_unavailable":                           "deploy_policy.max_unavailable",
}

func resourceYandexKubernetesNodeGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ngID := d.Id()
	log.Printf("[DEBUG] updating Kubernetes node group %q", ngID)

	req, err := getKubernetesNodeGroupUpdateRequest(d)
	if err != nil {
		return err
	}

	var updatePath []string
	for field, path := range nodeGroupUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	if len(updatePath) == 0 {
		return fmt.Errorf("error while updating Kubernetes node group, didn't detect any changes")
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Kubernetes().NodeGroup().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update Kubernetes node group %q: %s", ngID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error updating Kubernetes node group %q: %s", ngID, err)
	}

	return resourceYandexKubernetesNodeGroupRead(d, meta)
}

func getKubernetesNodeGroupUpdateRequest(d *schema.ResourceData) (*k8s.UpdateNodeGroupRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating Kubernetes node group: %s", err)
	}

	tpl, err := getNodeGroupTemplate(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group template while updating Kubernetes node group: %s", err)
	}

	mp, err := getNodeGroupMaintenancePolicy(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group maintenance policy while updating Kubernetes node group: %s", err)
	}

	sp, err := getNodeGroupScalePolicy(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group scale policy for a Kubernetes node group update: %s", err)
	}

	dp, err := getNodeGroupDeployPolicy(d)
	if err != nil {
		return nil, fmt.Errorf("error getting node group deploy policy while updating Kubernetes node group: %s", err)
	}

	req := &k8s.UpdateNodeGroupRequest{
		NodeGroupId:  d.Id(),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		NodeTemplate: tpl,
		ScalePolicy:  sp,
		Version: &k8s.UpdateVersionSpec{
			Specifier: &k8s.UpdateVersionSpec_Version{
				Version: d.Get("version").(string),
			},
		},
		MaintenancePolicy: mp,
		DeployPolicy:      dp,
		NodeLabels:        getNodeGroupNodeLabels(d),
		AllocationPolicy:  getNodeGroupAllocationPolicy(d),
	}

	return req, nil

}

func resourceYandexKubernetesNodeGroupDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ngID := d.Id()

	log.Printf("[DEBUG] Deleting Kubernetes node group %q", ngID)

	req := &k8s.DeleteNodeGroupRequest{
		NodeGroupId: ngID,
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Kubernetes().NodeGroup().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kubernetes node group %q", ngID))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Kubernetes node group %q", ngID)
	return nil
}

func flattenKubernetesNodeGroupTemplate(ngTpl *k8s.NodeTemplate) []map[string]interface{} {
	tpl := map[string]interface{}{
		"platform_id":               ngTpl.GetPlatformId(),
		"nat":                       ngTpl.GetV4AddressSpec().GetOneToOneNatSpec().GetIpVersion() == k8s.IpVersion_IPV4, //nolint
		"resources":                 flattenKubernetesNodeGroupTemplateResources(ngTpl.GetResourcesSpec()),
		"boot_disk":                 flattenKubernetesNodeGroupTemplateBootDisk(ngTpl.GetBootDiskSpec()),
		"metadata":                  ngTpl.GetMetadata(),
		"scheduling_policy":         flattenKubernetesNodeGroupTemplateSchedulingPolicy(ngTpl.GetSchedulingPolicy()),
		"network_interface":         flattenKubernetesNodeGroupNetworkInterfaces(ngTpl.GetNetworkInterfaceSpecs()),
		"placement_policy":          flattenKubernetesNodeGroupTemplatePlacementPolicy(ngTpl.GetPlacementPolicy()),
		"network_acceleration_type": strings.ToLower(ngTpl.GetNetworkSettings().GetType().String()),
		"container_runtime":         flattenKubernetesNodeGroupTemplateContainerRuntime(ngTpl.GetContainerRuntimeSettings()),
		"name":                      ngTpl.GetName(),
		"labels":                    ngTpl.GetLabels(),
		"container_network":         flattenKubernetesNodeGroupTemplateContainerNetwork(ngTpl.GetContainerNetworkSettings()),
		"gpu_settings":              flattenKubernetesNodeGroupTemplateGPUSettings(ngTpl.GetGpuSettings()),
	}

	return []map[string]interface{}{tpl}
}

func flattenKubernetesNodeGroupNetworkInterfaces(ifs []*k8s.NetworkInterfaceSpec) []map[string]interface{} {
	nifs := []map[string]interface{}{}
	for _, i := range ifs {
		nifs = append(nifs, flattenKubernetesNodeGroupNetworkInterface(i))
	}

	return nifs
}

func flattenKubernetesNodeGroupNetworkInterface(nif *k8s.NetworkInterfaceSpec) map[string]interface{} {
	res := map[string]interface{}{
		"subnet_ids":         nif.SubnetIds,
		"security_group_ids": nif.SecurityGroupIds,
		"nat":                flattenKubernetesNodeGroupNat(nif),
		"ipv4":               nif.PrimaryV4AddressSpec != nil,
		"ipv6":               nif.PrimaryV6AddressSpec != nil,
	}
	if nif.PrimaryV4AddressSpec != nil {
		res["ipv4_dns_records"] = flattenK8SNodeGroupDNSRecords(nif.GetPrimaryV4AddressSpec().GetDnsRecordSpecs())
	}
	if nif.PrimaryV6AddressSpec != nil {
		res["ipv6_dns_records"] = flattenK8SNodeGroupDNSRecords(nif.GetPrimaryV6AddressSpec().GetDnsRecordSpecs())
	}

	return res
}

func flattenKubernetesNodeGroupNat(nif *k8s.NetworkInterfaceSpec) bool {
	return nif.GetPrimaryV4AddressSpec().GetOneToOneNatSpec().GetIpVersion() == k8s.IpVersion_IPV4
}

func flattenKubernetesNodeGroupLocation(l *k8s.NodeGroupLocation) map[string]interface{} {
	return map[string]interface{}{
		"zone":      l.GetZoneId(),
		"subnet_id": l.GetSubnetId(),
	}
}

func flattenKubernetesNodeGroupMaintenancePolicy(mp *k8s.NodeGroupMaintenancePolicy) ([]map[string]interface{}, error) {
	mw, err := flattenMaintenanceWindow(mp.GetMaintenanceWindow())
	if err != nil {
		return nil, err
	}

	p := map[string]interface{}{
		"auto_upgrade":       mp.GetAutoUpgrade(),
		"auto_repair":        mp.GetAutoRepair(),
		"maintenance_window": mw,
	}

	return []map[string]interface{}{
		p,
	}, nil
}

func flattenKubernetesNodeGroupVersionInfo(vi *k8s.VersionInfo) []map[string]interface{} {
	info := map[string]interface{}{
		"current_version":        vi.GetCurrentVersion(),
		"new_revision_available": vi.GetNewRevisionAvailable(),
		"new_revision_summary":   vi.GetNewRevisionSummary(),
		"version_deprecated":     vi.GetVersionDeprecated(),
	}

	return []map[string]interface{}{
		info,
	}
}

func flattenKubernetesNodeGroupAllocationPolicy(ap *k8s.NodeGroupAllocationPolicy) []map[string]interface{} {
	locations := []map[string]interface{}{}
	for _, l := range ap.GetLocations() {
		locations = append(locations, flattenKubernetesNodeGroupLocation(l))
	}
	return []map[string]interface{}{
		{
			"location": locations,
		},
	}
}

func flattenKubernetesNodeScalePolicy(sp *k8s.ScalePolicy) []map[string]interface{} {
	if sp.GetFixedScale() != nil {
		return []map[string]interface{}{
			{
				"fixed_scale": []map[string]interface{}{
					{
						"size": sp.GetFixedScale().GetSize(),
					},
				},
			},
		}
	}
	return []map[string]interface{}{
		{
			"auto_scale": []map[string]interface{}{
				{
					"min":     sp.GetAutoScale().GetMinSize(),
					"max":     sp.GetAutoScale().GetMaxSize(),
					"initial": sp.GetAutoScale().GetInitialSize(),
				},
			},
		},
	}
}

func flattenKubernetesNodeGroupDeployPolicy(mp *k8s.DeployPolicy) ([]map[string]interface{}, error) {
	p := map[string]interface{}{
		"max_expansion":   mp.GetMaxExpansion(),
		"max_unavailable": mp.GetMaxUnavailable(),
	}

	return []map[string]interface{}{
		p,
	}, nil
}

func flattenKubernetesNodeGroupTemplateSchedulingPolicy(p *k8s.SchedulingPolicy) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"preemptible": p.GetPreemptible(),
		},
	}
}

func flattenKubernetesNodeGroupTemplatePlacementPolicy(p *k8s.PlacementPolicy) []map[string]interface{} {
	if p == nil {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{
		{
			"placement_group_id": p.PlacementGroupId,
		},
	}
}

func flattenKubernetesNodeGroupTemplateContainerRuntime(p *k8s.NodeTemplate_ContainerRuntimeSettings) []map[string]interface{} {
	if p == nil {
		return []map[string]interface{}{}
	}

	// TODO: if container_runtime is not explicitly specified on creation, then API returns
	// TYPE_UNSPECIFIED container runtime type. This type is not documented and should not be returned to end user.
	// Backend should fill container runtime info properly to avoid such situations (fix needed). For now
	// TYPE_UNSPECIFIED is ignored.
	if p.GetType() == k8s.NodeTemplate_ContainerRuntimeSettings_TYPE_UNSPECIFIED {
		return []map[string]interface{}{}
	}

	return []map[string]interface{}{
		{
			"type": strings.ToLower(p.GetType().String()),
		},
	}
}

func flattenKubernetesNodeGroupTemplateContainerNetwork(p *k8s.NodeTemplate_ContainerNetworkSettings) []map[string]interface{} {
	if p == nil {
		return []map[string]interface{}{}
	}
	return []map[string]interface{}{
		{
			"pod_mtu": int(p.GetPodMtu()),
		},
	}
}

func flattenKubernetesNodeGroupTemplateGPUSettings(p *k8s.GpuSettings) []map[string]interface{} {
	if p == nil {
		return []map[string]interface{}{}
	}
	gpuEnvironment := ""
	if p.GetGpuEnvironment() != k8s.GpuSettings_GPU_ENVIRONMENT_UNSPECIFIED {
		gpuEnvironment = strings.ToLower(p.GetGpuEnvironment().String())
	}
	return []map[string]interface{}{
		{
			"gpu_cluster_id":  p.GetGpuClusterId(),
			"gpu_environment": gpuEnvironment,
		},
	}
}

func flattenKubernetesNodeGroupTemplateBootDisk(d *k8s.DiskSpec) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"size": toGigabytes(d.GetDiskSize()),
			"type": d.GetDiskTypeId(),
		},
	}
}

func flattenKubernetesNodeGroupTemplateResources(r *k8s.ResourcesSpec) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"memory":        toGigabytesInFloat(r.GetMemory()),
			"cores":         int(r.GetCores()),
			"core_fraction": int(r.GetCoreFraction()),
			"gpus":          int(r.GetGpus()),
		},
	}
}
