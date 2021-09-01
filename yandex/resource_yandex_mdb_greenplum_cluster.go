package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
)

const (
	yandexMDBGreenplumClusterDefaultTimeout = 120 * time.Minute
	yandexMDBGreenplumClusterUpdateTimeout  = 120 * time.Minute
)

func resourceYandexMDBGreenplumCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBGreenplumClusterCreate,
		Read:   resourceYandexMDBGreenplumClusterRead,
		Update: resourceYandexMDBGreenplumClusterUpdate,
		Delete: resourceYandexMDBGreenplumClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBGreenplumClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBGreenplumClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBGreenplumClusterDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseGreenplumEnv),
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"subnet_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"assign_public_ip": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
			},
			"master_host_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"segment_host_count": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"segment_in_host": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"master_subcluster": {
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
									"resource_preset_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"segment_subcluster": {
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
									"resource_preset_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"master_hosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"segment_hosts": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"user_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceYandexMDBGreenplumClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := prepareCreateGreenplumRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.MDB().Greenplum().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Greenplum Cluster: %s", err)
	}
	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Greenplum create operation metadata: %s", err)
	}
	md, ok := protoMetadata.(*greenplum.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get Greenplum Cluster ID from create operation metadata")
	}
	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create Greenplum Cluster: %s", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Greenplum Cluster creation failed: %s", err)
	}
	return resourceYandexMDBGreenplumClusterRead(d, meta)
}

func prepareCreateGreenplumRequest(d *schema.ResourceData, meta *Config) (*greenplum.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, fmt.Errorf("Error while expanding labels on Greenplum Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating Greenplum Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseGreenplumEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating Greenplum Cluster: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	req := greenplum.CreateClusterRequest{
		FolderId:         folderID,
		Name:             d.Get("name").(string),
		Description:      d.Get("description").(string),
		NetworkId:        d.Get("network_id").(string),
		Environment:      env,
		Labels:           labels,
		SecurityGroupIds: securityGroupIds,

		MasterHostCount:  int64(d.Get("master_host_count").(int)),
		SegmentInHost:    int64(d.Get("segment_in_host").(int)),
		SegmentHostCount: int64(d.Get("segment_host_count").(int)),

		Config: &greenplum.GreenplumConfig{
			Version:        d.Get("version").(string),
			ZoneId:         d.Get("zone").(string),
			SubnetId:       d.Get("subnet_id").(string),
			AssignPublicIp: d.Get("assign_public_ip").(bool),
		},
		MasterConfig: &greenplum.MasterSubclusterConfigSpec{
			Resources: &greenplum.Resources{
				ResourcePresetId: d.Get("master_subcluster.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("master_subcluster.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("master_subcluster.0.resources.0.disk_size").(int)),
			},
		},
		SegmentConfig: &greenplum.SegmentSubclusterConfigSpec{
			Resources: &greenplum.Resources{
				ResourcePresetId: d.Get("segment_subcluster.0.resources.0.resource_preset_id").(string),
				DiskTypeId:       d.Get("segment_subcluster.0.resources.0.disk_type_id").(string),
				DiskSize:         toBytes(d.Get("segment_subcluster.0.resources.0.disk_size").(int)),
			},
		},

		UserName:     d.Get("user_name").(string),
		UserPassword: d.Get("user_password").(string),
	}
	return &req, nil
}

func resourceYandexMDBGreenplumClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Greenplum().Cluster().Get(ctx, &greenplum.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	d.Set("folder_id", cluster.GetFolderId())
	d.Set("name", cluster.GetName())
	d.Set("description", cluster.GetDescription())
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("network_id", cluster.GetNetworkId())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("version", cluster.GetConfig().GetVersion())

	d.Set("zone", cluster.GetConfig().ZoneId)
	d.Set("subnet_id", cluster.GetConfig().SubnetId)
	d.Set("assign_public_ip", cluster.GetConfig().AssignPublicIp)
	d.Set("version", cluster.GetConfig().Version)

	d.Set("master_host_count", cluster.GetMasterHostCount())
	d.Set("segment_host_count", cluster.GetSegmentHostCount())
	d.Set("segment_in_host", cluster.GetSegmentInHost())

	d.Set("user_name", cluster.GetUserName())

	masterSubcluster := map[string]interface{}{}
	masterResources := map[string]interface{}{}
	masterResources["resource_preset_id"] = cluster.GetMasterConfig().Resources.ResourcePresetId
	masterResources["disk_type_id"] = cluster.GetMasterConfig().Resources.DiskTypeId
	masterResources["disk_size"] = toGigabytes(cluster.GetMasterConfig().Resources.DiskSize)
	masterSubcluster["resources"] = []map[string]interface{}{masterResources}
	d.Set("master_subcluster", []map[string]interface{}{masterSubcluster})

	segmentSubcluster := map[string]interface{}{}
	segmentResources := map[string]interface{}{}
	segmentResources["resource_preset_id"] = cluster.GetMasterConfig().Resources.ResourcePresetId
	segmentResources["disk_type_id"] = cluster.GetMasterConfig().Resources.DiskTypeId
	segmentResources["disk_size"] = toGigabytes(cluster.GetMasterConfig().Resources.DiskSize)
	segmentSubcluster["resources"] = []map[string]interface{}{segmentResources}
	d.Set("segment_subcluster", []map[string]interface{}{segmentSubcluster})

	if cluster.Labels == nil {
		if err = d.Set("labels", make(map[string]string)); err != nil {
			return err
		}
	} else if err = d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if cluster.SecurityGroupIds == nil {
		if err = d.Set("security_group_ids", make([]string, 0)); err != nil {
			return err
		}
	} else if err = d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	masterHosts, err := listGreenplumMasterHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	mHost := make([]map[string]interface{}, 0, len(masterHosts))
	for _, h := range masterHosts {
		mHost = append(mHost, map[string]interface{}{"fqdn": h.Name, "assign_public_ip": h.AssignPublicIp})
	}
	if err = d.Set("master_hosts", mHost); err != nil {
		return err
	}

	segmentHosts, err := listGreenplumSegmentHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	sHost := make([]map[string]interface{}, 0, len(segmentHosts))
	for _, h := range segmentHosts {
		sHost = append(sHost, map[string]interface{}{"fqdn": h.Name})
	}
	if err = d.Set("segment_hosts", sHost); err != nil {
		return err
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	return d.Set("created_at", createdAt)
}

func listGreenplumMasterHosts(ctx context.Context, config *Config, id string) ([]*greenplum.Host, error) {
	hosts := []*greenplum.Host{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().Greenplum().Cluster().ListMasterHosts(ctx, &greenplum.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", id, err)
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts, nil
}
func listGreenplumSegmentHosts(ctx context.Context, config *Config, id string) ([]*greenplum.Host, error) {
	hosts := []*greenplum.Host{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().Greenplum().Cluster().ListSegmentHosts(ctx, &greenplum.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", id, err)
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts, nil
}

func resourceYandexMDBGreenplumClusterUpdate(d *schema.ResourceData, meta interface{}) error {

	return fmt.Errorf("Changing is not supported for Greenplum Cluster. Id: %v", d.Id())
}

func resourceYandexMDBGreenplumClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Greenplum Cluster %q", d.Id())

	req := &greenplum.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Greenplum().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Greenplum Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Greenplum Cluster %q", d.Id())
	return nil
}
