package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
)

const (
	yandexMDBMySQLClusterDefaultTimeout = 15 * time.Minute
	yandexMDBMySQLClusterUpdateTimeout  = 60 * time.Minute
)

func resourceYandexMDBMySQLCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBMySQLClusterCreate,
		Read:   resourceYandexMDBMySQLClusterRead,
		Update: resourceYandexMDBMySQLClusterUpdate,
		Delete: resourceYandexMDBMySQLClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBMySQLClusterDefaultTimeout),
			Update: schema.DefaultTimeout(yandexMDBMySQLClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBMySQLClusterDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateParsableValue(parseMysqlEnv),
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"version": {
				Type:     schema.TypeString,
				Required: true,
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
			"database": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      mysqlDatabaseHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeSet,
				Required: true,
				Set:      mysqlUserHash,
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
							Set:      mysqlUserPermissionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"database_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"roles": {
										Type: schema.TypeList,
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
				Type:     schema.TypeList,
				MinItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:     schema.TypeString,
							Required: true,
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

func resourceYandexMDBMySQLClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := prepareCreateMySQLRequest(d, config)
	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()
	op, err := config.sdk.WrapOperation(config.sdk.MDB().MySQL().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create MySQL Cluster: %s", err)
	}
	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get MySQL create operation metadata: %s", err)
	}
	md, ok := protoMetadata.(*mysql.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get MySQL Cluster ID from create operation metadata")
	}
	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create MySQL Cluster: %s", err)
	}
	if _, err := op.Response(); err != nil {
		return fmt.Errorf("MySQL Cluster creation failed: %s", err)
	}
	return resourceYandexMDBMySQLClusterRead(d, meta)
}

func prepareCreateMySQLRequest(d *schema.ResourceData, meta *Config) (*mysql.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))

	if err != nil {
		return nil, fmt.Errorf("Error while expanding labels on MySQL Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating MySQL Cluster: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parseMysqlEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating MySQL Cluster: %s", err)
	}

	resources := expandMysqlResources(d)

	backupWindowStart := expandMysqlBackupWindowStart(d)

	dbSpecs, err := expandMysqlDatabases(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding databases on Mysql Cluster create: %s", err)
	}

	hostsFromScheme, err := expandMysqlHosts(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding hostsFromScheme on MySQL Cluster create: %s", err)
	}

	users, err := expandMysqlUserSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("error while expanding user specs on MySQL Cluster create: %s", err)
	}

	version := d.Get("version").(string)
	configSpec := &mysql.ConfigSpec{
		Version:           version,
		Resources:         resources,
		BackupWindowStart: backupWindowStart,
	}

	hostSpecs := make([]*mysql.HostSpec, 0)
	for _, host := range hostsFromScheme {
		hostSpecs = append(hostSpecs, host.HostSpec)
	}
	req := mysql.CreateClusterRequest{
		FolderId:      folderID,
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		NetworkId:     d.Get("network_id").(string),
		Environment:   env,
		ConfigSpec:    configSpec,
		DatabaseSpecs: dbSpecs,
		UserSpecs:     users,
		HostSpecs:     hostSpecs,
		Labels:        labels,
	}
	return &req, nil
}

func resourceYandexMDBMySQLClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().MySQL().Cluster().Get(ctx, &mysql.GetClusterRequest{
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

	hosts, err := listMysqlHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	hostSpecs, err := expandMysqlHosts(d)
	if err != nil {
		return err
	}

	sortMysqlHosts(hosts, hostSpecs)

	fHosts, err := flattenMysqlHosts(hosts)

	if err != nil {
		return err
	}

	if err := d.Set("host", fHosts); err != nil {
		return err
	}

	userSpecs, err := expandMysqlUserSpecs(d)
	if err != nil {
		return err
	}
	passwords := mysqlUsersPasswords(userSpecs)
	users, err := listMysqlUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	fUsers, err := flattenMysqlUsers(users, passwords)
	if err != nil {
		return err
	}

	if err := d.Set("user", fUsers); err != nil {
		return err
	}

	databases, err := listMysqlDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	fDatabases := flattenMysqlDatabases(databases)
	if err := d.Set("database", fDatabases); err != nil {
		return err
	}

	mysqlResources, err := flattenMysqlResources(cluster.GetConfig().GetResources())
	if err != nil {
		return err
	}
	err = d.Set("resources", mysqlResources)
	if err != nil {
		return err
	}

	backupWindowStart, err := flattenMysqlBackupWindowStart(cluster.GetConfig().GetBackupWindowStart())
	if err != nil {
		return err
	}
	if err := d.Set("backup_window_start", backupWindowStart); err != nil {
		return err
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	return d.Set("created_at", createdAt)
}

func resourceYandexMDBMySQLClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if err := updateMysqlClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("database") {
		if err := updateMysqlClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updateMysqlClusterUsers(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updateMysqlClusterHosts(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)
	return resourceYandexMDBMySQLClusterRead(d, meta)
}

var mdbMysqlUpdateFieldsMap = map[string]string{
	"name":                "name",
	"description":         "description",
	"labels":              "labels",
	"backup_window_start": "config_spec.backup_window_start",
	"resources":           "config_spec.resources",
}

func updateMysqlClusterParams(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	req, err := getMysqlClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	resources := expandMysqlResources(d)
	backupWindowStart := expandMysqlBackupWindowStart(d)
	req.ConfigSpec = &mysql.ConfigSpec{
		Resources:         resources,
		Version:           d.Get("version").(string),
		BackupWindowStart: backupWindowStart,
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbMysqlUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
			onDone = append(onDone, func() {
				d.SetPartial(field)
			})
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	req.UpdateMask = &field_mask.FieldMask{Paths: updatePath}
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MySQL().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating MySQL Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}
	return nil
}

func getMysqlClusterUpdateRequest(d *schema.ResourceData) (*mysql.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating MySQL cluster: %s", err)
	}

	req := &mysql.UpdateClusterRequest{
		ClusterId:   d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
	}

	return req, nil
}

func updateMysqlClusterDatabases(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currDBs, err := listMysqlDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandMysqlDatabases(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mysqlDatabasesDiff(currDBs, targetDBs)

	for _, db := range toDelete {
		err := deleteMysqlDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createMysqlDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	d.SetPartial("database")
	return nil
}

// Takes the current list of dbs and the desirable list of dbs.
// Returns the slice of dbs to delete and the slice of dbs to add.
func mysqlDatabasesDiff(currDBs []*mysql.Database, targetDBs []*mysql.DatabaseSpec) ([]string, []string) {
	m := map[string]bool{}
	toAdd := []string{}
	toDelete := map[string]bool{}
	for _, db := range currDBs {
		toDelete[db.Name] = true
		m[db.Name] = true
	}

	for _, db := range targetDBs {
		delete(toDelete, db.Name)
		if _, ok := m[db.Name]; !ok {
			toAdd = append(toAdd, db.Name)
		}
	}

	toDel := []string{}
	for u := range toDelete {
		toDel = append(toDel, u)
	}

	return toDel, toAdd
}

func deleteMysqlDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Database().Delete(ctx, &mysql.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMysqlDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Database().Create(ctx, &mysql.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &mysql.DatabaseSpec{
				Name: dbName,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func updateMysqlClusterUsers(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()
	currUsers, err := listMysqlUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	targetUsers, err := expandMysqlUserSpecs(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := mysqlUsersDiff(currUsers, targetUsers)
	for _, u := range toDelete {
		err := deleteMysqlUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}
	for _, u := range toAdd {
		err := createMysqlUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")
	changedUsers, err := mysqlChangedUsers(oldSpecs.(*schema.Set), newSpecs.(*schema.Set))
	if err != nil {
		return err
	}
	for _, u := range changedUsers {
		err := updateMysqlUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	d.SetPartial("user")
	return nil
}

// Takes the current list of users and the desirable list of users.
// Returns the slice of usernames to delete and the slice of users to add.
func mysqlUsersDiff(currUsers []*mysql.User, targetUsers []*mysql.UserSpec) ([]string, []*mysql.UserSpec) {
	m := map[string]bool{}
	toDelete := map[string]bool{}
	toAdd := []*mysql.UserSpec{}

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

func deleteMysqlUser(ctx context.Context, config *Config, d *schema.ResourceData, userName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().User().Delete(ctx, &mysql.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  userName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func createMysqlUser(ctx context.Context, config *Config, d *schema.ResourceData, user *mysql.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().User().Create(ctx, &mysql.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

// Takes the old set of user specs and the new set of user specs.
// Returns the slice of user specs which have changed.
func mysqlChangedUsers(oldSpecs *schema.Set, newSpecs *schema.Set) ([]*mysql.UserSpec, error) {
	result := []*mysql.UserSpec{}
	m := map[string]*mysql.UserSpec{}
	for _, spec := range oldSpecs.List() {
		user, err := expandMysqlUser(spec.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		m[user.Name] = user
	}
	for _, spec := range newSpecs.List() {
		user, err := expandMysqlUser(spec.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		if u, ok := m[user.Name]; ok {
			if user.Password != u.Password || fmt.Sprintf("%v", user.Permissions) != fmt.Sprintf("%v", u.Permissions) {
				result = append(result, user)
			}
		}
	}
	return result, nil
}

func updateMysqlUser(ctx context.Context, config *Config, d *schema.ResourceData, user *mysql.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().User().Update(ctx, &mysql.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    user.Name,
			Password:    user.Password,
			Permissions: user.Permissions,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in MySQL Cluster %q: %s", d.Id(), err)
	}
	return nil
}

func validateMysqlAssignPublicIP(currentHosts []*mysql.Host, targetHosts []*MySQLHostSpec) error {
	for _, currentHost := range currentHosts {
		for _, targetHost := range targetHosts {
			if currentHost.Name == targetHost.Fqdn &&
				(currentHost.AssignPublicIp != targetHost.HostSpec.AssignPublicIp) {
				return fmt.Errorf("forbidden to change assign_public_ip setting for existing host %s in resource_yandex_mdb_mysql_cluster, "+
					"if you really need it you should delete one host and add another", currentHost.Name)
			}
		}
	}
	return nil
}

func updateMysqlClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currHosts, err := listMysqlHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetHosts, err := expandMysqlHosts(d)
	if err != nil {
		return err
	}

	err = validateMysqlAssignPublicIP(currHosts, targetHosts)
	if err != nil {
		return err
	}

	toDelete, toAdd := mysqlHostsDiff(currHosts, targetHosts)

	for _, h := range toAdd {
		err := addMysqlHost(ctx, config, d, h)
		if err != nil {
			return err
		}
	}

	if err := deleteMysqlHosts(ctx, config, d, toDelete); err != nil {
		return err
	}

	d.SetPartial("host")
	return nil
}

func mysqlHostsDiff(currHosts []*mysql.Host, targetHosts []*MySQLHostSpec) ([]string, []*mysql.HostSpec) {
	m := map[string]*MySQLHostSpec{}

	toAdd := []*mysql.HostSpec{}
	for _, h := range targetHosts {
		if !h.HasComputedFqdn {
			toAdd = append(toAdd, h.HostSpec)
		} else {
			m[h.Fqdn] = h
		}
	}

	toDelete := []string{}
	for _, h := range currHosts {
		_, ok := m[h.Name]
		if !ok {
			toDelete = append(toDelete, h.Name)
		}
	}

	return toDelete, toAdd
}

func addMysqlHost(ctx context.Context, config *Config, d *schema.ResourceData, host *mysql.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Cluster().AddHosts(ctx, &mysql.AddClusterHostsRequest{
			ClusterId: d.Id(),
			HostSpecs: []*mysql.HostSpec{host},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create host for MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating host for MySQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating host for MySQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func deleteMysqlHosts(ctx context.Context, config *Config, d *schema.ResourceData, names []string) error {
	if len(names) == 0 {
		return nil
	}
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().MySQL().Cluster().DeleteHosts(ctx, &mysql.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: names,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from MySQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from MySQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting host from MySQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func resourceYandexMDBMySQLClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting MySQL Cluster %q", d.Id())

	req := &mysql.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().MySQL().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("MySQL Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting MySQL Cluster %q", d.Id())
	return nil
}
