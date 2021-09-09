package yandex

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

func dataSourceYandexMDBGreenplumCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexMDBGreenplumClusterRead,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"assign_public_ip": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_host_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"segment_host_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"segment_in_host": {
				Type:     schema.TypeInt,
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
				Set:      schema.HashString,
			},

			"master_subcluster": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"segment_subcluster": {
				Type:     schema.TypeList,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Computed: true,
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
				Computed: true,
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
				Computed: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexMDBGreenplumClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx := context.Background()

	err := checkOneOf(d, "cluster_id", "name")
	if err != nil {
		return err
	}

	clusterID := d.Get("cluster_id").(string)
	_, clusterNameOk := d.GetOk("name")

	if clusterNameOk {
		clusterID, err = resolveObjectID(ctx, config, d, sdkresolvers.GreenplumClusterResolver)
		if err != nil {
			return fmt.Errorf("failed to resolve data source Greenplum Cluster by name: %v", err)
		}
	}
	cluster, err := config.sdk.MDB().Greenplum().Cluster().Get(ctx, &greenplum.GetClusterRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	d.Set("folder_id", cluster.GetFolderId())
	d.Set("cluster_id", cluster.Id)
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

	masterHosts, err := listGreenplumMasterHosts(ctx, config, cluster.GetId())
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

	segmentHosts, err := listGreenplumSegmentHosts(ctx, config, cluster.GetId())
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

	d.Set("deletion_protection", cluster.DeletionProtection)

	d.Set("created_at", getTimestamp(cluster.CreatedAt))

	d.SetId(cluster.Id)
	return nil
}
