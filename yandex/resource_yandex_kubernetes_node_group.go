package yandex

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

const (
	yandexKubernetesNodeGroupReadTimeout   = 10 * time.Minute
	yandexKubernetesNodeGroupCreateTimeout = 60 * time.Minute
	yandexKubernetesNodeGroupUpdateTimeout = 60 * time.Minute
	yandexKubernetesNodeGroupDeleteTimeout = 20 * time.Minute
)

func resourceYandexKubernetesNodeGroup() *schema.Resource {
	return &schema.Resource{
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
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_template": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"memory": {
										Type:         schema.TypeFloat,
										Optional:     true,
										Computed:     true,
										ValidateFunc: FloatAtLeast(0.0),
									},
									"cores": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"core_fraction": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"boot_disk": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(64),
									},
									"type": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
						"platform_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"nat": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
						"metadata": {
							Type:     schema.TypeMap,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
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
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"scale_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_scale": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
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
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"allocation_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"location": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:     schema.TypeString,
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
				},
			},
			"maintenance_policy": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auto_upgrade": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"auto_repair": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						//"maintenance_window": {
						//	Type:     schema.TypeString,
						//	Computed: true,
						//},
					},
				},
			},
			"instance_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_info": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
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
		return nil, fmt.Errorf("error getting node template while creating Kubernetes node group: %s", err)
	}

	req := &k8s.CreateNodeGroupRequest{
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		Labels:           labels,
		ClusterId:        d.Get("cluster_id").(string),
		NodeTemplate:     tpl,
		ScalePolicy:      getNodeGroupScalePolicy(d),
		AllocationPolicy: getNodeGroupAllocationPolicy(d),
		Version:          d.Get("version").(string),
	}

	return req, nil
}

func getNodeGroupAllocationPolicy(d *schema.ResourceData) *k8s.NodeGroupAllocationPolicy {
	return &k8s.NodeGroupAllocationPolicy{
		Locations: getNodeGroupAllocationPolicyLocations(d),
	}
}

func getNodeGroupAllocationPolicyLocations(d *schema.ResourceData) []*k8s.NodeGroupLocation {
	var locations []*k8s.NodeGroupLocation
	h := schemaHelper(d, "allocation_policy.0.location.")
	locationCount := h.GetInt("#")
	for i := 0; i < locationCount; i++ {
		location := h.Get(fmt.Sprintf("%d", i)).(map[string]interface{})
		locationSpec := &k8s.NodeGroupLocation{}

		if zone, ok := location["zone"]; ok {
			locationSpec.ZoneId = zone.(string)
		}

		if subnet, ok := location["subnet_id"]; ok {
			locationSpec.SubnetId = subnet.(string)
		}

		locations = append(locations, locationSpec)
	}
	return locations

}

func getNodeGroupScalePolicy(d *schema.ResourceData) *k8s.ScalePolicy {
	if size, ok := d.GetOk("scale_policy.0.fixed_scale.0.size"); ok {
		return &k8s.ScalePolicy{
			ScaleType: &k8s.ScalePolicy_FixedScale_{
				FixedScale: &k8s.ScalePolicy_FixedScale{
					Size: int64(size.(int)),
				},
			},
		}
	}

	return nil
}

func getNodeGroupTemplate(d *schema.ResourceData) (*k8s.NodeTemplate, error) {
	h := schemaHelper(d, "instance_template.0.")

	metadata, err := expandLabels(h.Get("metadata"))
	if err != nil {
		return nil, fmt.Errorf("error expanding metadata while creating Kubernetes node group: %s", err)
	}

	tpl := &k8s.NodeTemplate{
		PlatformId:       h.GetString("platform_id"),
		ResourcesSpec:    getNodeGroupResourceSpec(d),
		BootDiskSpec:     getNodeGroupBootDiskSpec(d),
		Metadata:         metadata,
		V4AddressSpec:    getNodeGroupAddressSpec(d),
		SchedulingPolicy: getNodeGroupTemplateSchedulingPolicy(d),
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
	}
	return spec
}

func flattenNodeGroupSchemaData(ng *k8s.NodeGroup, d *schema.ResourceData) error {
	createdAt, err := getTimestamp(ng.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("cluster_id", ng.ClusterId)
	d.Set("created_at", createdAt)
	d.Set("name", ng.Name)
	d.Set("description", ng.Description)
	d.Set("status", strings.ToLower(ng.Status.String()))

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

	maintenancePolicy := flattenKubernetesNodeGroupMaintenancePolicy(ng.GetMaintenancePolicy())
	if err := d.Set("maintenance_policy", maintenancePolicy); err != nil {
		return err
	}

	d.SetId(ng.Id)
	return nil
}

var nodeGroupUpdateFieldsMap = map[string]string{
	"name":                                                "name",
	"description":                                         "description",
	"labels":                                              "labels",
	"instance_template.0.platform_id":                     "node_template.platform_id",
	"instance_template.0.metadata":                        "node_template.metadata",
	"instance_template.0.resources.0.memory":              "node_template.resources_spec.memory",
	"instance_template.0.resources.0.cores":               "node_template.resources_spec.cores",
	"instance_template.0.resources.0.core_fraction":       "node_template.resources_spec.core_fraction",
	"instance_template.0.boot_disk.0.type":                "node_template.boot_disk_spec.disk_type_id",
	"instance_template.0.boot_disk.0.size":                "node_template.boot_disk_spec.disk_size",
	"instance_template.0.scheduling_policy.0.preemptible": "node_template.scheduling_policy.preemptible",
	"scale_policy.0.fixed_scale.0.size":                   "scale_policy",
	"version":                                             "version",
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
		return nil, fmt.Errorf("error getting node template while updating Kubernetes node group: %s", err)
	}

	req := &k8s.UpdateNodeGroupRequest{
		NodeGroupId:  d.Id(),
		Name:         d.Get("name").(string),
		Description:  d.Get("description").(string),
		Labels:       labels,
		NodeTemplate: tpl,
		ScalePolicy:  getNodeGroupScalePolicy(d),
		Version: &k8s.UpdateVersionSpec{
			Specifier: &k8s.UpdateVersionSpec_Version{
				Version: d.Get("version").(string),
			},
		},
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
		"platform_id":       ngTpl.GetPlatformId(),
		"nat":               ngTpl.GetV4AddressSpec().GetOneToOneNatSpec().GetIpVersion() == k8s.IpVersion_IPV4,
		"resources":         flattenKubernetesNodeGroupTemplateResources(ngTpl.GetResourcesSpec()),
		"boot_disk":         flattenKubernetesNodeGroupTemplateBootDisk(ngTpl.GetBootDiskSpec()),
		"metadata":          ngTpl.GetMetadata(),
		"scheduling_policy": flattenKubernetesNodeGroupTemplateSchedulingPolicy(ngTpl.GetSchedulingPolicy()),
	}

	return []map[string]interface{}{tpl}
}

func flattenKubernetesNodeGroupLocation(l *k8s.NodeGroupLocation) map[string]interface{} {
	return map[string]interface{}{
		"zone":      l.GetZoneId(),
		"subnet_id": l.GetSubnetId(),
	}
}

func flattenKubernetesNodeGroupMaintenancePolicy(mp *k8s.NodeGroupMaintenancePolicy) []map[string]interface{} {
	p := map[string]interface{}{
		"auto_upgrade": mp.GetAutoUpgrade(),
		"auto_repair":  mp.GetAutoRepair(),
		//"maintenance_window": mp.GetMaintenanceWindow(),
	}

	return []map[string]interface{}{
		p,
	}
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

func flattenKubernetesNodeGroupTemplateSchedulingPolicy(p *k8s.SchedulingPolicy) []map[string]interface{} {
	return []map[string]interface{}{
		{
			"preemptible": p.GetPreemptible(),
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
		},
	}
}
