package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1"
)

const (
	yandexMDBClickHouseClusterCreateTimeout = 60 * time.Minute
	yandexMDBClickHouseClusterDeleteTimeout = 30 * time.Minute
	yandexMDBClickHouseClusterUpdateTimeout = 90 * time.Minute
)

func resourceYandexMDBClickHouseCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBClickHouseClusterCreate,
		Read:   resourceYandexMDBClickHouseClusterRead,
		Update: resourceYandexMDBClickHouseClusterUpdate,
		Delete: resourceYandexMDBClickHouseClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBClickHouseClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBClickHouseClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBClickHouseClusterDeleteTimeout),
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
				ValidateFunc: validateParsableValue(parseClickHouseEnv),
			},
			"clickhouse": {
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
									},
								},
							},
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      clickHouseUserHash,
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
							Set:      clickHouseUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Required: true,
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
				Set:      clickHouseDatabaseHash,
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
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateParsableValue(parseClickHouseHostType),
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"shard_name": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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
			"version": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.NoZeroValues,
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
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"web_sql": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"data_lens": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"metrika": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"serverless": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"zookeeper": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resources": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_preset_id": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"disk_size": {
										Type:     schema.TypeInt,
										Optional: true,
										Computed: true,
									},
									"disk_type_id": {
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
			"folder_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
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
		},
	}
}

func resourceYandexMDBClickHouseClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, shards, err := prepareCreateClickHouseCreateRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Clickhouse().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to create ClickHouse Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("error while getting ClickHouse create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*clickhouse.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("could not get Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to create ClickHouse Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("ClickHouse Cluster creation failed: %s", err)
	}

	for shardName, shardHosts := range shards {
		err = createClickHouseShard(ctx, config, d, shardName, shardHosts)
		if err != nil {
			return err
		}
	}

	return resourceYandexMDBClickHouseClusterRead(d, meta)
}

// Returns request for creating the Cluster and the map of the remaining shards to add.
func prepareCreateClickHouseCreateRequest(d *schema.ResourceData, meta *Config) (*clickhouse.CreateClusterRequest, map[string][]*clickhouse.HostSpec, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding labels on ClickHouse Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, nil, fmt.Errorf("Error getting folder ID while creating ClickHouse Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseClickHouseEnv(e)
	if err != nil {
		return nil, nil, fmt.Errorf("Error resolving environment while creating ClickHouse Cluster: %s", err)
	}

	dbSpecs, err := expandClickHouseDatabases(d)
	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding databases on ClickHouse Cluster create: %s", err)
	}

	users, err := expandClickHouseUserSpecs(d)
	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding user specs on ClickHouse Cluster create: %s", err)
	}

	hosts, err := expandClickHouseHosts(d)
	if err != nil {
		return nil, nil, fmt.Errorf("error while expanding hosts on ClickHouse Cluster create: %s", err)
	}

	_, toAdd := clickHouseHostsDiff(nil, hosts)
	firstHosts := toAdd["zk"]
	firstShard := ""
	delete(toAdd, "zk")
	for shardName, shardHosts := range toAdd {
		firstShard = shardName
		firstHosts = append(firstHosts, shardHosts...)
		delete(toAdd, shardName)
		break
	}

	configSpec := &clickhouse.ConfigSpec{
		Version:           d.Get("version").(string),
		Clickhouse:        expandClickHouseSpec(d),
		Zookeeper:         expandClickHouseZookeeperSpec(d),
		BackupWindowStart: expandClickHouseBackupWindowStart(d),
		Access:            expandClickHouseAccess(d),
	}

	req := clickhouse.CreateClusterRequest{
		FolderId:      folderID,
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		NetworkId:     d.Get("network_id").(string),
		Environment:   env,
		DatabaseSpecs: dbSpecs,
		ConfigSpec:    configSpec,
		HostSpecs:     firstHosts,
		UserSpecs:     users,
		Labels:        labels,
		ShardName:     firstShard,
	}
	return &req, toAdd, nil
}

func resourceYandexMDBClickHouseClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().Clickhouse().Cluster().Get(ctx, &clickhouse.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Get("name").(string)))
	}
	chResources, err := flattenClickHouseResources(cluster.Config.Clickhouse.Resources)
	if err != nil {
		return err
	}
	err = d.Set("clickhouse", []map[string]interface{}{
		{
			"resources": chResources,
		},
	})
	if err != nil {
		return err
	}

	zkResources, err := flattenClickHouseResources(cluster.Config.Zookeeper.Resources)
	if err != nil {
		return err
	}
	err = d.Set("zookeeper", []map[string]interface{}{
		{
			"resources": zkResources,
		},
	})
	if err != nil {
		return err
	}

	bws := flattenClickHouseBackupWindowStart(cluster.Config.BackupWindowStart)
	if err := d.Set("backup_window_start", bws); err != nil {
		return err
	}

	ac := flattenClickHouseAccess(cluster.Config.Access)
	if err := d.Set("access", ac); err != nil {
		return err
	}

	hosts, err := listClickHouseHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	dHosts, err := expandClickHouseHosts(d)
	if err != nil {
		return err
	}

	hosts = sortClickHouseHosts(hosts, dHosts)
	hs, err := flattenClickHouseHosts(hosts)
	if err != nil {
		return err
	}

	if err := d.Set("host", hs); err != nil {
		return err
	}

	databases, err := listClickHouseDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}
	dbs := flattenClickHouseDatabases(databases)
	if err := d.Set("database", dbs); err != nil {
		return err
	}

	dUsers, err := expandClickHouseUserSpecs(d)
	if err != nil {
		return err
	}
	passwords := clickHouseUsersPasswords(dUsers)

	users, err := listClickHouseUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	us := flattenClickHouseUsers(users, passwords)
	if err := d.Set("user", us); err != nil {
		return err
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("name", cluster.Name)
	d.Set("folder_id", cluster.FolderId)
	d.Set("network_id", cluster.NetworkId)
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("description", cluster.Description)

	return d.Set("labels", cluster.Labels)
}

func resourceYandexMDBClickHouseClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating ClickHouse Cluster %q", d.Id())

	d.Partial(true)

	if err := updateClickHouseClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("database") {
		if err := updateClickHouseClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updateClickHouseClusterUsers(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updateClickHouseClusterHosts(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)

	log.Printf("[DEBUG] Finished updating ClickHouse Cluster %q", d.Id())
	return resourceYandexMDBClickHouseClusterRead(d, meta)
}

var mdbClickHouseUpdateFieldsMap = map[string]string{
	"name":                "name",
	"description":         "description",
	"labels":              "labels",
	"version":             "config_spec.version",
	"access":              "config_spec.access",
	"backup_window_start": "config_spec.backup_window_start",
	"clickhouse":          "config_spec.clickhouse",
}

func updateClickHouseClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := getClickHouseClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbClickHouseUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
			onDone = append(onDone, func() {
				d.SetPartial(field)
			})
		}
	}

	// We only can apply this if ZK subcluster already exists
	if d.HasChange("zookeeper") {
		ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
		defer cancel()

		currHosts, err := listClickHouseHosts(ctx, config, d.Id())
		if err != nil {
			return err
		}

		for _, h := range currHosts {
			if h.Type == clickhouse.Host_ZOOKEEPER {
				updatePath = append(updatePath, "config_spec.zookeeper")
				onDone = append(onDone, func() {
					d.SetPartial("zookeeper")
				})
				break
			}
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Clickhouse().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating ClickHouse Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func getClickHouseClusterUpdateRequest(d *schema.ResourceData) (*clickhouse.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating ClickHouse cluster: %s", err)
	}

	req := &clickhouse.UpdateClusterRequest{
		ClusterId:   d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		ConfigSpec: &clickhouse.ConfigSpec{
			Version:           d.Get("version").(string),
			Clickhouse:        expandClickHouseSpec(d),
			Zookeeper:         expandClickHouseZookeeperSpec(d),
			BackupWindowStart: expandClickHouseBackupWindowStart(d),
			Access:            expandClickHouseAccess(d),
		},
	}
	return req, nil
}

func updateClickHouseClusterDatabases(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currDBs, err := listClickHouseDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandClickHouseDatabases(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := clickHouseDatabasesDiff(currDBs, targetDBs)

	for _, db := range toDelete {
		err := deleteClickHouseDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createClickHouseDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	d.SetPartial("database")
	return nil
}

func updateClickHouseClusterUsers(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	currUsers, err := listClickHouseUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetUsers, err := expandClickHouseUserSpecs(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := clickHouseUsersDiff(currUsers, targetUsers)
	for _, u := range toDelete {
		err := deleteClickHouseUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}
	for _, u := range toAdd {
		err := createClickHouseUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")
	changedUsers := clickHouseChangedUsers(oldSpecs.(*schema.Set), newSpecs.(*schema.Set))
	for _, u := range changedUsers {
		err := updateClickHouseUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	d.SetPartial("user")
	return nil
}

func updateClickHouseClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currHosts, err := listClickHouseHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetHosts, err := expandClickHouseHosts(d)
	if err != nil {
		return err
	}
	currZkHosts := []*clickhouse.Host{}
	for _, h := range currHosts {
		if h.Type == clickhouse.Host_ZOOKEEPER {
			currZkHosts = append(currZkHosts, h)
		}
	}
	targetZkHosts := []*clickhouse.HostSpec{}
	for _, h := range targetHosts {
		if h.Type == clickhouse.Host_ZOOKEEPER {
			targetZkHosts = append(targetZkHosts, h)
		}
	}

	toDelete, toAdd := clickHouseHostsDiff(currHosts, targetHosts)

	// Check if any shard has HA-configuration (2+ hosts)
	needZk := false
	m := map[string][]struct{}{}
	for _, h := range targetHosts {
		if h.Type == clickhouse.Host_CLICKHOUSE {
			shardName := "shard1"
			if h.ShardName != "" {
				shardName = h.ShardName
			}
			m[shardName] = append(m[shardName], struct{}{})
			if len(m[shardName]) > 1 {
				needZk = true
				break
			}
		}
	}

	// We need to create a ZooKeeper subcluster first
	if len(currZkHosts) == 0 && (needZk || len(toAdd["zk"]) > 0) {
		zkSpecs := toAdd["zk"]
		delete(toAdd, "zk")
		zk := expandClickHouseZookeeperSpec(d)

		err = createClickHouseZooKeeper(ctx, config, d, zk.Resources, zkSpecs)
		if err != nil {
			return err
		}
	}

	// Do not remove implicit ZooKeeper subcluster.
	if len(currZkHosts) > 1 && len(targetZkHosts) == 0 {
		delete(toDelete, "zk")
	}

	currShards, err := listClickHouseShards(ctx, config, d.Id())
	if err != nil {
		return err
	}

	for shardName, specs := range toAdd {
		shardExists := false
		for _, s := range currShards {
			if s.Name == shardName {
				shardExists = true
			}
		}

		if shardName != "" && shardName != "zk" && !shardExists {
			err = createClickHouseShard(ctx, config, d, shardName, specs)
			if err != nil {
				return err
			}
		} else {
			for _, h := range specs {
				err := createClickHouseHost(ctx, config, d, h)
				if err != nil {
					return err
				}
			}
		}
	}

	for shardName, fqdns := range toDelete {
		deleteShard := true
		for _, th := range targetHosts {
			if th.ShardName == shardName {
				deleteShard = false
			}
		}
		if shardName != "zk" && deleteShard {
			err = deleteClickHouseShard(ctx, config, d, shardName)
			if err != nil {
				return err
			}
		} else {
			for _, h := range fqdns {
				err := deleteClickHouseHost(ctx, config, d, h)
				if err != nil {
					return err
				}
			}
		}
	}

	d.SetPartial("host")
	return nil
}

func createClickHouseDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Database().Create(ctx, &clickhouse.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &clickhouse.DatabaseSpec{
				Name: dbName,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Database().Delete(ctx, &clickhouse.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, user *clickhouse.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().User().Create(ctx, &clickhouse.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().User().Delete(ctx, &clickhouse.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateClickHouseUser(ctx context.Context, config *Config, d *schema.ResourceData, user *clickhouse.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().User().Update(ctx, &clickhouse.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in ClickHouse Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseHost(ctx context.Context, config *Config, d *schema.ResourceData, spec *clickhouse.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().AddHosts(ctx, &clickhouse.AddClusterHostsRequest{
			ClusterId: d.Id(),
			HostSpecs: []*clickhouse.HostSpec{spec},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add host to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding host to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseHost(ctx context.Context, config *Config, d *schema.ResourceData, fqdn string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().DeleteHosts(ctx, &clickhouse.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: []string{fqdn},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseShard(ctx context.Context, config *Config, d *schema.ResourceData, name string, specs []*clickhouse.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().AddShard(ctx, &clickhouse.AddClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: name,
			HostSpecs: specs,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to add shard to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding shard to ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteClickHouseShard(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().DeleteShard(ctx, &clickhouse.DeleteClusterShardRequest{
			ClusterId: d.Id(),
			ShardName: name,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete shard from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting shard from ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createClickHouseZooKeeper(ctx context.Context, config *Config, d *schema.ResourceData, resources *clickhouse.Resources, specs []*clickhouse.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().Clickhouse().Cluster().AddZookeeper(ctx, &clickhouse.AddClusterZookeeperRequest{
			ClusterId: d.Id(),
			Resources: resources,
			HostSpecs: specs,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create ZooKeeper subcluster in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating ZooKeeper subcluster in ClickHouse Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func listClickHouseHosts(ctx context.Context, config *Config, id string) ([]*clickhouse.Host, error) {
	hosts := []*clickhouse.Host{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().Cluster().ListHosts(ctx, &clickhouse.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of hosts for '%s': %s", id, err)
		}
		hosts = append(hosts, resp.Hosts...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return hosts, nil
}

func listClickHouseUsers(ctx context.Context, config *Config, id string) ([]*clickhouse.User, error) {
	users := []*clickhouse.User{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().User().List(ctx, &clickhouse.ListUsersRequest{
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

func listClickHouseDatabases(ctx context.Context, config *Config, id string) ([]*clickhouse.Database, error) {
	dbs := []*clickhouse.Database{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().Database().List(ctx, &clickhouse.ListDatabasesRequest{
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

func listClickHouseShards(ctx context.Context, config *Config, id string) ([]*clickhouse.Shard, error) {
	shards := []*clickhouse.Shard{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().Clickhouse().Cluster().ListShards(ctx, &clickhouse.ListClusterShardsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error while getting list of shards for '%s': %s", id, err)
		}
		shards = append(shards, resp.Shards...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return shards, nil
}

func resourceYandexMDBClickHouseClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting ClickHouse Cluster %q", d.Id())

	req := &clickhouse.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().Clickhouse().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("ClickHouse Cluster %q", d.Get("name").(string)))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting ClickHouse Cluster %q", d.Id())
	return nil
}
