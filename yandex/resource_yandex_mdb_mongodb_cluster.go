package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
)

func resourceYandexMDBMongodbCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceYandexMDBMongodbClusterCreate,
		ReadContext:   resourceYandexMDBMongodbClusterRead,
		UpdateContext: resourceYandexMDBMongodbClusterUpdate,
		DeleteContext: resourceYandexMDBMongodbClusterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update:  schema.DefaultTimeout(60 * time.Minute),
			Default: schema.DefaultTimeout(30 * time.Minute),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseMongoDBEnv),
			},
			"user": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      mongodbUserHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"password": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
						"permission": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Set:      mongodbUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"roles": {
										Type:     schema.TypeList,
										Optional: true,
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
			"database": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      mongodbDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"host": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"zone_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"health": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"shard_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
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
						"disk_size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"disk_type_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"cluster_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"version": {
							Type:     schema.TypeString,
							Required: true,
						},
						"feature_compatibility_version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"backup_window_start": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"hours": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 23),
									},
									"minutes": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      0,
										ValidateFunc: validation.IntBetween(0, 59),
									},
								},
							},
						},
						"access": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_lens": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						"mongod": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audit_log": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"filter": {
													Type:     schema.TypeString,
													Optional: true,
												},
												"runtime_configuration": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"set_parameter": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"audit_authorization_success": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
									"security": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enable_encryption": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"kmip": {
													Type:     schema.TypeList,
													MaxItems: 1,
													Optional: true,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"server_name": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"port": {
																Type:     schema.TypeInt,
																Optional: true,
															},
															"server_ca": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"client_certificate": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"key_identifier": {
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
					},
				},
			},
			"cluster_id": {
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
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"sharded": {
				Type:     schema.TypeBool,
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
				Optional: true,
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
							ValidateFunc: validateParsableValue(parseMongoDBWeekDay),
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
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func prepareCreateMongodbRequest(d *schema.ResourceData, meta *Config) (*mongodb.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Mongodb Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("error getting folder ID while creating Mongodb Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseMongoDBEnv(e)
	if err != nil {
		return nil, fmt.Errorf("error resolving environment while creating Mongodb Cluster: %s", err)
	}

	version, err := extractVersion(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on Mongodb Cluster create: %s", err)
	}
	configSpec := &mongodb.ConfigSpec{Version: version, FeatureCompatibilityVersion: version}
	if cfgCompVer := d.Get("cluster_config.0.feature_compatibility_version"); cfgCompVer != nil {
		configSpec.FeatureCompatibilityVersion = cfgCompVer.(string)
	}

	if backupStart := d.Get("cluster_config.0.backup_window_start"); backupStart != nil {
		configSpec.BackupWindowStart = expandMongoDBBackupWindowStart(d)
	}

	if access := d.Get("cluster_config.0.access"); access != nil {
		configSpec.Access = &mongodb.Access{
			DataLens: d.Get("cluster_config.0.access.0.data_lens").(bool),
		}
	}

	mongodbSpecHelper := GetMongodbSpecHelper(version)
	configSpec.MongodbSpec = mongodbSpecHelper.Expand(d)

	hosts, err := expandMongoDBHosts(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding hosts on MongoDB Cluster create: %s", err)
	}

	dbSpecs, err := expandMongoDBDatabases(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding databases on MongoDB Cluster create: %s", err)
	}

	users, err := expandMongoDBUserSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding user specs on MongoDB Cluster create: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on MongoDB Cluster create: %s", err)
	}

	req := mongodb.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Environment:        env,
		ConfigSpec:         configSpec,
		HostSpecs:          hosts,
		UserSpecs:          users,
		DatabaseSpecs:      dbSpecs,
		Labels:             labels,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}
	return &req, nil
}

func resourceYandexMDBMongodbClusterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	req, err := prepareCreateMongodbRequest(d, config)

	if err != nil {
		return diag.FromErr(err)
	}

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Create(ctx, req))
	if err != nil {
		return diag.Errorf("error while requesting API to create Mongodb Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return diag.Errorf("error while get Mongodb create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*mongodb.CreateClusterMetadata)
	if !ok {
		return diag.Errorf("could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return diag.Errorf("error while waiting for operation to create Mongodb Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return diag.Errorf("Mongodb Cluster creation failed: %s", err)
	}

	mw, err := expandMongoDBMaintenanceWindow(d)
	if err != nil {
		return diag.FromErr(err)
	}
	if mw != nil {
		err = updateMongoDBMaintenanceWindow(ctx, config, d, mw)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceYandexMDBMongodbClusterRead(ctx, d, meta)
}

func updateMongoDBMaintenanceWindow(ctx context.Context, config *Config, d *schema.ResourceData, mw *mongodb.MaintenanceWindow) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().Update(ctx, &mongodb.UpdateClusterRequest{
			ClusterId:         d.Id(),
			MaintenanceWindow: mw,
			UpdateMask:        &field_mask.FieldMask{Paths: []string{"maintenance_window"}},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update maintenance window in MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating maintenance window in MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func listMongodbHosts(ctx context.Context, config *Config, d *schema.ResourceData) ([]*mongodb.Host, error) {
	var hosts []*mongodb.Host
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MongoDB().Cluster().ListHosts(ctx, &mongodb.ListClusterHostsRequest{
			ClusterId: d.Id(),
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for '%s': %s", d.Id(), err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}

func resourceYandexMDBMongodbClusterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	cluster, err := config.sdk.MDB().MongoDB().Cluster().Get(ctx, &mongodb.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string))))
	}

	if err := d.Set("created_at", getTimestamp(cluster.CreatedAt)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", cluster.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("folder_id", cluster.FolderId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("network_id", cluster.NetworkId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("environment", cluster.GetEnvironment().String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("health", cluster.GetHealth().String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("status", cluster.GetStatus().String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", cluster.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("sharded", cluster.Sharded); err != nil {
		return diag.FromErr(err)
	}

	mongodbSpecHelper := GetMongodbSpecHelper(cluster.Config.Version)
	flattenResources, err := mongodbSpecHelper.FlattenResources(cluster.Config)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources", flattenResources); err != nil {
		return diag.FromErr(err)
	}

	expandUsers, err := expandMongoDBUserSpecs(d)
	if err != nil {
		return diag.FromErr(err)
	}
	passwords := mongodbUsersPasswords(expandUsers)

	clusterUsers, err := listMongodbUsers(ctx, config, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	flattenUsers := flattenMongoDBUsers(clusterUsers, passwords)

	if err := d.Set("user", flattenUsers); err != nil {
		return diag.FromErr(err)
	}

	clusterDatabases, err := listMongodbDatabases(ctx, config, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	flattenDatabases := flattenMongoDBDatabases(clusterDatabases)

	if err := d.Set("database", flattenDatabases); err != nil {
		return diag.FromErr(err)
	}

	flattenClusterConfig, err := flattenMongoDBClusterConfig(cluster.Config, d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_config", flattenClusterConfig); err != nil {
		return diag.FromErr(err)
	}

	clusterHosts, err := listMongodbHosts(ctx, config, d)
	if err != nil {
		return diag.FromErr(err)
	}

	expandHosts, err := expandMongoDBHosts(d)
	if err != nil {
		return diag.FromErr(err)
	}

	hosts := sortMongoDBHosts(clusterHosts, expandHosts)

	flattenHosts, err := flattenMongoDBHosts(hosts)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("host", flattenHosts); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return diag.FromErr(err)
	}

	mw := flattenMongoDBMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("deletion_protection", cluster.DeletionProtection); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("labels", cluster.Labels); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sortMongoDBHosts(hosts []*mongodb.Host, specs []*mongodb.HostSpec) []*mongodb.Host {
	for i, h := range specs {
		for j := i + 1; j < len(hosts); j++ {
			if h.ZoneId == hosts[j].ZoneId && (h.ShardName == "" || h.ShardName == hosts[j].ShardName) && h.Type == hosts[j].Type {
				hosts[i], hosts[j] = hosts[j], hosts[i]
				break
			}
		}
	}

	return hosts
}

func resourceYandexMDBMongodbClusterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.Partial(true)

	if err := updateMongodbClusterParams(ctx, d, meta); err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("database") {
		if err := updateMongodbClusterDatabases(ctx, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("user") {
		if err := updateMongodbClusterUsers(ctx, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("host") {
		if err := updateMongoDBClusterHosts(ctx, d, meta); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(false)
	return resourceYandexMDBMongodbClusterRead(ctx, d, meta)
}

func resourceYandexMDBMongodbClusterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Mongodb Cluster %q", d.Id())

	req := &mongodb.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Delete(ctx, req))
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("Mongodb Cluster %q", d.Get("name").(string))))
	}

	err = op.Wait(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = op.Response()
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Finished deleting Mongodb Cluster %q", d.Id())
	return nil
}

func listMongodbUsers(ctx context.Context, config *Config, id string) ([]*mongodb.User, error) {
	var users []*mongodb.User
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MongoDB().User().List(ctx, &mongodb.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of users for '%s': %s", id, err)
		}
		users = append(users, resp.Users...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return users, nil
}

func listMongodbDatabases(ctx context.Context, config *Config, id string) ([]*mongodb.Database, error) {
	var dbs []*mongodb.Database
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().MongoDB().Database().List(ctx, &mongodb.ListDatabasesRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of databases for '%s': %s", id, err)
		}
		dbs = append(dbs, resp.Databases...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return dbs, nil
}

func getMongoDBClusterUpdateRequest(d *schema.ResourceData) (*mongodb.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating MongoDB cluster: %s", err)
	}

	version, err := extractVersion(d)
	if err != nil {
		return nil, err
	}
	mongodbSpecHelper := GetMongodbSpecHelper(version)
	req := &mongodb.UpdateClusterRequest{
		ClusterId:   d.Id(),
		Description: d.Get("description").(string),
		Labels:      labels,
		Name:        d.Get("name").(string),
		ConfigSpec: &mongodb.ConfigSpec{
			Version:           version,
			MongodbSpec:       mongodbSpecHelper.Expand(d),
			BackupWindowStart: expandMongoDBBackupWindowStart(d),
			Access:            &mongodb.Access{DataLens: d.Get("cluster_config.0.access.0.data_lens").(bool)},
		},
		SecurityGroupIds: expandSecurityGroupIds(d.Get("security_group_ids")),
	}
	return req, nil
}

var mdbMongodbUpdateFieldsMap = map[string]string{
	"description":                          "description",
	"labels":                               "labels",
	"cluster_config.0.version":             "config_spec.version",
	"cluster_config.0.access":              "config_spec.access",
	"cluster_config.0.backup_window_start": "config_spec.backup_window_start",
	"security_group_ids":                   "security_group_ids",
	"deletion_protection":                  "deletion_protection",
	"name":                                 "name",
}

func updateMongodbClusterParams(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := getMongoDBClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	var updatePath []string
	for field, path := range mdbMongodbUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	version, err := extractVersion(d)
	if err != nil {
		return err
	}

	if d.HasChange("resources") {
		resourcesSpecPath := fmt.Sprintf("config_spec.mongodb_spec_%s.mongod.resources", flattendVersion(version))
		updatePath = append(updatePath, resourcesSpecPath)
	}

	if d.HasChange("cluster_config.0.mongod") {
		configSpecPath := fmt.Sprintf("config_spec.mongodb_spec_%s.mongod.config", flattendVersion(version))
		updatePath = append(updatePath, configSpecPath)
	}

	if d.HasChange("maintenance_window") {
		mw, err := expandMongoDBMaintenanceWindow(d)
		if err != nil {
			return err
		}
		req.MaintenanceWindow = mw
		updatePath = append(updatePath, "maintenance_window")
	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = d.Get("deletion_protection").(bool)
	}

	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating MongoDB Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func updateMongodbClusterDatabases(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	currDBs, err := listMongodbDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandMongoDBDatabases(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mongodbDatabasesDiff(currDBs, targetDBs)

	for _, db := range toDelete {
		err := deleteMongoDBDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createMongoDBDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateMongodbClusterUsers(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	currUsers, err := listMongodbUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetUsers, err := expandMongoDBUserSpecs(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mongodbUsersDiff(currUsers, targetUsers)
	for _, u := range toDelete {
		err := deleteMongoDBUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}
	for _, u := range toAdd {
		err := createMongoDBUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")
	changedUsers := mongodbChangedUsers(oldSpecs.(*schema.Set), newSpecs.(*schema.Set))
	for _, u := range changedUsers {
		err := updateMongoDBUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func updateMongoDBClusterHosts(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	currHosts, err := listMongodbHosts(ctx, config, d)
	if err != nil {
		return err
	}
	targetHosts, err := expandMongoDBHosts(d)
	if err != nil {
		return err
	}

	currHosts = sortMongoDBHosts(currHosts, targetHosts)

	toDelete, toAdd := mongodbHostsDiff(currHosts, targetHosts)

	for _, hs := range toDelete {
		for _, h := range hs {
			err := deleteMongoDBHost(ctx, config, d, h)
			if err != nil {
				return err
			}
		}
	}

	for _, hs := range toAdd {
		for _, h := range hs {
			err := createMongoDBHost(ctx, config, d, h)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func createMongoDBDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Database().Create(ctx, &mongodb.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &mongodb.DatabaseSpec{
				Name: dbName,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteMongoDBDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Database().Delete(ctx, &mongodb.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMongoDBUser(ctx context.Context, config *Config, d *schema.ResourceData, user *mongodb.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().User().Create(ctx, &mongodb.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteMongoDBUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().User().Delete(ctx, &mongodb.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateMongoDBUser(ctx context.Context, config *Config, d *schema.ResourceData, user *mongodb.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().User().Update(ctx, &mongodb.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMongoDBHost(ctx context.Context, config *Config, d *schema.ResourceData, spec *mongodb.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().AddHosts(ctx, &mongodb.AddClusterHostsRequest{
			ClusterId: d.Id(),
			HostSpecs: []*mongodb.HostSpec{spec},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add host to MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding host to MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteMongoDBHost(ctx context.Context, config *Config, d *schema.ResourceData, fqdn string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MongoDB().Cluster().DeleteHosts(ctx, &mongodb.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: []string{fqdn},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from MongoDB Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from MongoDB Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func mongodbHostsDiff(currHosts []*mongodb.Host, targetHosts []*mongodb.HostSpec) (map[string][]string, map[string][]*mongodb.HostSpec) {
	m := map[string][]*mongodb.HostSpec{}

	for _, h := range targetHosts {
		key := h.Type.String() + h.ZoneId + h.SubnetId
		m[key] = append(m[key], h)
	}

	toDelete := map[string][]string{}
	for _, h := range currHosts {
		key := h.Type.String() + h.ZoneId + h.SubnetId
		hs, ok := m[key]
		if !ok {
			toDelete[h.ShardName] = append(toDelete[h.ShardName], h.Name)
		}
		if len(hs) > 1 {
			m[key] = hs[1:]
		} else {
			delete(m, key)
		}
	}

	toAdd := map[string][]*mongodb.HostSpec{}
	for _, hs := range m {
		for _, h := range hs {
			toAdd[h.ShardName] = append(toAdd[h.ShardName], h)
		}
	}

	return toDelete, toAdd
}

func mongodbUsersDiff(currUsers []*mongodb.User, targetUsers []*mongodb.UserSpec) ([]string, []*mongodb.UserSpec) {
	m := map[string]bool{}
	toDelete := map[string]bool{}
	var toAdd []*mongodb.UserSpec

	for _, u := range currUsers {
		toDelete[u.Name] = true
		m[u.Name] = true
	}

	for _, u := range targetUsers {
		delete(toDelete, u.Name)
		if _, ok := m[u.Name]; !ok {
			toAdd = append(toAdd, u)
		}
	}

	var toDel []string
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func mongodbChangedUsers(oldSpecs *schema.Set, newSpecs *schema.Set) []*mongodb.UserSpec {
	var result []*mongodb.UserSpec
	m := map[string]*mongodb.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user := expandMongoDBUser(spec.(map[string]interface{}))
		m[user.Name] = user
	}
	for _, spec := range newSpecs.List() {
		user := expandMongoDBUser(spec.(map[string]interface{}))
		if u, ok := m[user.Name]; ok {
			if user.Password != u.Password || fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				result = append(result, user)
			}
		}
	}
	return result
}
