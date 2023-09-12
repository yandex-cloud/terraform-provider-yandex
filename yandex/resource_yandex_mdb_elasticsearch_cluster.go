package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/elasticsearch/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBElasticsearchClusterCreateTimeout = 30 * time.Minute
	yandexMDBElasticsearchClusterDeleteTimeout = 15 * time.Minute
	yandexMDBElasticsearchClusterUpdateTimeout = 60 * time.Minute
)

func resourceYandexMDBElasticsearchCluster() *schema.Resource {
	return &schema.Resource{

		Create: resourceYandexMDBElasticsearchClusterCreate,
		Read:   resourceYandexMDBElasticsearchClusterRead,
		Update: resourceYandexMDBElasticsearchClusterUpdate,
		Delete: resourceYandexMDBElasticsearchClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBElasticsearchClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBElasticsearchClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBElasticsearchClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		CustomizeDiff: elasticsearchHostDiffCustomize,

		Schema: map[string]*schema.Schema{

			// Name of the Elasticsearch cluster. The name must be unique within the folder.
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			// Description of the Elasticsearch cluster.
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Custom labels for the Elasticsearch cluster as `key:value` pairs.
			"labels": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			// ID of the folder that the Elasticsearch cluster belongs to.
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true, // TODO impl move cluster
			},

			// Deployment environment of the Elasticsearch cluster.
			"environment": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// ID of the network that the cluster belongs to.
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service_account_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Configuration of the Elasticsearch cluster.
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"edition": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},

						"admin_password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},

						"data_node": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						},

						"master_node": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// suppress diff for not defined masters nodes
								h, _ := expandElasticsearchHosts(d.Get("host"))
								return !h.HasMasters()
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
													Required: true,
													ForceNew: true,
												},
											},
										},
									},
								},
							},
						}, // masternode

						"plugins": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Set:      schema.HashString,
							Optional: true,
						},
					},
				},
			},

			// User security groups
			"security_group_ids": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},

			// Aggregated cluster health.
			"health": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Current state of the cluster.
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Creation timestamp.
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Hosts of the Elasticsearch cluster
			"host": {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					// cause error attribute supports 1 item as a minimum, config has 0 declared
					return []interface{}{}, nil
				},
				Computed: true,
				Elem:     elasticsearchHostResource,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"maintenance_window": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"ANYTIME", "WEEKLY"}, false),
							Required:     true,
						},
						"day": {
							Type:         schema.TypeString,
							ValidateFunc: validateParsableValue(parseElasticsearchWeekDay),
							Optional:     true,
						},
						"hour": {
							Type:         schema.TypeInt,
							ValidateFunc: validation.IntBetween(1, 24),
							Optional:     true,
						},
					},
				},
			},
		},
	}
}

var elasticsearchHostResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		// Unique host name
		"name": {
			Type:     schema.TypeString,
			Required: true,
		},
		// Domain name
		"fqdn": {
			Type:     schema.TypeString,
			Computed: true,
		},
		// Availability zone
		"zone": {
			Type:     schema.TypeString,
			Required: true,
		},
		// Host Type
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validateParsableValue(parseElasticsearchHostType),
		},
		"assign_public_ip": {
			Type:     schema.TypeBool,
			Optional: true,
			Default:  false,
		},
		// Host subnet
		"subnet_id": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
	},
}

func resourceYandexMDBElasticsearchClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().ElasticSearch().Cluster().Get(ctx, &elasticsearch.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())

	d.Set("folder_id", cluster.GetFolderId())
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("network_id", cluster.GetNetworkId())

	d.Set("name", cluster.GetName())
	d.Set("description", cluster.GetDescription())
	d.Set("service_account_id", cluster.GetServiceAccountId())

	if err := d.Set("labels", cluster.GetLabels()); err != nil {
		return err
	}

	password := ""
	if v, ok := d.GetOk("config.0.admin_password"); ok {
		password = v.(string)
	}
	clusterConfig := flattenElasticsearchClusterConfig(cluster.Config, password)

	if err := d.Set("config", clusterConfig); err != nil {
		return err
	}

	if cluster.SecurityGroupIds == nil {
		cluster.SecurityGroupIds = []string{}
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	actualHosts, err := listElasticsearchHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	hosts, err := expandElasticsearchHosts(d.Get("host"))
	if err != nil {
		return err
	}

	result, err := flattenElasticsearchHosts(mapElasticsearchHostNames(actualHosts, hosts))
	if err != nil {
		return err
	}

	if err := d.Set("host", result); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.GetDeletionProtection())

	mw := flattenElasticsearchMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	return nil
}

func resourceYandexMDBElasticsearchClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateElasticsearchRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().ElasticSearch().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Elasticsearch Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get Elasticsearch Cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*elasticsearch.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get Elasticsearch Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create Elasticsearch Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Elasticsearch Cluster creation failed: %s", err)
	}

	return resourceYandexMDBElasticsearchClusterRead(d, meta)
}

func prepareCreateElasticsearchRequest(d *schema.ResourceData, meta *Config) (*elasticsearch.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Elasticsearch Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating Elasticsearch Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseElasticsearchEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating Elasticsearch Cluster: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	config := expandElasticsearchConfigSpec(d)

	hosts, err := expandElasticsearchHosts(d.Get("host"))
	if err != nil {
		return nil, fmt.Errorf("Error while expanding hosts on Elasticsearch Cluster create: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on Elasticsearch Cluster create: %s", err)
	}

	mntWindow, err := expandElasticsearchMaintenanceWindow(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding maintenance window on Elasticsearch Cluster create: %s", err)
	}

	req := &elasticsearch.CreateClusterRequest{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,

		FolderId:    folderID,
		Environment: env,
		NetworkId:   networkID,

		ConfigSpec: config,

		HostSpecs: convertElasticsearchHostsToSpecs(hosts),

		SecurityGroupIds:   securityGroupIds,
		ServiceAccountId:   d.Get("service_account_id").(string),
		DeletionProtection: d.Get("deletion_protection").(bool),
		MaintenanceWindow:  mntWindow,
	}

	return req, nil
}

func resourceYandexMDBElasticsearchClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Elasticsearch Cluster %q", d.Id())

	req := &elasticsearch.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().ElasticSearch().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Elasticsearch Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Elasticsearch Cluster %q", d.Id())

	return nil
}

func resourceYandexMDBElasticsearchClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if err := updateElasticsearchClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("host") {
		if err := updateElasticsearchClusterHosts(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)
	return resourceYandexMDBElasticsearchClusterRead(d, meta)
}

func updateElasticsearchClusterHosts(d *schema.ResourceData, meta interface{}) error {
	os, ns := d.GetChange("host")
	oldHosts, newHosts := os.(*schema.Set), ns.(*schema.Set)

	ofs := schema.NewSet(elasticsearchHostFQDNHash, oldHosts.List())
	nfs := schema.NewSet(elasticsearchHostFQDNHash, newHosts.List())

	toUpdate, err := expandElasticsearchHosts(ofs.Intersection(nfs))
	if err != nil {
		return err
	}

	var update = map[string]bool{}
	for i := range toUpdate {
		update[toUpdate[i].Fqdn] = true
	}
	// there are only name changes can be, just everything is done. (no host changes implemented)
	// toUpdate, err := expandElasticsearchHosts(updateHosts.List())

	toCreate, err := expandElasticsearchHosts(newHosts.Difference(oldHosts))
	if err != nil {
		return err
	}

	k := 0
	for i := range toCreate {
		if !update[toCreate[i].Fqdn] {
			toCreate[i], toCreate[k] = toCreate[k], toCreate[i]
			k++
		}
	}
	toCreate = toCreate[:k]

	log.Printf("[DEBUG] Create Hosts Elasticsearch Cluster %q: %d", d.Id(), len(toCreate))

	// api support only one by one
	for _, host := range toCreate {
		err := makeCreateElasticsearchHostRequest(d.Id(), host, d, meta)
		if err != nil {
			return err
		}
	}

	toDelete, err := expandElasticsearchHosts(oldHosts.Difference(newHosts))
	if err != nil {
		return err
	}

	k = 0
	for i := range toDelete {
		if !update[toDelete[i].Fqdn] {
			toDelete[i], toDelete[k] = toDelete[k], toDelete[i]
			k++
		}
	}
	toDelete = toDelete[:k]

	log.Printf("[DEBUG] Delete Hosts Elasticsearch Cluster %q: %d", d.Id(), len(toDelete))

	for _, host := range toDelete {
		err := makeDeleteElasticsearchHostRequest(d.Id(), host, d, meta)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateElasticsearchClusterParams(d *schema.ResourceData, meta interface{}) error {
	req := &elasticsearch.UpdateClusterRequest{
		ClusterId: d.Id(),
		UpdateMask: &field_mask.FieldMask{
			Paths: make([]string, 0, 16),
		},
	}
	changed := make([]string, 0, 16)

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")

		changed = append(changed, "description")
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")

		changed = append(changed, "name")
	}

	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")

		changed = append(changed, "labels")
	}

	if d.HasChange("security_group_ids") {
		securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

		req.SecurityGroupIds = securityGroupIds
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "security_group_ids")

		changed = append(changed, "security_group_ids")
	}

	if d.HasChange("service_account_id") {
		req.ServiceAccountId = d.Get("service_account_id").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "service_account_id")

		changed = append(changed, "service_account_id")
	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = d.Get("deletion_protection").(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "deletion_protection")

		changed = append(changed, "deletion_protection")
	}

	// TODO  folder

	if d.HasChange("config") {
		req.ConfigSpec = expandElasticsearchConfigSpecUpdate(d)

		fields := map[string]string{
			"config.0.data_node.0.resources.0.resource_preset_id":   "config_spec.elasticsearch_spec.data_node.resources.resource_preset_id",
			"config.0.data_node.0.resources.0.disk_size":            "config_spec.elasticsearch_spec.data_node.resources.disk_size",
			"config.0.master_node.0.resources.0.resource_preset_id": "config_spec.elasticsearch_spec.master_node.resources.resource_preset_id",
			"config.0.master_node.0.resources.0.disk_size":          "config_spec.elasticsearch_spec.master_node.resources.disk_size",
			"config.0.plugins":        "config_spec.elasticsearch_spec.plugins",
			"config.0.admin_password": "config_spec.admin_password",
			"config.0.edition":        "config_spec.edition",
			"config.0.version":        "config_spec.version",
		}

		for key, path := range fields {
			if d.HasChange(key) {
				req.UpdateMask.Paths = append(req.UpdateMask.Paths, path)
				changed = append(changed, key)
			}
		}

	}

	if d.HasChange("maintenance_window") {
		mw, err := expandElasticsearchMaintenanceWindow(d)
		if err != nil {
			return err
		}

		req.MaintenanceWindow = mw
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "maintenance_window")

		changed = append(changed, "maintenance_window")
	}

	if len(changed) == 0 {
		return nil // nothing to update
	}

	err := makeElasticsearchClusterUpdateRequest(req, d, meta)
	if err != nil {
		return err
	}

	return nil
}

func makeElasticsearchClusterUpdateRequest(req *elasticsearch.UpdateClusterRequest, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().ElasticSearch().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update Elasticsearch Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating Elasticsearch Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func makeCreateElasticsearchHostRequest(clusterID string, host *ElasticsearchHost, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().ElasticSearch().Cluster().AddHosts(ctx, &elasticsearch.AddClusterHostsRequest{
		ClusterId: clusterID,
		HostSpecs: convertElasticsearchHostsToSpecs([]*ElasticsearchHost{host}),
	}))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Elasticsearch Host %q for Cluster %q: %s", host.Name, clusterID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error creating Elasticsearch Host %q for Cluster %q: %s", host.Name, clusterID, err)
	}
	return nil
}

func makeDeleteElasticsearchHostRequest(clusterID string, host *ElasticsearchHost, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().ElasticSearch().Cluster().DeleteHosts(ctx, &elasticsearch.DeleteClusterHostsRequest{
		ClusterId: clusterID,
		HostNames: []string{host.Fqdn},
	}))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create Elasticsearch Host %q for Cluster %q: %s", host.Fqdn, clusterID, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error creating Elasticsearch Host %q for Cluster %q: %s", host.Fqdn, clusterID, err)
	}
	return nil
}

func listElasticsearchHosts(ctx context.Context, config *Config, clusterID string) ([]*elasticsearch.Host, error) {
	hosts := []*elasticsearch.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().ElasticSearch().Cluster().ListHosts(ctx, &elasticsearch.ListClusterHostsRequest{
			ClusterId: clusterID,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of hosts for '%s': %s", clusterID, err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" || resp.NextPageToken == "0" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}
