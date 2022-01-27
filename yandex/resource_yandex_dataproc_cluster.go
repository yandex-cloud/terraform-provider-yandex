package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/dataproc/v1"
)

const (
	yandexDataprocClusterCreateTimeout = 60 * time.Minute
	yandexDataprocClusterDeleteTimeout = 60 * time.Minute
	yandexDataprocClusterUpdateTimeout = 60 * time.Minute
)

func isVersionPrefix(prefix string, version string) bool {
	prefixParts := strings.Split(prefix, ".")
	versionParts := strings.Split(version, ".")
	if len(prefixParts) > len(versionParts) {
		return false
	}
	for i, value := range prefixParts {
		if value != versionParts[i] {
			return false
		}
	}
	return true
}

func resourceYandexDataprocCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexDataprocClusterCreate,
		Read:   resourceYandexDataprocClusterRead,
		Update: resourceYandexDataprocClusterUpdate,
		Delete: resourceYandexDataprocClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexDataprocClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexDataprocClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexDataprocClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			"cluster_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subcluster_spec": {
							Type:     schema.TypeList,
							MinItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},

									"role": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice(
											[]string{"MASTERNODE", "DATANODE", "COMPUTENODE"}, false),
									},

									"resources": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_preset_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"disk_size": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"disk_type_id": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
													Default:  "network-hdd",
												},
											},
										},
									},

									"autoscaling_config": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_hosts_count": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"preemptible": {
													Type:     schema.TypeBool,
													Optional: true,
													Default:  false,
												},
												"measurement_duration": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  -1,
												},
												"warmup_duration": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  -1,
												},
												"stabilization_duration": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  -1,
												},
												"cpu_utilization_target": {
													Type:     schema.TypeFloat,
													Optional: true,
													Default:  -1,
												},
												"decommission_timeout": {
													Type:     schema.TypeInt,
													Optional: true,
													Default:  -1,
												},
											},
										},
									},

									"subnet_id": {
										Type:     schema.TypeString,
										Required: true,
									},

									"hosts_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},

									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"hadoop": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"services": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(dataprocServiceNames(), false),
										},
									},

									"properties": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},

									"ssh_public_keys": {
										Type:     schema.TypeSet,
										Optional: true,
										ForceNew: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"version_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return isVersionPrefix(new, old)
							},
						},
					},
				},
			},

			"bucket": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"ui_proxy": {
				Type:     schema.TypeBool,
				Optional: true,
			},

			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},

			"host_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
				ForceNew: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceYandexDataprocClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareDataprocCreateClusterRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Data Proc Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while getting Data Proc Cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*dataproc.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Data Proc Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create Data Proc Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("failed to create Data Proc Cluster: %s", err)
	}

	return resourceYandexDataprocClusterRead(d, meta)
}

func resourceYandexDataprocClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.Dataproc().Cluster().Get(ctx, &dataproc.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("cluster %q", d.Id()))
	}

	return populateDataprocClusterResourceData(d, config, cluster)
}

func populateDataprocClusterResourceData(d *schema.ResourceData, config *Config, cluster *dataproc.Cluster) error {
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	subclusters, err := listDataprocSubclusters(ctx, config, cluster.Id)
	if err != nil {
		return err
	}
	subclusters = reorderDataprocSubclusters(d, subclusters)

	if err := d.Set("folder_id", cluster.FolderId); err != nil {
		return err
	}

	if err := d.Set("name", cluster.Name); err != nil {
		return err
	}

	if err := d.Set("description", cluster.Description); err != nil {
		return err
	}

	if err := d.Set("zone_id", cluster.ZoneId); err != nil {
		return err
	}

	if err := d.Set("service_account_id", cluster.ServiceAccountId); err != nil {
		return err
	}

	if err := d.Set("bucket", cluster.Bucket); err != nil {
		return err
	}

	if err := d.Set("ui_proxy", cluster.UiProxy); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	if err := d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if err := d.Set("cluster_config", flattenDataprocClusterConfig(cluster, subclusters)); err != nil {
		return err
	}

	if err := d.Set("created_at", getTimestamp(cluster.CreatedAt)); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	return nil
}

func reorderDataprocSubclusters(d *schema.ResourceData, subclusters []*dataproc.Subcluster) []*dataproc.Subcluster {
	subclustersReordered := make([]*dataproc.Subcluster, len(subclusters))
	subclusterSpecs := d.Get("cluster_config.0.subcluster_spec").([]interface{})

	for i, value := range subclusterSpecs {
		subclusterSpec := value.(map[string]interface{})
		var id, name string
		if val, ok := subclusterSpec["id"]; ok {
			id = val.(string)
		}
		if val, ok := subclusterSpec["name"]; ok {
			name = val.(string)
		}
		for j, subcluster := range subclusters {
			if subcluster != nil && (subcluster.Id == id || subcluster.Name == name) {
				subclustersReordered[i] = subcluster
				subclusters[j] = nil
				break
			}
		}
	}

	j := 0
	for i, s := range subclustersReordered {
		if s == nil {
			subcluster := (*dataproc.Subcluster)(nil)
			for ; subcluster == nil; j++ {
				subcluster = subclusters[j]
			}
			subclustersReordered[i] = subcluster
		}
	}

	return subclustersReordered
}

func getDataprocZoneID(d *schema.ResourceData, config *Config) (string, error) {
	res, ok := d.GetOk("zone_id")
	if !ok {
		if config.Zone != "" {
			return config.Zone, nil
		}
		return "", fmt.Errorf("cannot determine zone: please set 'zone_id' key in this resource or at provider level")
	}
	return res.(string), nil
}

func prepareDataprocCreateClusterRequest(d *schema.ResourceData, meta *Config) (*dataproc.CreateClusterRequest, error) {
	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting folder ID while creating Data Proc Cluster: %s", err)
	}

	zoneID, err := getDataprocZoneID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting zone while creating Data Proc Cluster: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Data Proc Cluster create: %s", err)
	}

	req := dataproc.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		ConfigSpec:         expandDataprocCreateClusterConfigSpec(d),
		ZoneId:             zoneID,
		ServiceAccountId:   d.Get("service_account_id").(string),
		Bucket:             d.Get("bucket").(string),
		UiProxy:            d.Get("ui_proxy").(bool),
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		HostGroupIds:       expandHostGroupIds(d.Get("host_group_ids")),
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	return &req, nil
}

func listDataprocSubclusters(ctx context.Context, config *Config, id string) ([]*dataproc.Subcluster, error) {
	var subclusters []*dataproc.Subcluster
	pageToken := ""
	for {
		resp, err := config.sdk.Dataproc().Subcluster().List(ctx, &dataproc.ListSubclustersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of subclusters for Data Proc Cluster %q: %s", id, err)
		}
		subclusters = append(subclusters, resp.Subclusters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return subclusters, nil
}

func resourceYandexDataprocClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Data Proc Cluster %q", d.Id())

	req := &dataproc.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Data Proc Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Data Proc Cluster %q", d.Id())
	return nil
}

func resourceYandexDataprocClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Data Proc Cluster %q", d.Id())

	d.Partial(true)

	if err := updateDataprocClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("cluster_config.0.subcluster_spec") {
		if err := updateDataprocSubclusters(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)

	log.Printf("[DEBUG] Finished updating Data Proc Cluster %q", d.Id())
	return resourceYandexDataprocClusterRead(d, meta)
}

func updateDataprocClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := getDataprocClusterUpdateRequest(d)
	if err != nil {
		return err
	}
	if req == nil {
		return nil
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update Data Proc Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating Data Proc Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func getDataprocClusterUpdateRequest(d *schema.ResourceData) (*dataproc.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Data Proc Cluster update: %s", err)
	}

	req := &dataproc.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Description:        d.Get("description").(string),
		Labels:             labels,
		Name:               d.Get("name").(string),
		ServiceAccountId:   d.Get("service_account_id").(string),
		Bucket:             d.Get("bucket").(string),
		UiProxy:            d.Get("ui_proxy").(bool),
		SecurityGroupIds:   expandSecurityGroupIds(d.Get("security_group_ids")),
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	var updatePaths []string
	fieldNames := []string{"description", "labels", "name", "service_account_id", "bucket", "ui_proxy", "security_group_ids", "deletion_protection"}
	for _, fieldName := range fieldNames {
		if d.HasChange(fieldName) {
			updatePaths = append(updatePaths, fieldName)
		}
	}

	propertiesPath := "cluster_config.0.hadoop.0.properties"
	if d.HasChange(propertiesPath) {
		req.ConfigSpec = &dataproc.UpdateClusterConfigSpec{
			Hadoop: &dataproc.HadoopConfig{
				Properties: convertStringMap(d.Get(propertiesPath).(map[string]interface{})),
			},
		}
		updatePaths = append(updatePaths, "config_spec.hadoop.properties")
	}

	if len(updatePaths) == 0 {
		return nil, nil
	}
	req.UpdateMask = &field_mask.FieldMask{Paths: updatePaths}
	return req, nil
}

func updateDataprocSubclusters(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	subclusters, err := listDataprocSubclusters(ctx, config, d.Id())
	if err != nil {
		return err
	}

	createReqs, updateReqs, deleteReqs, err := partitionDataprocSubclustersByAction(d, subclusters)
	if err != nil {
		return err
	}

	for _, deleteReq := range deleteReqs {
		err := deleteDataprocSubcluster(deleteReq, config, d.Timeout(schema.TimeoutDelete))
		if err != nil {
			return err
		}
	}

	for _, createReq := range createReqs {
		err := createDataprocSubcluster(createReq, config, d.Timeout(schema.TimeoutCreate))
		if err != nil {
			return err
		}
	}

	for _, updateReq := range updateReqs {
		err := updateDataprocSubcluster(updateReq, config, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return err
		}

	}

	return nil
}

func partitionDataprocSubclustersByAction(d *schema.ResourceData, subclusters []*dataproc.Subcluster) ([]*dataproc.CreateSubclusterRequest,
	[]*dataproc.UpdateSubclusterRequest, []*dataproc.DeleteSubclusterRequest, error) {
	subclusterSpecs := d.Get("cluster_config.0.subcluster_spec").([]interface{})
	updateReqs := make([]*dataproc.UpdateSubclusterRequest, 0, len(subclusters))
	createReqs := make([]*dataproc.CreateSubclusterRequest, 0, len(subclusters))
	clusterID := d.Id()

	subclusterByID := make(map[string]*dataproc.Subcluster)
	for _, subcluster := range subclusters {
		subclusterByID[subcluster.Id] = subcluster
	}

	for i, subclusterSpec := range subclusterSpecs {
		var id string
		if val, ok := subclusterSpec.(map[string]interface{})["id"]; ok {
			id = val.(string)
		}
		if _, exist := subclusterByID[id]; exist {
			delete(subclusterByID, id)
			path := fmt.Sprintf("cluster_config.0.subcluster_spec.%d", i)
			req, err := getDataprocSubclusterUpdateRequest(d, path)
			if err != nil {
				return nil, nil, nil, err
			}
			if req != nil {
				updateReqs = append(updateReqs, req)
			}
		} else {
			req := getDataprocSubclusterCreateRequest(clusterID, subclusterSpec)
			createReqs = append(createReqs, req)
		}
	}

	deleteReqs := make([]*dataproc.DeleteSubclusterRequest, 0, len(subclusters))
	for _, subcluster := range subclusterByID {
		req := &dataproc.DeleteSubclusterRequest{
			ClusterId:    clusterID,
			SubclusterId: subcluster.Id,
		}
		deleteReqs = append(deleteReqs, req)
	}

	return createReqs, updateReqs, deleteReqs, nil
}

func getDataprocSubclusterCreateRequest(clusterID string, element interface{}) *dataproc.CreateSubclusterRequest {
	createSpec := expandDataprocSubclusterSpec(element)

	return &dataproc.CreateSubclusterRequest{
		ClusterId:         clusterID,
		Name:              createSpec.Name,
		Role:              createSpec.Role,
		Resources:         createSpec.Resources,
		SubnetId:          createSpec.SubnetId,
		HostsCount:        createSpec.HostsCount,
		AutoscalingConfig: createSpec.AutoscalingConfig,
	}
}

func getDataprocSubclusterUpdateRequest(d *schema.ResourceData, path string) (*dataproc.UpdateSubclusterRequest, error) {
	subclusterSpec := d.Get(path).(map[string]interface{})
	resourcesSpec := subclusterSpec["resources"].([]interface{})[0]

	req := &dataproc.UpdateSubclusterRequest{
		ClusterId:    d.Id(),
		SubclusterId: subclusterSpec["id"].(string),
		Resources:    expandDataprocResources(resourcesSpec),
		Name:         subclusterSpec["name"].(string),
		HostsCount:   int64(subclusterSpec["hosts_count"].(int)),
	}
	autoscalingConfigs := subclusterSpec["autoscaling_config"].([]interface{})
	if len(autoscalingConfigs) > 0 {
		req.AutoscalingConfig = expandDataprocAutoscalingConfig(autoscalingConfigs[0])
	}

	constFields := []string{"role", "subnet_id"}
	for _, fieldName := range constFields {
		field := path + "." + fieldName
		if d.HasChange(field) {
			return nil, fmt.Errorf("error while trying to update Data Proc Subcluster %q:"+
				" changing %q of existing subcluster is not supported", req.SubclusterId, fieldName)
		}
	}

	var updatePaths []string
	fieldNames := []string{
		"resources",
		"name",
		"hosts_count",
	}
	for _, fieldName := range fieldNames {
		field := path + "." + fieldName
		if d.HasChange(field) {
			updatePaths = append(updatePaths, fieldName)
		}
	}

	// later resources also can be added here
	structureFieldNames := map[string]string{
		"max_hosts_count":        "autoscaling_config",
		"preemptible":            "autoscaling_config",
		"warmup_duration":        "autoscaling_config",
		"stabilization_duration": "autoscaling_config",
		"measurement_duration":   "autoscaling_config",
		"cpu_utilization_target": "autoscaling_config",
		"decommission_timeout":   "autoscaling_config",
	}

	for fieldName, structureName := range structureFieldNames {
		field := path + "." + structureName + ".0." + fieldName
		if d.HasChange(field) {
			updatePaths = append(updatePaths, structureName+"."+fieldName)
		}
	}

	log.Printf("[DEBUG] fieldMask = %s", updatePaths)
	if len(updatePaths) == 0 {
		return nil, nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePaths}
	return req, nil
}

func deleteDataprocSubcluster(deleteReq *dataproc.DeleteSubclusterRequest, config *Config, timeout time.Duration) error {
	log.Printf("[DEBUG] Deleting Data Proc Subcluster %q", deleteReq.SubclusterId)

	ctx, cancel := config.ContextWithTimeout(timeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Subcluster().Delete(ctx, deleteReq))
	if err != nil {
		return fmt.Errorf("error while requesting API to delete Data Proc Subcluster %q: %s", deleteReq.SubclusterId, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting Data Proc Subcluster %q: %s", deleteReq.SubclusterId, err)
	}

	log.Printf("[DEBUG] Deleted Data Proc Subcluster %q", deleteReq.SubclusterId)
	return nil
}

func createDataprocSubcluster(createReq *dataproc.CreateSubclusterRequest, config *Config, timeout time.Duration) error {
	log.Printf("[DEBUG] Creating Data Proc Subcluster %q", createReq.Name)

	ctx, cancel := config.ContextWithTimeout(timeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Subcluster().Create(ctx, createReq))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Data Proc Subcluster %q: %s", createReq.Name, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating Data Proc Subcluster %q: %s", createReq.Name, err)
	}

	log.Printf("[DEBUG] Created Data Proc Subcluster %q", createReq.Name)
	return nil
}

func updateDataprocSubcluster(updateReq *dataproc.UpdateSubclusterRequest, config *Config, timeout time.Duration) error {
	log.Printf("[DEBUG] Updating Data Proc Subcluster %q", updateReq.SubclusterId)

	ctx, cancel := config.ContextWithTimeout(timeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Subcluster().Update(ctx, updateReq))
	if err != nil {
		return fmt.Errorf("error while requesting API to update Data Proc Subcluster %q: %s", updateReq.SubclusterId, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating Data Proc Subcluster %q: %s", updateReq.SubclusterId, err)
	}

	log.Printf("[DEBUG] Updated Data Proc Subcluster %q", updateReq.SubclusterId)
	return nil
}

func dataprocServiceNames() []string {
	names := make([]string, len(dataproc.HadoopConfig_Service_name)-1)
	for idx, serviceName := range dataproc.HadoopConfig_Service_name {
		if idx > 0 {
			names[idx-1] = serviceName
		}
	}
	return names
}
