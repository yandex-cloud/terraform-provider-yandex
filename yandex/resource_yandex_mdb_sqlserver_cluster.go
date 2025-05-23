package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/sqlserver/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	yandexMDBSQLServerClusterDefaultTimeout = 60 * time.Minute
	yandexMDBSQLServerClusterUpdateTimeout  = 120 * time.Minute
)

func resourceYandexMDBSQLServerCluster() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a SQLServer cluster within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-sqlserver/).\n\nPlease read [Pricing for Managed Service for SQL Server](https://yandex.cloud/docs/managed-sqlserver/pricing#prices) before using SQLServer cluster.\n",

		Create: resourceYandexMDBSQLServerClusterCreate,
		Read:   resourceYandexMDBSQLServerClusterRead,
		Update: resourceYandexMDBSQLServerClusterUpdate,
		Delete: resourceYandexMDBSQLServerClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBSQLServerClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBSQLServerClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBSQLServerClusterDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["name"],
				Required:    true,
			},
			"folder_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["folder_id"],
				Computed:    true,
				Optional:    true,
				ForceNew:    true,
			},
			"environment": {
				Type:         schema.TypeString,
				Description:  "Deployment environment of the SQLServer cluster. (PRODUCTION, PRESTABLE).",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseSQLServerEnv),
			},
			"network_id": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["network_id"],
				Required:    true,
				ForceNew:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Description: "Version of the SQLServer cluster. (2016sp2std, 2016sp2ent).",
				Required:    true,
			},
			"resources": {
				Type:        schema.TypeList,
				Description: "Resources allocated to hosts of the SQLServer cluster.",
				Required:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_preset_id": {
							Type:        schema.TypeString,
							Description: "The ID of the preset for computational resources available to a SQLServer host (CPU, memory etc.). For more information, see [the official documentation](https://yandex.cloud/docs/managed-sqlserver/concepts/instance-types).",
							Required:    true,
						},
						"disk_type_id": {
							Type:        schema.TypeString,
							Description: "Type of the storage of SQLServer hosts.",
							Required:    true,
						},
						"disk_size": {
							Type:        schema.TypeInt,
							Description: "Volume of the storage available to a SQLServer host, in gigabytes.",
							Required:    true,
						},
					},
				},
			},
			"database": {
				Type:        schema.TypeList,
				Description: "A database of the SQLServer cluster. ",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the database.",
							Required:    true,
						},
					},
				},
			},
			"user": {
				Type:        schema.TypeList,
				Description: "A user of the SQLServer cluster.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The name of the user.",
							Required:    true,
						},
						"password": {
							Type:        schema.TypeString,
							Description: "The password of the user.",
							Required:    true,
							Sensitive:   true,
						},
						"permission": {
							Type:        schema.TypeSet,
							Description: "Set of permissions granted to the user.",
							Optional:    true,
							Set:         sqlserverUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:        schema.TypeString,
										Description: "The name of the database that the permission grants access to.",
										Required:    true,
									},
									"roles": {
										Type:        schema.TypeSet,
										Description: "List user's roles in the database. Allowed roles: `OWNER`, `SECURITYADMIN`, `ACCESSADMIN`, `BACKUPOPERATOR`, `DDLADMIN`, `DATAWRITER`, `DATAREADER`, `DENYDATAWRITER`, `DENYDATAREADER`.",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"host": {
				Type:        schema.TypeList,
				Description: "A host of the SQLServer cluster.",
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:        schema.TypeString,
							Description: common.ResourceDescriptions["zone"],
							Required:    true,
						},
						"assign_public_ip": {
							Type:        schema.TypeBool,
							Description: "Sets whether the host should get a public IP address on creation. Changing this parameter for an existing host is not supported at the moment.",
							Optional:    true,
							Default:     false,
						},
						"subnet_id": {
							Type:        schema.TypeString,
							Description: "The ID of the subnet, to which the host belongs. The subnet must be a part of the network to which the cluster belongs.",
							Optional:    true,
							Computed:    true,
						},
						"fqdn": {
							Type:        schema.TypeString,
							Description: "The fully qualified domain name of the host.",
							Computed:    true,
						},
					},
				},
			},
			"host_group_ids": {
				Type:        schema.TypeSet,
				Description: "A list of IDs of the host groups hosting VMs of the cluster.",
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
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
				Set:         schema.HashString,
			},
			"backup_window_start": {
				Type:        schema.TypeList,
				Description: "Time to start the daily backup, in the UTC.",
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hours": {
							Type:         schema.TypeInt,
							Description:  "The hour at which backup will be started.",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 23),
						},
						"minutes": {
							Type:         schema.TypeInt,
							Description:  "The minute at which backup will be started.",
							Optional:     true,
							Default:      0,
							ValidateFunc: validation.IntBetween(0, 59),
						},
					},
				},
			},

			"created_at": {
				Type:        schema.TypeString,
				Description: common.ResourceDescriptions["created_at"],
				Computed:    true,
			},
			"health": {
				Type:        schema.TypeString,
				Description: "Aggregated health of the cluster.",
				Computed:    true,
			},
			"status": {
				Type:        schema.TypeString,
				Description: "Status of the cluster.",
				Computed:    true,
			},
			"security_group_ids": {
				Type:        schema.TypeSet,
				Description: common.ResourceDescriptions["security_group_ids"],
				Elem:        &schema.Schema{Type: schema.TypeString},
				Set:         schema.HashString,
				Optional:    true,
			},
			"sqlserver_config": {
				Type:             schema.TypeMap,
				Description:      "SQLServer cluster config. Detail info in `SQLServer config` section.",
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbSQLServerSettingsFieldsInfo),
				ValidateFunc:     generateMapSchemaValidateFunc(mdbSQLServerSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"deletion_protection": {
				Type:        schema.TypeBool,
				Description: common.ResourceDescriptions["deletion_protection"],
				Optional:    true,
				Computed:    true,
			},
			"sqlcollation": {
				Type:        schema.TypeString,
				Description: "SQL Collation cluster will be created with. This attribute cannot be changed when cluster is created!",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceYandexMDBSQLServerClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := prepareCreateSQLServerRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.MDB().SQLServer().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create SQLServer Cluster: %s", err)
	}
	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get SQLServer create operation metadata: %s", err)
	}
	md, ok := protoMetadata.(*sqlserver.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get SQLServer Cluster ID from create operation metadata")
	}
	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create SQLServer Cluster: %s", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("SQLServer Cluster creation failed: %s", err)
	}
	return resourceYandexMDBSQLServerClusterRead(d, meta)
}

func prepareCreateSQLServerRequest(d *schema.ResourceData, meta *Config) (*sqlserver.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, fmt.Errorf("Error while expanding labels on SQLServer Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating SQLServer Cluster: %s", err)
	}

	hosts, err := expandSQLServerHosts(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding hosts on SQLServer Cluster create: %s", err)
	}
	e := d.Get("environment").(string)
	env, err := parseSQLServerEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating SQLServer Cluster: %s", err)
	}

	backupWindowStart := expandMDBBackupWindowStart(d, "backup_window_start.0")

	resources := expandSQLServerResources(d)

	userSpecs, err := expandSQLServerUserSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding users on SQLServer Cluster create: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	hostGroupIds := expandHostGroupIds(d.Get("host_group_ids"))

	dbs := expandSQLServerDatabaseSpecs(d)

	configSpec := &sqlserver.ConfigSpec{
		Version:           d.Get("version").(string),
		Resources:         resources,
		BackupWindowStart: backupWindowStart,
	}

	_, _, err = expandSQLServerConfigSpecSettings(d, configSpec)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding sqlserver_config on SQLServer Cluster create: %s", err)
	}

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on SQLServer Cluster create: %s", err)
	}

	req := sqlserver.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Environment:        env,
		ConfigSpec:         configSpec,
		DatabaseSpecs:      dbs,
		UserSpecs:          userSpecs,
		HostSpecs:          hosts,
		Labels:             labels,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
		Sqlcollation:       d.Get("sqlcollation").(string),
		HostGroupIds:       hostGroupIds,
	}
	return &req, nil
}

func resourceYandexMDBSQLServerClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().SQLServer().Cluster().Get(ctx, &sqlserver.GetClusterRequest{
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

	if err := d.Set("labels", cluster.Labels); err != nil {
		return err
	}

	if err := d.Set("resources", flattenSQLServerResources(cluster.Config.Resources)); err != nil {
		return err
	}

	usersSpec, err := listSQLServerUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	passwords := expandSQLServerUserPasswords(d)

	users, err := flattenSQLServerUsers(usersSpec, passwords)

	if err != nil {
		return err
	}

	sortInterfaceListByResourceData(users, d, "user", "name")

	if err = d.Set("user", users); err != nil {
		return err
	}
	if err = d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}
	if err = d.Set("host_group_ids", cluster.HostGroupIds); err != nil {
		return err
	}

	hostsSpec, err := listSQLServerHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	hosts, err := flattenSQLServerHosts(d, hostsSpec)
	if err != nil {
		return err
	}
	if err = d.Set("host", hosts); err != nil {
		return err
	}

	databasesSpec, err := listSQLServerDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	databases := flattenSQLServerDatabases(databasesSpec)

	sortInterfaceListByResourceData(databases, d, "database", "name")

	if err = d.Set("database", databases); err != nil {
		return err
	}

	backupWindowStart := flattenMDBBackupWindowStart(cluster.GetConfig().GetBackupWindowStart())
	if err = d.Set("backup_window_start", backupWindowStart); err != nil {
		return err
	}

	clusterConfig, err := flattenSQLServerSettings(cluster.Config)
	if err != nil {
		return err
	}

	if err := d.Set("sqlserver_config", clusterConfig); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)
	d.Set("sqlcollation", cluster.Sqlcollation)

	return d.Set("created_at", getTimestamp(cluster.CreatedAt))
}

func listSQLServerUsers(ctx context.Context, config *Config, id string) ([]*sqlserver.User, error) {
	users := []*sqlserver.User{}
	pageToken := ""
	for {
		resp, err := config.sdk.MDB().SQLServer().User().List(ctx, &sqlserver.ListUsersRequest{
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

func listSQLServerHosts(ctx context.Context, config *Config, id string) ([]*sqlserver.Host, error) {
	hosts := []*sqlserver.Host{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().SQLServer().Cluster().ListHosts(ctx, &sqlserver.ListClusterHostsRequest{
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

func listSQLServerDatabases(ctx context.Context, config *Config, id string) ([]*sqlserver.Database, error) {
	databases := []*sqlserver.Database{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().SQLServer().Database().List(ctx, &sqlserver.ListDatabasesRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of databases for '%s': %s", id, err)
		}

		databases = append(databases, resp.Databases...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return databases, nil
}

func resourceYandexMDBSQLServerClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	if d.HasChange("version") {
		return fmt.Errorf("Changing version is not supported for SQLServer Cluster. Id: %v", d.Id())
	}

	if d.HasChange("host") {
		return fmt.Errorf("Changing hosts is not supported for SQLServer Cluster. Id: %v", d.Id())
	}

	if d.HasChange("resources.0.disk_type_id") {
		return fmt.Errorf("Changing disk_type_id is not supported for SQLServer Cluster. Id: %v", d.Id())
	}

	if err := sqlserverClusterUpdate(ctx, config, d); err != nil {
		return err
	}

	if d.HasChange("database") {
		if err := sqlserverDatabaseUpdate(ctx, config, d); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := sqlserverUserUpdate(ctx, config, d); err != nil {
			return err
		}
	}

	return resourceYandexMDBSQLServerClusterRead(d, meta)
}

var mdbSQLServerUpdateFieldsMap = map[string]string{
	"name":                           "name",
	"description":                    "description",
	"labels":                         "labels",
	"backup_window_start":            "config_spec.backup_window_start",
	"resources.0.resource_preset_id": "config_spec.resources.resource_preset_id",
	"resources.0.disk_size":          "config_spec.resources.disk_size",
	"security_group_ids":             "security_group_ids",
	"deletion_protection":            "deletion_protection",
}

func sqlserverClusterUpdate(ctx context.Context, config *Config, d *schema.ResourceData) error {

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("error expanding labels while updating SQLServer cluster: %s", err)
	}

	sqlcollation := d.Get("sqlcollation").(string)
	if sqlcollation != "" {
		return fmt.Errorf("SQL Collation cannot be changed!")
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	req := &sqlserver.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
		ConfigSpec: &sqlserver.ConfigSpec{
			Resources:         expandSQLServerResources(d),
			Version:           d.Get("version").(string),
			BackupWindowStart: expandMDBBackupWindowStart(d, "backup_window_start.0"),
		},
	}

	updateFieldConfigName, fields, err := expandSQLServerConfigSpecSettings(d, req.ConfigSpec)
	if err != nil {
		return err
	}

	updatePath := []string{}
	for field, path := range mdbSQLServerUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
		}
	}

	if d.HasChange("sqlserver_config") && len(fields) > 0 {
		for _, field := range fields {
			updatePath = append(updatePath, "config_spec."+updateFieldConfigName+"."+field)
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}

	op, err := config.sdk.WrapOperation(config.sdk.MDB().SQLServer().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update SQLServer Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating SQLServer Cluster %q: %s", d.Id(), err)
	}

	return nil
}

func sqlserverUserUpdate(ctx context.Context, config *Config, d *schema.ResourceData) error {
	newUsersSpecs, changedUsersSpecs, dropUserNames, err := usersDiffSQLServer(ctx, config, d)

	if err != nil {
		return err
	}

	for _, user := range newUsersSpecs {
		err = createMDBSQLServerUser(ctx, config, d, user)
		if err != nil {
			return err
		}
	}

	for _, user := range changedUsersSpecs {
		err = updateSQLServerUser(ctx, config, d, user)
		if err != nil {
			return err
		}
	}

	for _, userName := range dropUserNames {
		err = deleteSQLServerUser(ctx, config, d, userName)
		if err != nil {
			return err
		}
	}

	return nil
}

func sqlserverDatabaseUpdate(ctx context.Context, config *Config, d *schema.ResourceData) error {
	newDatabaseSpecs, dropDatabaseNames, err := databaseDiffSQLServer(ctx, config, d)

	if err != nil {
		return err
	}

	for _, db := range newDatabaseSpecs {
		err = createMDBSQLServerDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	for _, dbName := range dropDatabaseNames {
		err = deleteSQLServerDatabase(ctx, config, d, dbName)
		if err != nil {
			return err
		}
	}

	return nil
}

func createMDBSQLServerUser(ctx context.Context, config *Config, d *schema.ResourceData, user *sqlserver.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().SQLServer().User().Create(ctx, &sqlserver.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for SQLServer Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for SQLServer Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateSQLServerUser(ctx context.Context, config *Config, d *schema.ResourceData, user *sqlserver.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().SQLServer().User().Update(ctx, &sqlserver.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in SQLServer Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in SQLServer Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteSQLServerUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().SQLServer().User().Delete(ctx, &sqlserver.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from SQLServer Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from SQLServer Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMDBSQLServerDatabase(ctx context.Context, config *Config, d *schema.ResourceData, db *sqlserver.DatabaseSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().SQLServer().Database().Create(ctx, &sqlserver.CreateDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseSpec: db,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create databse for SQLServer Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating databse for SQLServer Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func deleteSQLServerDatabase(ctx context.Context, config *Config, d *schema.ResourceData, databaseName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().SQLServer().Database().Delete(ctx, &sqlserver.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: databaseName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete databse from SQLServer Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting databse from SQLServer Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func resourceYandexMDBSQLServerClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting SQLServer Cluster %q", d.Id())

	req := &sqlserver.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().SQLServer().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("SQLServer Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting SQLServer Cluster %q", d.Id())
	return nil
}
