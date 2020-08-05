package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
)

const (
	yandexMDBPostgreSQLClusterCreateTimeout = 30 * time.Minute
	yandexMDBPostgreSQLClusterDeleteTimeout = 15 * time.Minute
	yandexMDBPostgreSQLClusterUpdateTimeout = 60 * time.Minute
)

func resourceYandexMDBPostgreSQLCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBPostgreSQLClusterCreate,
		Read:   resourceYandexMDBPostgreSQLClusterRead,
		Update: resourceYandexMDBPostgreSQLClusterUpdate,
		Delete: resourceYandexMDBPostgreSQLClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBPostgreSQLClusterCreateTimeout),
			Update: schema.DefaultTimeout(yandexMDBPostgreSQLClusterUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBPostgreSQLClusterDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"environment": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
									"disk_size": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"disk_type_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"autofailover": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"pooler_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"pooling_mode": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"pool_discard": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
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
									"data_lens": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
					},
				},
			},
			"database": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"owner": {
							Type:     schema.TypeString,
							Required: true,
						},
						"lc_collate": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "C",
						},
						"lc_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "C",
						},
						"extension": {
							Type:     schema.TypeSet,
							Set:      pgExtensionHash,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"version": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"user": {
				Type:     schema.TypeList,
				Required: true,
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
						"login": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"grants": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"permission": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Set:      pgUserPermissionHash,
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
						"subnet_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"assign_public_ip": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"fqdn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"role": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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

func resourceYandexMDBPostgreSQLClusterRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	createdAt, err := getTimestamp(cluster.CreatedAt)
	if err != nil {
		return err
	}

	d.Set("created_at", createdAt)
	d.Set("health", cluster.GetHealth().String())
	d.Set("status", cluster.GetStatus().String())
	d.Set("folder_id", cluster.GetFolderId())
	d.Set("name", cluster.GetName())
	d.Set("description", cluster.GetDescription())
	d.Set("environment", cluster.GetEnvironment().String())
	d.Set("network_id", cluster.GetNetworkId())
	if err := d.Set("labels", cluster.GetLabels()); err != nil {
		return err
	}

	pgClusterConf, err := flattenPGClusterConfig(cluster.Config)
	if err != nil {
		return err
	}
	if err := d.Set("config", pgClusterConf); err != nil {
		return err
	}

	userSpecs, err := expandPGUserSpecs(d)
	if err != nil {
		return err
	}
	passwords := pgUsersPasswords(userSpecs)
	users, err := listPGUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}
	sortPGUsers(users, userSpecs)

	fUsers, err := flattenPGUsers(users, passwords)
	if err != nil {
		return err
	}
	if err := d.Set("user", fUsers); err != nil {
		return err
	}

	hosts, err := listPGHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	hostSpecs, err := expandPGHosts(d)
	if err != nil {
		return err
	}

	sortPGHosts(hosts, hostSpecs)

	fHosts, err := flattenPGHosts(hosts)
	if err != nil {
		return err
	}

	if err := d.Set("host", fHosts); err != nil {
		return err
	}

	databases, err := listPGDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	databaseSpecs, err := expandPGDatabaseSpecs(d)
	if err != nil {
		return err
	}
	sortPGDatabases(databases, databaseSpecs)

	return d.Set("database", flattenPGDatabases(databases))
}

func sortPGUsers(users []*postgresql.User, specs []*postgresql.UserSpec) {
	for i, spec := range specs {
		for j := i + 1; j < len(users); j++ {
			if spec.Name == users[j].Name {
				users[i], users[j] = users[j], users[i]
				break
			}
		}
	}
}

func sortPGDatabases(databases []*postgresql.Database, specs []*postgresql.DatabaseSpec) {
	for i, spec := range specs {
		for j := i + 1; j < len(databases); j++ {
			if spec.Name == databases[j].Name {
				databases[i], databases[j] = databases[j], databases[i]
				break
			}
		}
	}
}

func resourceYandexMDBPostgreSQLClusterCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req, err := prepareCreatePostgreSQLRequest(d, config)

	if err != nil {
		return err
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().PostgreSQL().Cluster().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create PostgreSQL Cluster: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get PostgreSQL Cluster create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*postgresql.CreateClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get PostgreSQL Cluster ID from create operation metadata")
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create PostgreSQL Cluster: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("PostgreSQL Cluster creation failed: %s", err)
	}

	return resourceYandexMDBPostgreSQLClusterRead(d, meta)
}

func prepareCreatePostgreSQLRequest(d *schema.ResourceData, meta *Config) (*postgresql.CreateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error while expanding labels on PostgreSQL Cluster create: %s", err)
	}

	folderID, err := getFolderID(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error getting folder ID while creating PostgreSQL Cluster: %s", err)
	}

	hostsFromScheme, err := expandPGHosts(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding host specs on PostgreSQL Cluster create: %s", err)
	}

	e := d.Get("environment").(string)
	env, err := parsePostgreSQLEnv(e)
	if err != nil {
		return nil, fmt.Errorf("Error resolving environment while creating PostgreSQL Cluster: %s", err)
	}

	confSpec, err := expandPGConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding cluster config on PostgreSQL Cluster create: %s", err)
	}

	userSpecs, err := expandPGUserSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding user specs on PostgreSQL Cluster create: %s", err)
	}

	databaseSpecs, err := expandPGDatabaseSpecs(d)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding database specs on PostgreSQL Cluster create: %s", err)
	}
	hostSpecs := make([]*postgresql.HostSpec, 0)
	for _, host := range hostsFromScheme {
		hostSpecs = append(hostSpecs, host.HostSpec)
	}
	req := &postgresql.CreateClusterRequest{
		FolderId:      folderID,
		Name:          d.Get("name").(string),
		Description:   d.Get("description").(string),
		NetworkId:     d.Get("network_id").(string),
		Labels:        labels,
		Environment:   env,
		ConfigSpec:    confSpec,
		UserSpecs:     userSpecs,
		DatabaseSpecs: databaseSpecs,
		HostSpecs:     hostSpecs,
	}

	return req, nil
}

func resourceYandexMDBPostgreSQLClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	d.Partial(true)

	if err := updatePGClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("user") {
		if err := updatePGClusterUsersWithoutPermissions(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("database") {
		if err := updatePGClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updatePGClusterUserPermissions(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updatePGClusterHosts(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)
	return resourceYandexMDBPostgreSQLClusterRead(d, meta)
}

func updatePGClusterParams(d *schema.ResourceData, meta interface{}) error {
	req, err := getPGClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	mdbPGUpdateFieldsMap := map[string]string{
		"name":                         "name",
		"description":                  "description",
		"labels":                       "labels",
		"config.0.version":             "config_spec.version",
		"config.0.autofailover":        "config_spec.autofailover",
		"config.0.pooler_config":       "config_spec.pooler_config",
		"config.0.access":              "config_spec.access",
		"config.0.backup_window_start": "config_spec.backup_window_start",
		"config.0.resources":           "config_spec.resources",
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbPGUpdateFieldsMap {
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

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().PostgreSQL().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to update PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}

	return nil
}

func getPGClusterUpdateRequest(d *schema.ResourceData) (*postgresql.UpdateClusterRequest, error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, fmt.Errorf("error expanding labels while updating PostgreSQL Cluster: %s", err)
	}

	configSpec, err := expandPGConfigSpec(d)
	if err != nil {
		return nil, fmt.Errorf("error expanding config while updating PostgreSQL Cluster: %s", err)
	}

	req := &postgresql.UpdateClusterRequest{
		ClusterId:   d.Id(),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Labels:      labels,
		ConfigSpec:  configSpec,
	}

	return req, nil
}

func updatePGClusterDatabases(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currDBs, err := listPGDatabases(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetDBs, err := expandPGDatabaseSpecs(d)
	if err != nil {
		return err
	}

	err = validateNoUpdatingCollation(currDBs, targetDBs)
	if err != nil {
		return err
	}

	toDelete, toAdd := pgDatabasesDiff(currDBs, targetDBs)

	for _, dbn := range toDelete {
		err := deletePGDatabase(ctx, config, d, dbn)
		if err != nil {
			return err
		}
	}
	for _, db := range toAdd {
		err := createPGDatabase(ctx, config, d, db)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("database")

	changedDatabases, err := pgChangedDatabases(oldSpecs.([]interface{}), newSpecs.([]interface{}))
	if err != nil {
		return err
	}

	for _, u := range changedDatabases {
		err := updatePGDatabase(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	d.SetPartial("database")
	return nil
}

func validateNoUpdatingCollation(currentDatabases []*postgresql.Database, targetDatabases []*postgresql.DatabaseSpec) error {
	for _, currentDatabase := range currentDatabases {
		for _, targetDatabase := range targetDatabases {
			if currentDatabase.Name == targetDatabase.Name &&
				(currentDatabase.LcCollate != targetDatabase.LcCollate || currentDatabase.LcCtype != targetDatabase.LcCtype) {
				return fmt.Errorf("impossible to change lc_collate or lc_type for PostgreSQL Cluster database %s", currentDatabase.Name)
			}
		}
	}
	return nil
}

func updatePGClusterUsersWithoutPermissions(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currUsers, err := listPGUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetUsers, err := expandPGUserSpecs(d)
	if err != nil {
		return err
	}

	toDelete, toAdd := pgUsersDiff(currUsers, targetUsers)
	for _, u := range toDelete {
		err := deletePGUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}
	for _, u := range toAdd {
		u.Permissions = make([]*postgresql.Permission, 0)
		err := createPGUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	oldSpecs, newSpecs := d.GetChange("user")

	changedUsers, err := pgChangedUsers(oldSpecs.([]interface{}), newSpecs.([]interface{}), false)
	if err != nil {
		return err
	}
	for _, user := range changedUsers {
		err := updatePGUser(ctx, config, d, user)
		if err != nil {
			return err
		}
	}

	d.SetPartial("user")

	return nil
}

func updatePGClusterUserPermissions(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	oldSpecs, newSpecs := d.GetChange("user")
	changedUsers, err := pgChangedUsers(oldSpecs.([]interface{}), newSpecs.([]interface{}), true)
	if err != nil {
		return err
	}
	for _, u := range changedUsers {
		err := updatePGUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	d.SetPartial("user")
	return nil
}

func validatePGAssignPublicIP(currentHosts []*postgresql.Host, targetHosts []*PostgreSQLHostSpec) error {
	for _, currentHost := range currentHosts {
		for _, targetHost := range targetHosts {
			if currentHost.Name == targetHost.Fqdn &&
				(currentHost.AssignPublicIp != targetHost.HostSpec.AssignPublicIp) {
				return fmt.Errorf("forbidden to change assign_public_ip setting for existing host %s in resource_yandex_mdb_postgresql_cluster, "+
					"if you really need it you should delete one host and add another", currentHost.Name)
			}
		}
	}
	return nil
}

func updatePGClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currHosts, err := listPGHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	targetHosts, err := expandPGHosts(d)
	if err != nil {
		return err
	}

	err = validatePGAssignPublicIP(currHosts, targetHosts)
	if err != nil {
		return err
	}

	toDelete, toAdd := pgHostsDiff(currHosts, targetHosts)

	for _, h := range toAdd {
		err := addPGHost(ctx, config, d, h)
		if err != nil {
			return err
		}
	}

	if len(toDelete) != 0 {
		if err := deletePGHosts(ctx, config, d, toDelete); err != nil {
			return err
		}
	}

	d.SetPartial("host")
	return nil
}

func resourceYandexMDBPostgreSQLClusterDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	log.Printf("[DEBUG] Deleting PostgreSQL Cluster %q", d.Id())

	req := &postgresql.DeleteClusterRequest{
		ClusterId: d.Id(),
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().PostgreSQL().Cluster().Delete(ctx, req))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("PostgreSQL Cluster %q", d.Id()))
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Finished deleting PostgreSQL Cluster %q", d.Id())

	return nil
}

func createPGUser(ctx context.Context, config *Config, d *schema.ResourceData, user *postgresql.UserSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().User().Create(ctx, &postgresql.CreateUserRequest{
			ClusterId: d.Id(),
			UserSpec:  user,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create user for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating user for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating user for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func updatePGUser(ctx context.Context, config *Config, d *schema.ResourceData, user IndexedUserSpec) error {
	mdbPGUserUpdateFieldsMap := map[string]string{
		"user.%d.password":   "password",
		"user.%d.permission": "permissions",
		"user.%d.login":      "login",
		"user.%d.grants":     "grants",
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbPGUserUpdateFieldsMap {
		if d.HasChange(fmt.Sprintf(field, user.index)) {
			updatePath = append(updatePath, path)
			onDone = append(onDone, func() {
				d.SetPartial(field)
			})
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().User().Update(ctx, &postgresql.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    user.user.Name,
			Password:    user.user.Password,
			Permissions: user.user.Permissions,
			ConnLimit:   user.user.ConnLimit.GetValue(),
			Login:       user.user.Login,
			Grants:      user.user.Grants,
			UpdateMask:  &field_mask.FieldMask{Paths: updatePath},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update user in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating user in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	for _, f := range onDone {
		f()
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating user for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func deletePGUser(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().User().Delete(ctx, &postgresql.DeleteUserRequest{
			ClusterId: d.Id(),
			UserName:  name,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting user from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting user from PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func listPGUsers(ctx context.Context, config *Config, id string) ([]*postgresql.User, error) {
	users := []*postgresql.User{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().PostgreSQL().User().List(ctx, &postgresql.ListUsersRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of users for PostgreSQL Cluster '%q': %s", id, err)
		}

		users = append(users, resp.Users...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return users, nil
}

func createPGDatabase(ctx context.Context, config *Config, d *schema.ResourceData, db *postgresql.DatabaseSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Database().Create(ctx, &postgresql.CreateDatabaseRequest{
			ClusterId: d.Id(),
			DatabaseSpec: &postgresql.DatabaseSpec{
				Name:       db.Name,
				Owner:      db.Owner,
				LcCollate:  db.LcCollate,
				LcCtype:    db.LcCtype,
				Extensions: db.Extensions,
			},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create database in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while adding database to PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating database for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func updatePGDatabase(ctx context.Context, config *Config, d *schema.ResourceData, db *postgresql.DatabaseSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Database().Update(ctx, &postgresql.UpdateDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: db.Name,
			Extensions:   db.Extensions,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update database in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating database in PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating database for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func deletePGDatabase(ctx context.Context, config *Config, d *schema.ResourceData, dbName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Database().Delete(ctx, &postgresql.DeleteDatabaseRequest{
			ClusterId:    d.Id(),
			DatabaseName: dbName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting database from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting database from PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func listPGDatabases(ctx context.Context, config *Config, id string) ([]*postgresql.Database, error) {
	databases := []*postgresql.Database{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().PostgreSQL().Database().List(ctx, &postgresql.ListDatabasesRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of databases for PostgreSQL Cluster '%q': %s", id, err)
		}

		databases = append(databases, resp.Databases...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return databases, nil
}

func addPGHost(ctx context.Context, config *Config, d *schema.ResourceData, host *postgresql.HostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Cluster().AddHosts(ctx, &postgresql.AddClusterHostsRequest{
			ClusterId: d.Id(),
			HostSpecs: []*postgresql.HostSpec{host},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to create host for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while creating host for PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating host for PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func deletePGHosts(ctx context.Context, config *Config, d *schema.ResourceData, names []string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Cluster().DeleteHosts(ctx, &postgresql.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: names,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to delete host from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while deleting host from PostgreSQL Cluster %q: %s", d.Id(), err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting host from PostgreSQL Cluster %q failed: %s", d.Id(), err)
	}

	return nil
}

func listPGHosts(ctx context.Context, config *Config, id string) ([]*postgresql.Host, error) {
	hosts := []*postgresql.Host{}
	pageToken := ""

	for {
		resp, err := config.sdk.MDB().PostgreSQL().Cluster().ListHosts(ctx, &postgresql.ListClusterHostsRequest{
			ClusterId: id,
			PageSize:  defaultMDBPageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("Error while getting list of hosts for PostgreSQL Cluster '%q': %s", id, err)
		}

		hosts = append(hosts, resp.Hosts...)

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return hosts, nil
}
