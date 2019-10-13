package yandex

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
)

const (
	yandexKubernetesClusterCreateTimeout  = 15 * time.Minute
	yandexKubernetesClusterDefaultTimeout = 5 * time.Minute
)

func resourceYandexKubernetesCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexKubernetesClusterCreate,
		Read:   resourceYandexKubernetesClusterRead,
		Update: resourceYandexKubernetesClusterUpdate,
		Delete: resourceYandexKubernetesClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexKubernetesClusterCreateTimeout),
			Read:   schema.DefaultTimeout(yandexKubernetesClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexKubernetesClusterDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexKubernetesClusterDefaultTimeout),
		},
		Schema: map[string]*schema.Schema{
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"node_service_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"master": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"zonal": {
							Type:          schema.TypeList,
							Computed:      true,
							Optional:      true,
							ForceNew:      true,
							ConflictsWith: []string{"master.0.regional"},
							MaxItems:      1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"zone": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
								},
							},
						},
						"regional": {
							Type:          schema.TypeList,
							MaxItems:      1,
							Optional:      true,
							Computed:      true,
							ForceNew:      true,
							ConflictsWith: []string{"master.0.zonal"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"region": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
										ForceNew: true,
									},
									"location": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										ForceNew: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"zone": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"subnet_id": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"internal_v4_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"external_v4_address": {
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
						"cluster_ca_certificate": {
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
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				Computed: true,
			},
			"cluster_ipv4_range": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"service_ipv4_range": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"release_channel": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
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
		},
	}
}

func resourceYandexKubernetesClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateKubernetesClusterRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Kubernetes().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Kubernetes cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while get Kubernetes cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*k8s.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Kubernetes cluster ID from create operation metadata")
	}

	d.SetId(md.GetClusterId())

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting operation to create Kubernetes cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Kubernetes cluster creation failed: %s", err)
	}

	return resourceYandexKubernetesClusterRead(d, meta)
}

func resourceYandexKubernetesClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	clusterID := d.Id()

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.Kubernetes().Cluster().Get(ctx, &k8s.GetClusterRequest{
		ClusterId: clusterID,
	})

	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kubernetes cluster with ID %q", clusterID))
	}

	return flattenKubernetesClusterAttributes(cluster, d, true)
}

var updateKubernetesClusterFieldsMap = map[string]string{
	"name":                    "name",
	"description":             "description",
	"labels":                  "labels",
	"service_account_id":      "service_account_id",
	"node_service_account_id": "node_service_account_id",
}

func resourceYandexKubernetesClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	clusterID := d.Id()
	log.Printf("[DEBUG] updating Kubernetes cluster %q", clusterID)

	req, err := getKubernetesClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	var updatePath []string
	for field, path := range updateKubernetesClusterFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	if len(updatePath) == 0 {
		return fmt.Errorf("error while updating Kubernetes cluster, didn't detect any changes")
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Kubernetes().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update Kubernetes cluster %q: %s", clusterID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error updating Kubernetes cluster %q: %s", clusterID, err)
	}

	return resourceYandexKubernetesClusterRead(d, meta)
}

func getKubernetesClusterUpdateRequest(d *schema.ResourceData) (*k8s.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating Kubernetes cluster: %s", err)
	}

	req := &k8s.UpdateClusterRequest{
		ClusterId:            d.Id(),
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Labels:               labels,
		ServiceAccountId:     d.Get("service_account_id").(string),
		NodeServiceAccountId: d.Get("node_service_account_id").(string),
	}

	return req, nil

}

func resourceYandexKubernetesClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	clusterID := d.Id()

	log.Printf("[DEBUG] Deleting Kubernetes cluster %q", d.Id())

	req := &k8s.DeleteClusterRequest{
		ClusterId: clusterID,
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Kubernetes().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Kubernetes cluster %q", clusterID))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Kubernetes cluster %q", d.Id())
	return nil
}

func prepareCreateKubernetesClusterRequest(d *schema.ResourceData, meta *Config) (*k8s.CreateClusterRequest, error) {
	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting folder ID while creating Kubernetes cluster: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while creating Kubernetes cluster: %s", err)
	}

	masterSpec, err := getKubernetesClusterMasterSpec(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting master spec while creating Kubernetes cluster: %s", err)
	}
	releaseChannel, err := getKubernetesClusterReleaseChannel(d)
	if err != nil {
		return nil, err
	}

	req := &k8s.CreateClusterRequest{
		FolderId:             folderID,
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Labels:               labels,
		NetworkId:            d.Get("network_id").(string),
		MasterSpec:           masterSpec,
		IpAllocationPolicy:   getIPAllocationPolicy(d),
		ServiceAccountId:     d.Get("service_account_id").(string),
		NodeServiceAccountId: d.Get("node_service_account_id").(string),
		ReleaseChannel:       releaseChannel,
	}

	return req, nil
}

func getIPAllocationPolicy(d *schema.ResourceData) *k8s.IPAllocationPolicy {
	p := &k8s.IPAllocationPolicy{
		ClusterIpv4CidrBlock: d.Get("cluster_ipv4_range").(string),
		ServiceIpv4CidrBlock: d.Get("service_ipv4_range").(string),
	}

	return p
}

func getKubernetesClusterReleaseChannels() string {
	var values []string
	for k := range k8s.ReleaseChannel_value {
		values = append(values, k)
	}
	sort.Strings(values)

	return strings.Join(values, ",")
}

func getKubernetesClusterReleaseChannel(d *schema.ResourceData) (k8s.ReleaseChannel, error) {
	c, ok := d.GetOk("release_channel")
	if ok {
		if ch, ok := k8s.ReleaseChannel_value[c.(string)]; ok {
			return k8s.ReleaseChannel(ch), nil
		}

		err := fmt.Errorf("invalid release_channel field value, possible values: %v", getKubernetesClusterReleaseChannels())
		return k8s.ReleaseChannel_RELEASE_CHANNEL_UNSPECIFIED, err
	}

	return k8s.ReleaseChannel_RELEASE_CHANNEL_UNSPECIFIED, nil
}

func getKubernetesClusterMasterSpec(d *schema.ResourceData, meta *Config) (*k8s.MasterSpec, error) {
	spec := &k8s.MasterSpec{
		Version:    d.Get("master.0.version").(string),
		MasterType: nil,
	}

	if _, ok := d.GetOk("master.0.zonal"); ok {
		spec.MasterType = getKubernetesClusterZonalMaster(d, meta)
		return spec, nil
	}

	if _, ok := d.GetOk("master.0.regional"); ok {
		spec.MasterType = getKubernetesClusterRegionalMaster(d, meta)
		return spec, nil
	}

	return nil, fmt.Errorf("either zonal or regional master should be specified for Kubernetes cluster")

}

func getKubernetesClusterZonalMaster(d *schema.ResourceData, meta *Config) *k8s.MasterSpec_ZonalMasterSpec {
	return &k8s.MasterSpec_ZonalMasterSpec{
		ZonalMasterSpec: &k8s.ZonalMasterSpec{
			ZoneId:                getZonalMasterZone(d, meta),
			InternalV4AddressSpec: getZonalMasterInternalAddressSpec(d),
			ExternalV4AddressSpec: getZonalMasterExternalAddressSpec(d),
		},
	}
}

func getKubernetesClusterRegionalMaster(d *schema.ResourceData, _ *Config) *k8s.MasterSpec_RegionalMasterSpec {
	return &k8s.MasterSpec_RegionalMasterSpec{
		RegionalMasterSpec: &k8s.RegionalMasterSpec{
			RegionId:              d.Get("master.0.regional.0.region").(string),
			Locations:             getKubernetesClusterRegionalMasterLocations(d),
			ExternalV4AddressSpec: getZonalMasterExternalAddressSpec(d),
		},
	}
}

func getKubernetesClusterRegionalMasterLocations(d *schema.ResourceData) []*k8s.MasterLocation {
	var locations []*k8s.MasterLocation
	locationCount := d.Get("master.0.regional.0.location.#").(int)
	for i := 0; i < locationCount; i++ {
		location := d.Get(fmt.Sprintf("master.0.regional.0.location.%d", i)).(map[string]interface{})
		locationSpec := &k8s.MasterLocation{}

		if zone, ok := location["zone"]; ok {
			locationSpec.ZoneId = zone.(string)
		}

		if subnet, ok := location["subnet_id"]; ok {
			locationSpec.InternalV4AddressSpec = &k8s.InternalAddressSpec{
				SubnetId: subnet.(string),
			}
		}

		locations = append(locations, locationSpec)
	}
	return locations
}

func getZonalMasterZone(d *schema.ResourceData, config *Config) string {
	res, ok := d.GetOk("master.0.zonal.0.zone")
	if !ok {
		return config.Zone
	}
	return res.(string)
}

func getZonalMasterInternalAddressSpec(d *schema.ResourceData) *k8s.InternalAddressSpec {
	res, ok := d.GetOk("master.0.zonal.0.subnet_id")
	if ok {
		return &k8s.InternalAddressSpec{
			SubnetId: res.(string),
		}
	}
	return nil
}

func getZonalMasterExternalAddressSpec(d *schema.ResourceData) *k8s.ExternalAddressSpec {
	publicIP, ok := d.GetOk("master.0.public_ip")
	if ok && publicIP.(bool) {
		return &k8s.ExternalAddressSpec{}
	}

	return nil
}

func flattenKubernetesClusterAttributes(cluster *k8s.Cluster, d *schema.ResourceData, clusterResource bool) error {
	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("folder_id", cluster.FolderId)
	d.Set("name", cluster.Name)
	d.Set("description", cluster.Description)
	d.Set("status", strings.ToLower(cluster.Status.String()))
	d.Set("health", strings.ToLower(cluster.Health.String()))
	d.Set("network_id", cluster.NetworkId)
	d.Set("service_account_id", cluster.ServiceAccountId)
	d.Set("node_service_account_id", cluster.NodeServiceAccountId)
	d.Set("release_channel", cluster.ReleaseChannel.String())

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	h, err := flattenKubernetesMaster(cluster)
	if err != nil {
		return err
	}

	if clusterResource {
		h.fillClusterMasterResourceFields(cluster, d)
	} else {
		d.Set("cluster_id", cluster.Id)
	}

	err = d.Set("master", h.schema())
	if err != nil {
		return err
	}

	d.SetId(cluster.Id)
	return nil
}

type masterSchemaHelper struct {
	zonalMaster    map[string]interface{}
	regionalMaster map[string]interface{}
	versionInfo    map[string]interface{}
	master         map[string]interface{}
}

func constructKubernetesMasterSchemaHelper() *masterSchemaHelper {
	helper := &masterSchemaHelper{}
	helper.versionInfo = map[string]interface{}{}
	helper.master = map[string]interface{}{
		"version_info": []map[string]interface{}{
			helper.versionInfo,
		},
	}
	return helper
}

func (h *masterSchemaHelper) schema() []map[string]interface{} {
	return []map[string]interface{}{h.master}
}

func (h *masterSchemaHelper) fillClusterMasterResourceFields(cluster *k8s.Cluster, d *schema.ResourceData) {
	clusterMaster := cluster.GetMaster()
	h.master["version"] = clusterMaster.GetVersion()
	h.master["public_ip"] = clusterMaster.GetEndpoints().GetExternalV4Endpoint() != ""

	if subnet, ok := d.GetOk("master.0.zonal.0.subnet_id"); ok {
		h.getZonalMaster()["subnet_id"] = subnet
	}

	if region, ok := d.GetOk("master.0.regional.0.region"); ok {
		h.getRegionalMaster()["region"] = region
	}

	if locations, ok := d.GetOk("master.0.regional.0.location"); ok {
		h.getRegionalMaster()["location"] = locations
	}
}

func (h *masterSchemaHelper) getZonalMaster() map[string]interface{} {
	if h.zonalMaster == nil {
		h.zonalMaster = map[string]interface{}{}
		h.master["zonal"] = []map[string]interface{}{
			h.zonalMaster,
		}
	}

	return h.zonalMaster
}

func (h *masterSchemaHelper) getRegionalMaster() map[string]interface{} {
	if h.regionalMaster == nil {
		h.regionalMaster = map[string]interface{}{}
		h.master["regional"] = []map[string]interface{}{
			h.regionalMaster,
		}
	}

	return h.regionalMaster
}

func (h *masterSchemaHelper) flattenClusterZonalMaster(m *k8s.Master_ZonalMaster) {
	h.master["internal_v4_address"] = m.ZonalMaster.GetInternalV4Address()
	h.master["external_v4_address"] = m.ZonalMaster.GetExternalV4Address()

	h.getZonalMaster()["zone"] = m.ZonalMaster.GetZoneId()
}

func (h *masterSchemaHelper) flattenClusterRegionalMaster(m *k8s.Master_RegionalMaster) {
	h.master["internal_v4_address"] = m.RegionalMaster.GetInternalV4Address()
	h.master["external_v4_address"] = m.RegionalMaster.GetExternalV4Address()

	h.getRegionalMaster()["region"] = m.RegionalMaster.GetRegionId()
}

func flattenKubernetesMaster(cluster *k8s.Cluster) (*masterSchemaHelper, error) {
	h := constructKubernetesMasterSchemaHelper()
	clusterMaster := cluster.GetMaster()
	if clusterMaster == nil {
		return nil, fmt.Errorf("failed to get cluster master spec")
	}

	h.master["internal_v4_endpoint"] = clusterMaster.GetEndpoints().GetInternalV4Endpoint()
	h.master["external_v4_endpoint"] = clusterMaster.GetEndpoints().GetExternalV4Endpoint()
	h.master["cluster_ca_certificate"] = clusterMaster.GetMasterAuth().GetClusterCaCertificate()

	switch m := clusterMaster.GetMasterType().(type) {
	case *k8s.Master_ZonalMaster:
		h.flattenClusterZonalMaster(m)
	case *k8s.Master_RegionalMaster:
		h.flattenClusterRegionalMaster(m)
	default:
		return nil, fmt.Errorf("unsupported Kubernetes master type (currently only zonal and regional master are supported)")
	}

	versionInfo := clusterMaster.GetVersionInfo()
	if versionInfo == nil {
		return nil, fmt.Errorf("failed to get Kubernetes master version info")
	}

	h.versionInfo["current_version"] = versionInfo.GetCurrentVersion()
	h.versionInfo["new_revision_available"] = versionInfo.GetNewRevisionAvailable()
	h.versionInfo["new_revision_summary"] = versionInfo.GetNewRevisionSummary()
	h.versionInfo["version_deprecated"] = versionInfo.GetVersionDeprecated()
	return h, nil
}
