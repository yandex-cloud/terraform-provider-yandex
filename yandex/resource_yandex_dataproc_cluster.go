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
	"github.com/yandex-cloud/terraform-provider-yandex/common"
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
		Description: "Manages a Yandex Data Processing cluster. For more information, see [the official documentation](https://yandex.cloud/docs/data-proc/).",

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
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},

			"service_account_id": {
				Type:        schema.TypeString,
				Description: "Service account to be used by the Yandex Data Processing agent to access resources of Yandex Cloud. Selected service account should have `mdb.dataproc.agent` role on the folder where the Yandex Data Processing cluster will be located.",
				Required:    true,
			},

			"cluster_config": {
				Type:        schema.TypeList,
				Description: "Configuration and resources for hosts that should be created with the cluster.",
				MaxItems:    1,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subcluster_spec": {
							Type:        schema.TypeList,
							Description: "Configuration of the Yandex Data Processing subcluster.",
							MinItems:    1,
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Description: "Name of the Yandex Data Processing subcluster.",
										Required:    true,
									},

									"role": {
										Type:        schema.TypeString,
										Description: "Role of the subcluster in the Yandex Data Processing cluster.",
										Required:    true,
										ValidateFunc: validation.StringInSlice(
											[]string{"MASTERNODE", "DATANODE", "COMPUTENODE"}, false),
									},

									"resources": {
										Type:        schema.TypeList,
										Description: "Resources allocated to each host of the Yandex Data Processing subcluster.",
										MaxItems:    1,
										Required:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"resource_preset_id": {
													Type:        schema.TypeString,
													Description: "The ID of the preset for computational resources available to a host. All available presets are listed in the [documentation](https://yandex.cloud/docs/data-proc/concepts/instance-types).",
													Required:    true,
												},
												"disk_size": {
													Type:        schema.TypeInt,
													Description: "Volume of the storage available to a host, in gigabytes.",
													Required:    true,
												},
												"disk_type_id": {
													Type:        schema.TypeString,
													Description: "Type of the storage of a host. One of `network-hdd` (default) or `network-ssd`.",
													Optional:    true,
													ForceNew:    true,
													Default:     "network-hdd",
												},
											},
										},
									},

									"autoscaling_config": {
										Type:        schema.TypeList,
										Description: "Autoscaling configuration for compute subclusters.",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"max_hosts_count": {
													Type:        schema.TypeInt,
													Description: "Maximum number of nodes in autoscaling subclusters.",
													Required:    true,
												},
												"preemptible": {
													Type:        schema.TypeBool,
													Description: "Use preemptible compute instances. Preemptible instances are stopped at least once every 24 hours, and can be stopped at any time if their resources are needed by Compute. For more information, see [Preemptible Virtual Machines](https://yandex.cloud/docs/compute/concepts/preemptible-vm).",
													Optional:    true,
													Default:     false,
												},
												"measurement_duration": {
													Type:         schema.TypeString,
													Description:  "Time in seconds allotted for averaging metrics.",
													Optional:     true,
													Computed:     true,
													ValidateFunc: ConvertableToInt(),
												},
												"warmup_duration": {
													Type:         schema.TypeString,
													Description:  "The warmup time of the instance in seconds. During this time, traffic is sent to the instance, but instance metrics are not collected.",
													Optional:     true,
													Computed:     true,
													ValidateFunc: ConvertableToInt(),
												},
												"stabilization_duration": {
													Type:         schema.TypeString,
													Description:  "Minimum amount of time in seconds allotted for monitoring before Instance Groups can reduce the number of instances in the group. During this time, the group size doesn't decrease, even if the new metric values indicate that it should.",
													Optional:     true,
													Computed:     true,
													ValidateFunc: ConvertableToInt(),
												},
												"cpu_utilization_target": {
													Type:         schema.TypeString,
													Description:  "Defines an autoscaling rule based on the average CPU utilization of the instance group. If not set default autoscaling metric will be used.",
													Optional:     true,
													Computed:     true,
													ValidateFunc: ConvertableToInt(),
												},
												"decommission_timeout": {
													Type:         schema.TypeString,
													Description:  "Timeout to gracefully decommission nodes during downscaling. In seconds.",
													Optional:     true,
													Computed:     true,
													ValidateFunc: ConvertableToInt(),
												},
											},
										},
									},

									"subnet_id": {
										Type:        schema.TypeString,
										Description: "The ID of the subnet, to which hosts of the subcluster belong. Subnets of all the subclusters must belong to the same VPC network.",
										Required:    true,
									},

									"hosts_count": {
										Type:         schema.TypeInt,
										Description:  "Number of hosts within Yandex Data Processing subcluster.",
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},

									"id": {
										Type:        schema.TypeString,
										Description: "ID of the subcluster.",
										Computed:    true,
									},

									"assign_public_ip": {
										Type:        schema.TypeBool,
										Description: "If `true` then assign public IP addresses to the hosts of the subclusters.",
										Optional:    true,
										ForceNew:    true,
										Default:     false,
									},
								},
							},
						},
						"hadoop": {
							Type:        schema.TypeList,
							Description: "Yandex Data Processing specific options.",
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"services": {
										Type:        schema.TypeSet,
										Description: "List of services to run on Yandex Data Processing cluster.",
										Optional:    true,
										ForceNew:    true,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringInSlice(dataprocServiceNames(), false),
										},
									},

									"properties": {
										Type:        schema.TypeMap,
										Description: "A set of key/value pairs that are used to configure cluster services.",
										Optional:    true,
										Elem:        &schema.Schema{Type: schema.TypeString},
									},

									"ssh_public_keys": {
										Type:        schema.TypeSet,
										Description: "List of SSH public keys to put to the hosts of the cluster. For information on how to connect to the cluster, see [the official documentation](https://yandex.cloud/docs/data-proc/operations/connect).",
										Optional:    true,
										ForceNew:    true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},

									"initialization_action": {
										Type:        schema.TypeList,
										Description: "List of initialization scripts.",
										Optional:    true,
										ForceNew:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"uri": {
													Type:        schema.TypeString,
													Description: "Script URI.",
													Required:    true,
													ForceNew:    true,
												},
												"args": {
													Type:        schema.TypeList,
													Description: "List of arguments of the initialization script.",
													Optional:    true,
													ForceNew:    true,
													Computed:    true,
													Elem: &schema.Schema{
														Type: schema.TypeString,
													},
												},
												"timeout": {
													Type:         schema.TypeString,
													Description:  "Script execution timeout, in seconds.",
													Optional:     true,
													ForceNew:     true,
													Computed:     true,
													ValidateFunc: ConvertableToInt(),
												},
											},
										},
									},
									"oslogin": {
										Type:        schema.TypeBool,
										Description: "Whether to enable authorization via OS Login.",
										Optional:    true,
									},
								},
							},
						},
						"version_id": {
							Type:        schema.TypeString,
							Description: "Version of Yandex Data Processing image.",
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return isVersionPrefix(new, old)
							},
						},
					},
				},
			},

			"bucket": {
				Type:        schema.TypeString,
				Description: "Name of the Object Storage bucket to use for Yandex Data Processing jobs. Yandex Data Processing Agent saves output of job driver's process to specified bucket. In order for this to work service account (specified by the `service_account_id` argument) should be given permission to create objects within this bucket.",
				Optional:    true,
			},

			"description": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["description"],
				Optional:    true,
			},

			"labels": {
				Type:        schema.TypeMap,
				Description: common.ResourceDescriptions["labels"],
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},

			"ui_proxy": {
				Type:        schema.TypeBool,
				Description: "Whether to enable UI Proxy feature.",
				Optional:    true,
			},

			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
			},

			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: "A list of host group IDs to place VMs of the cluster on.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				ForceNew:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"zone_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["zone"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},

			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},

			"log_group_id": {
				Type:        schema.TypeString,
				Description: "ID of the cloud logging group for cluster logs.",
				Optional:    true,
			},

			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the cluster. Can be either `PRESTABLE` or `PRODUCTION`. The default is `PRESTABLE`.",
				Optional:     true,
				ForceNew:     true,
				Default:      "PRESTABLE",
				ValidateFunc: validateParsableValue(parseDataprocEnv),
			},
			"autoscaling_service_account_id": {
				Type:        schema.TypeString,
				Description: "Service account to be used for managing hosts in an autoscaled subcluster.",
				Optional:    true,
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
		return fmt.Errorf("error while requesting API to create Yandex Data Processing Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while getting Yandex Data Processing Cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*dataproc.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Yandex Data Processing Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create Yandex Data Processing Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("failed to create Yandex Data Processing Cluster: %s", err)
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

	if err := d.Set("log_group_id", cluster.LogGroupId); err != nil {
		return err
	}

	if err := d.Set("environment", cluster.Environment.String()); err != nil {
		return err
	}

	if err := d.Set("cluster_config", flattenDataprocClusterConfig(cluster, subclusters)); err != nil {
		return err
	}

	if err := d.Set("created_at", getTimestamp(cluster.CreatedAt)); err != nil {
		return err
	}

	if err := d.Set("deletion_protection", cluster.DeletionProtection); err != nil {
		return err
	}

	if err := d.Set("autoscaling_service_account_id", cluster.AutoscalingServiceAccountId); err != nil {
		return err
	}

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
		return nil, fmt.Errorf("error getting folder ID while creating Yandex Data Processing Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseDataprocEnv(e)
	if err != nil {
		return nil, fmt.Errorf("error resolving environment while creating Yandex Data Processing Cluster: %s", err)
	}

	zoneID, err := getDataprocZoneID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting zone while creating Yandex Data Processing Cluster: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Yandex Data Processing Cluster create: %s", err)
	}

	configSpec, err := expandDataprocCreateClusterConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding config on Yandex Data Processing Cluster create: %s", err)
	}

	req := dataproc.CreateClusterRequest{
		FolderId:                    folderID,
		Name:                        d.Get("name").(string),
		Description:                 d.Get("description").(string),
		Labels:                      labels,
		ConfigSpec:                  configSpec,
		ZoneId:                      zoneID,
		ServiceAccountId:            d.Get("service_account_id").(string),
		Bucket:                      d.Get("bucket").(string),
		UiProxy:                     d.Get("ui_proxy").(bool),
		SecurityGroupIds:            expandSecurityGroupIds(d.Get("security_group_ids")),
		HostGroupIds:                expandHostGroupIds(d.Get("host_group_ids")),
		DeletionProtection:          d.Get("deletion_protection").(bool),
		LogGroupId:                  d.Get("log_group_id").(string),
		Environment:                 env,
		AutoscalingServiceAccountId: d.Get("autoscaling_service_account_id").(string),
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
			return nil, fmt.Errorf("error while getting list of subclusters for Yandex Data Processing Cluster %q: %s", id, err)
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

	log.Printf("[DEBUG] Deleting Yandex Data Processing Cluster %q", d.Id())

	req := &dataproc.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Yandex Data Processing Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Yandex Data Processing Cluster %q", d.Id())
	return nil
}

func resourceYandexDataprocClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating Yandex Data Processing Cluster %q", d.Id())

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

	log.Printf("[DEBUG] Finished updating Yandex Data Processing Cluster %q", d.Id())
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
		return fmt.Errorf("error while requesting API to update Yandex Data Processing Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating Yandex Data Processing Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func getDataprocClusterUpdateRequest(d *schema.ResourceData) (*dataproc.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Yandex Data Processing Cluster update: %s", err)
	}

	req := &dataproc.UpdateClusterRequest{
		ClusterId:                   d.Id(),
		Description:                 d.Get("description").(string),
		Labels:                      labels,
		Name:                        d.Get("name").(string),
		ServiceAccountId:            d.Get("service_account_id").(string),
		Bucket:                      d.Get("bucket").(string),
		UiProxy:                     d.Get("ui_proxy").(bool),
		SecurityGroupIds:            expandSecurityGroupIds(d.Get("security_group_ids")),
		LogGroupId:                  d.Get("log_group_id").(string),
		DeletionProtection:          d.Get("deletion_protection").(bool),
		AutoscalingServiceAccountId: d.Get("autoscaling_service_account_id").(string),
	}

	var updatePaths []string
	fieldNames := []string{"description", "labels", "name", "service_account_id", "bucket", "ui_proxy", "security_group_ids", "deletion_protection", "log_group_id", "autoscaling_service_account_id"}
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
			req, err := getDataprocSubclusterCreateRequest(clusterID, subclusterSpec)
			if err != nil {
				return nil, nil, nil, err
			}
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

func getDataprocSubclusterCreateRequest(clusterID string, element interface{}) (*dataproc.CreateSubclusterRequest, error) {
	createSpec, err := expandDataprocSubclusterSpec(element)
	if err != nil {
		return nil, err
	}

	return &dataproc.CreateSubclusterRequest{
		ClusterId:         clusterID,
		Name:              createSpec.Name,
		Role:              createSpec.Role,
		Resources:         createSpec.Resources,
		SubnetId:          createSpec.SubnetId,
		HostsCount:        createSpec.HostsCount,
		AutoscalingConfig: createSpec.AutoscalingConfig,
	}, nil
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
		autoscalingConfig, err := expandDataprocAutoscalingConfig(autoscalingConfigs[0])
		if err != nil {
			return nil, err
		}
		req.AutoscalingConfig = autoscalingConfig
	}

	constFields := []string{"role", "subnet_id"}
	for _, fieldName := range constFields {
		field := path + "." + fieldName
		if d.HasChange(field) {
			return nil, fmt.Errorf("error while trying to update Yandex Data Processing Subcluster %q:"+
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
	log.Printf("[DEBUG] Deleting Yandex Data Processing Subcluster %q", deleteReq.SubclusterId)

	ctx, cancel := config.ContextWithTimeout(timeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Subcluster().Delete(ctx, deleteReq))
	if err != nil {
		return fmt.Errorf("error while requesting API to delete Yandex Data Processing Subcluster %q: %s", deleteReq.SubclusterId, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting Yandex Data Processing Subcluster %q: %s", deleteReq.SubclusterId, err)
	}

	log.Printf("[DEBUG] Deleted Yandex Data Processing Subcluster %q", deleteReq.SubclusterId)
	return nil
}

func createDataprocSubcluster(createReq *dataproc.CreateSubclusterRequest, config *Config, timeout time.Duration) error {
	log.Printf("[DEBUG] Creating Yandex Data Processing Subcluster %q", createReq.Name)

	ctx, cancel := config.ContextWithTimeout(timeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Subcluster().Create(ctx, createReq))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Yandex Data Processing Subcluster %q: %s", createReq.Name, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating Yandex Data Processing Subcluster %q: %s", createReq.Name, err)
	}

	log.Printf("[DEBUG] Created Yandex Data Processing Subcluster %q", createReq.Name)
	return nil
}

func updateDataprocSubcluster(updateReq *dataproc.UpdateSubclusterRequest, config *Config, timeout time.Duration) error {
	log.Printf("[DEBUG] Updating Yandex Data Processing Subcluster %q", updateReq.SubclusterId)

	ctx, cancel := config.ContextWithTimeout(timeout)
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.Dataproc().Subcluster().Update(ctx, updateReq))
	if err != nil {
		return fmt.Errorf("error while requesting API to update Yandex Data Processing Subcluster %q: %s", updateReq.SubclusterId, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating Yandex Data Processing Subcluster %q: %s", updateReq.SubclusterId, err)
	}

	log.Printf("[DEBUG] Updated Yandex Data Processing Subcluster %q", updateReq.SubclusterId)
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
