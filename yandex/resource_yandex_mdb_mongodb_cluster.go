package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
)

const (
	yandexMDBMongodbClusterDefaultTimeout = 30 * time.Minute
	yandexMDBMongodbClusterUpdateTimeout  = 60 * time.Minute
)

func resourceYandexMDBMongodbCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBMongodbClusterCreate,
		Read:   resourceYandexMDBMongodbClusterRead,
		Update: resourceYandexMDBMongodbClusterUpdate,
		Delete: resourceYandexMDBMongodbClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBMongodbClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBMongodbClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBMongodbClusterDefaultTimeout),
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

	res := mongodb.Resources{
		DiskSize:         toBytes(d.Get("resources.0.disk_size").(int)),
		ResourcePresetId: d.Get("resources.0.resource_preset_id").(string),
		DiskTypeId:       d.Get("resources.0.disk_type_id").(string),
	}

	cfgVer := d.Get("cluster_config.0.version")
	configSpec := &mongodb.ConfigSpec{Version: cfgVer.(string), FeatureCompatibilityVersion: cfgVer.(string)}
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

	switch ver := cfgVer.(string); ver {
	case "5.0":
		{
			configSpec.MongodbSpec = &mongodb.ConfigSpec_MongodbSpec_5_0{
				MongodbSpec_5_0: &mongodb.MongodbSpec5_0{
					Mongod: &mongodb.MongodbSpec5_0_Mongod{
						Resources: &res,
					},
					Mongos: &mongodb.MongodbSpec5_0_Mongos{
						Resources: &res,
					},
					Mongocfg: &mongodb.MongodbSpec5_0_MongoCfg{
						Resources: &res,
					},
				},
			}
		}
	case "4.4":
		{
			configSpec.MongodbSpec = &mongodb.ConfigSpec_MongodbSpec_4_4{
				MongodbSpec_4_4: &mongodb.MongodbSpec4_4{
					Mongod: &mongodb.MongodbSpec4_4_Mongod{
						Resources: &res,
					},
					Mongos: &mongodb.MongodbSpec4_4_Mongos{
						Resources: &res,
					},
					Mongocfg: &mongodb.MongodbSpec4_4_MongoCfg{
						Resources: &res,
					},
				},
			}
		}
	case "4.2":
		{
			configSpec.MongodbSpec = &mongodb.ConfigSpec_MongodbSpec_4_2{
				MongodbSpec_4_2: &mongodb.MongodbSpec4_2{
					Mongod: &mongodb.MongodbSpec4_2_Mongod{
						Resources: &res,
					},
					Mongos: &mongodb.MongodbSpec4_2_Mongos{
						Resources: &res,
					},
					Mongocfg: &mongodb.MongodbSpec4_2_MongoCfg{
						Resources: &res,
					},
				},
			}
		}
	case "4.0":
		{
			configSpec.MongodbSpec = &mongodb.ConfigSpec_MongodbSpec_4_0{
				MongodbSpec_4_0: &mongodb.MongodbSpec4_0{
					Mongod: &mongodb.MongodbSpec4_0_Mongod{
						Resources: &res,
					},
					Mongos: &mongodb.MongodbSpec4_0_Mongos{
						Resources: &res,
					},
					Mongocfg: &mongodb.MongodbSpec4_0_MongoCfg{
						Resources: &res,
					},
				},
			}
		}
	case "3.6":
		{
			configSpec.MongodbSpec = &mongodb.ConfigSpec_MongodbSpec_3_6{
				MongodbSpec_3_6: &mongodb.MongodbSpec3_6{
					Mongod: &mongodb.MongodbSpec3_6_Mongod{
						Resources: &res,
					},
					Mongos: &mongodb.MongodbSpec3_6_Mongos{
						Resources: &res,
					},
					Mongocfg: &mongodb.MongodbSpec3_6_MongoCfg{
						Resources: &res,
					},
				},
			}
		}
	}

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

func resourceYandexMDBMongodbClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreateMongodbRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create Mongodb Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while get Mongodb create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*mongodb.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create Mongodb Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Mongodb Cluster creation failed: %s", err)
	}

	mw, err := expandMongoDBMaintenanceWindow(d)
	if err != nil {
		return err
	}
	if mw != nil {
		err = updateMongoDBMaintenanceWindow(ctx, config, d, mw)
		if err != nil {
			return err
		}
	}

	return resourceYandexMDBMongodbClusterRead(d, meta)
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
	hosts := []*mongodb.Host{}
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

func resourceYandexMDBMongodbClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().MongoDB().Cluster().Get(ctx, &mongodb.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)
	d.Set("sharded", cluster.Sharded)

	ver := cluster.Config.Version

	mongo := cluster.Config

	resources, err := extractMongodbResources(ver, mongo)
	if err != nil {
		return err
	}

	if err := d.Set("resources", resources); err != nil {
		return err
	}

	dUsers, err := expandMongoDBUserSpecs(d)
	if err != nil {
		return err
	}
	passwords := mongodbUsersPasswords(dUsers)

	usrs, err := listMongodbUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	us := flattenMongoDBUsers(usrs, passwords)

	if err := d.Set("user", us); err != nil {
		return err
	}

	dbases, err := listMongodbDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	dbs := flattenMongoDBDatabases(dbases)

	if err := d.Set("database", dbs); err != nil {
		return err
	}

	conf := extractMongoDBConfig(cluster.Config)

	err = d.Set("cluster_config", []map[string]interface{}{
		{
			"backup_window_start":           []*map[string]interface{}{conf.backupWindowStart},
			"feature_compatibility_version": conf.featureCompatibilityVersion,
			"version":                       conf.version,
			"access": []interface{}{
				map[string]interface{}{
					"data_lens": conf.access.DataLens,
				},
			},
		},
	})

	if err != nil {
		return err
	}

	hosts, err := listMongodbHosts(ctx, config, d)
	if err != nil {
		return err
	}

	dHosts, err := expandMongoDBHosts(d)
	if err != nil {
		return err
	}

	hosts = sortMongoDBHosts(hosts, dHosts)

	hs, err := flattenMongoDBHosts(hosts)
	if err != nil {
		return err
	}

	if err := d.Set("host", hs); err != nil {
		return err
	}

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	mw := flattenMongoDBMaintenanceWindow(cluster.MaintenanceWindow)
	if err := d.Set("maintenance_window", mw); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

	return d.Set("labels", cluster.Labels)
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

func resourceYandexMDBMongodbClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if err := updateMongodbClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("database") {
		if err := updateMongodbClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updateMongodbClusterUsers(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updateMongoDBClusterHosts(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)
	return resourceYandexMDBMongodbClusterRead(d, meta)
}

func resourceYandexMDBMongodbClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting Mongodb Cluster %q", d.Id())

	req := &mongodb.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Mongodb Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting Mongodb Cluster %q", d.Id())
	return nil
}

func extractMongodbResources(version string, mongo *mongodb.ClusterConfig) ([]map[string]interface{}, error) {
	if mongo == nil {
		return nil, nil
	}

	switch version {
	case "5.0":
		{
			mongocfg := mongo.Mongodb.(*mongodb.ClusterConfig_Mongodb_5_0).Mongodb_5_0
			d := mongocfg.Mongod
			if d != nil {
				return flattenMongoDBResources(d.Resources)
			}

			s := mongocfg.Mongos
			if s != nil {
				return flattenMongoDBResources(s.Resources)
			}

			cfg := mongocfg.Mongocfg
			if cfg != nil {
				return flattenMongoDBResources(cfg.Resources)
			}
		}
	case "4.4":
		{
			mongocfg := mongo.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_4).Mongodb_4_4
			d := mongocfg.Mongod
			if d != nil {
				return flattenMongoDBResources(d.Resources)
			}

			s := mongocfg.Mongos
			if s != nil {
				return flattenMongoDBResources(s.Resources)
			}

			cfg := mongocfg.Mongocfg
			if cfg != nil {
				return flattenMongoDBResources(cfg.Resources)
			}
		}
	case "4.2":
		{
			mongocfg := mongo.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_2).Mongodb_4_2
			d := mongocfg.Mongod
			if d != nil {
				return flattenMongoDBResources(d.Resources)
			}

			s := mongocfg.Mongos
			if s != nil {
				return flattenMongoDBResources(s.Resources)
			}

			cfg := mongocfg.Mongocfg
			if cfg != nil {
				return flattenMongoDBResources(cfg.Resources)
			}
		}
	case "4.0":
		{
			mongocfg := mongo.Mongodb.(*mongodb.ClusterConfig_Mongodb_4_0).Mongodb_4_0
			d := mongocfg.Mongod
			if d != nil {
				return flattenMongoDBResources(d.Resources)
			}

			s := mongocfg.Mongos
			if s != nil {
				return flattenMongoDBResources(s.Resources)
			}

			cfg := mongocfg.Mongocfg
			if cfg != nil {
				return flattenMongoDBResources(cfg.Resources)
			}
		}
	case "3.6":
		{
			mongocfg := mongo.Mongodb.(*mongodb.ClusterConfig_Mongodb_3_6).Mongodb_3_6
			d := mongocfg.Mongod
			if d != nil {
				return flattenMongoDBResources(d.Resources)
			}

			s := mongocfg.Mongos
			if s != nil {
				return flattenMongoDBResources(s.Resources)
			}

			cfg := mongocfg.Mongocfg
			if cfg != nil {
				return flattenMongoDBResources(cfg.Resources)
			}
		}
	}

	return nil, fmt.Errorf("unexpected error during resources extraction")
}

func listMongodbUsers(ctx context.Context, config *Config, id string) ([]*mongodb.User, error) {
	users := []*mongodb.User{}
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
	dbs := []*mongodb.Database{}
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

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	switch d.Get("cluster_config.0.version").(string) {
	case "5.0":
		{
			req := &mongodb.UpdateClusterRequest{
				ClusterId:   d.Id(),
				Description: d.Get("description").(string),
				Labels:      labels,
				Name:        d.Get("name").(string),
				ConfigSpec: &mongodb.ConfigSpec{
					Version:           d.Get("cluster_config.0.version").(string),
					MongodbSpec:       expandMongoDBSpec5_0(d),
					BackupWindowStart: expandMongoDBBackupWindowStart(d),
					Access:            &mongodb.Access{DataLens: d.Get("cluster_config.0.access.0.data_lens").(bool)},
				},
				SecurityGroupIds: securityGroupIds,
			}
			return req, nil
		}
	case "4.4":
		{
			req := &mongodb.UpdateClusterRequest{
				ClusterId:   d.Id(),
				Description: d.Get("description").(string),
				Labels:      labels,
				Name:        d.Get("name").(string),
				ConfigSpec: &mongodb.ConfigSpec{
					Version:           d.Get("cluster_config.0.version").(string),
					MongodbSpec:       expandMongoDBSpec4_4(d),
					BackupWindowStart: expandMongoDBBackupWindowStart(d),
					Access:            &mongodb.Access{DataLens: d.Get("cluster_config.0.access.0.data_lens").(bool)},
				},
				SecurityGroupIds: securityGroupIds,
			}
			return req, nil
		}
	case "4.2":
		{
			req := &mongodb.UpdateClusterRequest{
				ClusterId:   d.Id(),
				Description: d.Get("description").(string),
				Labels:      labels,
				Name:        d.Get("name").(string),
				ConfigSpec: &mongodb.ConfigSpec{
					Version:           d.Get("cluster_config.0.version").(string),
					MongodbSpec:       expandMongoDBSpec4_2(d),
					BackupWindowStart: expandMongoDBBackupWindowStart(d),
					Access:            &mongodb.Access{DataLens: d.Get("cluster_config.0.access.0.data_lens").(bool)},
				},
				SecurityGroupIds: securityGroupIds,
			}
			return req, nil
		}
	case "4.0":
		{
			req := &mongodb.UpdateClusterRequest{
				ClusterId:   d.Id(),
				Description: d.Get("description").(string),
				Labels:      labels,
				Name:        d.Get("name").(string),
				ConfigSpec: &mongodb.ConfigSpec{
					Version:           d.Get("cluster_config.0.version").(string),
					MongodbSpec:       expandMongoDBSpec4_0(d),
					BackupWindowStart: expandMongoDBBackupWindowStart(d),
					Access:            &mongodb.Access{DataLens: d.Get("cluster_config.0.access.0.data_lens").(bool)},
				},
				SecurityGroupIds: securityGroupIds,
			}
			return req, nil
		}
	case "3.6":
		{
			req := &mongodb.UpdateClusterRequest{
				ClusterId:   d.Id(),
				Description: d.Get("description").(string),
				Labels:      labels,
				Name:        d.Get("name").(string),
				ConfigSpec: &mongodb.ConfigSpec{
					Version:           d.Get("cluster_config.0.version").(string),
					MongodbSpec:       expandMongoDBSpec3_6(d),
					BackupWindowStart: expandMongoDBBackupWindowStart(d),
					Access:            &mongodb.Access{DataLens: d.Get("cluster_config.0.access.0.data_lens").(bool)},
				},
				SecurityGroupIds: securityGroupIds,
			}
			return req, nil
		}
	default:
		{
			return nil, fmt.Errorf("wrong MongoDB version: required either 5.0, 4.4, 4.2, 4.0 or 3.6, got %s", d.Get("cluster_config.version"))
		}
	}
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

func updateMongodbClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := getMongoDBClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbMongodbUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
			onDone = append(onDone, func() {

			})
		}
	}

	if d.HasChange("resources") {
		switch d.Get("cluster_config.0.version").(string) {
		case "5.0":
			{
				updatePath = append(updatePath, "config_spec.mongodb_spec_5_0")
			}
		case "4.4":
			{
				updatePath = append(updatePath, "config_spec.mongodb_spec_4_4")
			}
		case "4.2":
			{
				updatePath = append(updatePath, "config_spec.mongodb_spec_4_2")
			}
		case "4.0":
			{
				updatePath = append(updatePath, "config_spec.mongodb_spec_4_0")
			}
		case "3.6":
			{
				updatePath = append(updatePath, "config_spec.mongodb_spec_3_6")
			}
		default:
			{
				return fmt.Errorf("wrong MongoDB version: required either 5.0, 4.4, 4.2, 4.0 or 3.6, got %s", d.Get("cluster_config.version"))
			}
		}

		if d.HasChange("resources.0.disk_size") {
			//updatePath = append(updatePath, "config_spec.mongodb_spec.resources.disk_size")
			onDone = append(onDone, func() {

			})
		}

		if d.HasChange("resources.0.disk_type_id") {
			//updatePath = append(updatePath, "config_spec.mongodb_spec.resources.disk_type_id")
			onDone = append(onDone, func() {

			})
		}

		if d.HasChange("resources.0.resource_preset_id") {
			//updatePath = append(updatePath, "config_spec.mongodb_spec.resources.resource_preset_id")
			onDone = append(onDone, func() {

			})
		}
	}

	if d.HasChange("maintenance_window") {
		mw, err := expandMongoDBMaintenanceWindow(d)
		if err != nil {
			return err
		}
		req.MaintenanceWindow = mw
		updatePath = append(updatePath, "maintenance_window")

		onDone = append(onDone, func() {

		})
	}

	if d.HasChange("deletion_protection") {
		req.DeletionProtection = d.Get("deletion_protection").(bool)
	}

	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MongoDB().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update MongoDB Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating MongoDB Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func updateMongodbClusterDatabases(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

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

func updateMongodbClusterUsers(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()
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

func updateMongoDBClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

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
	toAdd := []*mongodb.UserSpec{}

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

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func mongodbChangedUsers(oldSpecs *schema.Set, newSpecs *schema.Set) []*mongodb.UserSpec {
	result := []*mongodb.UserSpec{}
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
