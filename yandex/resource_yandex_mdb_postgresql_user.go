package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const (
	yandexMDBPostgreSQLUserCreateTimeout = 10 * time.Minute
	yandexMDBPostgreSQLUserReadTimeout   = 1 * time.Minute
	yandexMDBPostgreSQLUserUpdateTimeout = 10 * time.Minute
	yandexMDBPostgreSQLUserDeleteTimeout = 10 * time.Minute
)

func resourceYandexMDBPostgreSQLUser() *schema.Resource {
	return &schema.Resource{
		Description: "Manages a PostgreSQL user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-postgresql/).",

		Create: resourceYandexMDBPostgreSQLUserCreate,
		Read:   resourceYandexMDBPostgreSQLUserRead,
		Update: resourceYandexMDBPostgreSQLUserUpdate,
		Delete: resourceYandexMDBPostgreSQLUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBPostgreSQLUserCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBPostgreSQLUserReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBPostgreSQLUserUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBPostgreSQLUserDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:        schema.TypeString,
				Description: "The ID of the PostgreSQL cluster.",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the user.",
				Required:    true,
			},
			"password": {
				Type:        schema.TypeString,
				Description: "The password of the user.",
				Optional:    true,
				Sensitive:   true,
			},
			"login": {
				Type:        schema.TypeBool,
				Description: "User's ability to login.",
				Optional:    true,
				Default:     true,
			},
			"grants": {
				Type:        schema.TypeList,
				Description: "List of the user's grants.",
				Optional:    true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
			// TODO change to permissions
			"permission": {
				Type:        schema.TypeSet,
				Description: "Set of permissions granted to the user.",
				Optional:    true,
				Set:         pgUserPermissionHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:        schema.TypeString,
							Description: "The name of the database that the permission grants access to.",
							Required:    true,
						},
					},
				},
			},
			"conn_limit": {
				Type:        schema.TypeInt,
				Description: "The maximum number of connections per user. (Default 50).",
				Optional:    true,
				Computed:    true,
			},
			"settings": {
				Type:         schema.TypeMap,
				Description:  "Map of user settings. [Full description](https://yandex.cloud/docs/managed-postgresql/api-ref/grpc/Cluster/create#yandex.cloud.mdb.postgresql.v1.UserSettings).\n\n* `default_transaction_isolation` - defines the default isolation level to be set for all new SQL transactions. One of:  - 0: `unspecified`\n  - 1: `read uncommitted`\n  - 2: `read committed`\n  - 3: `repeatable read`\n  - 4: `serializable`\n\n* `lock_timeout` - The maximum time (in milliseconds) for any statement to wait for acquiring a lock on an table, index, row or other database object (default 0)\n\n* `log_min_duration_statement` - This setting controls logging of the duration of statements. (default -1 disables logging of the duration of statements.)\n\n* `synchronous_commit` - This setting defines whether DBMS will commit transaction in a synchronous way. One of:\n  - 0: `unspecified`\n  - 1: `on`\n  - 2: `off`\n  - 3: `local`\n  - 4: `remote write`\n  - 5: `remote apply`\n\n* `temp_file_limit` - The maximum storage space size (in kilobytes) that a single process can use to create temporary files.\n\n* `log_statement` - This setting specifies which SQL statements should be logged (on the user level). One of:\n  - 0: `unspecified`\n  - 1: `none`\n  - 2: `ddl`\n  - 3: `mod`\n  - 4: `all`\n\n* `pool_mode` - Mode that the connection pooler is working in with specified user. One of:\n  - 1: `session`\n  - 2: `transaction`\n  - 3: `statement`\n\n* `prepared_statements_pooling` - This setting allows user to use prepared statements with transaction pooling. Boolean.\n\n* `catchup_timeout` - The connection pooler setting. It determines the maximum allowed replication lag (in seconds). Pooler will reject connections to the replica with a lag above this threshold. Default value is 0, which disables this feature. Integer.\n\n* `wal_sender_timeout` - The maximum time (in milliseconds) to wait for WAL replication (can be set only for PostgreSQL 12+). Terminate replication connections that are inactive for longer than this amount of time. Integer.\n\n* `idle_in_transaction_session_timeout` - Sets the maximum allowed idle time (in milliseconds) between queries, when in a transaction. Value of 0 (default) disables the timeout. Integer.\n\n* `statement_timeout` - The maximum time (in milliseconds) to wait for statement. Value of 0 (default) disables the timeout. Integer\n\n",
				Optional:     true,
				ValidateFunc: generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"deletion_protection": {
				Type:         schema.TypeString,
				Description:  common.ResourceDescriptions["deletion_protection"],
				Optional:     true,
				Default:      "unspecified",
				ValidateFunc: validation.StringInSlice([]string{"true", "false", "unspecified"}, false),
			},
			"connection_manager": {
				Type:        schema.TypeMap,
				Description: "Connection Manager connection configuration. Filled in by the server automatically.",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"generate_password": {
				Type:        schema.TypeBool,
				Description: "Generate password using Connection Manager. Allowed values: true or false. It's used only during user creation and is ignored during updating.\n\n~> **Must specify either password or generate_password**.\n",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceYandexMDBPostgreSQLUserCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	userSpec, err := expandPgUserSpec(d)
	if err != nil {
		return err
	}

	if !isValidPGPasswordConfiguration(userSpec) {
		return fmt.Errorf("must specify either password or generate_password")
	}

	request := &postgresql.CreateUserRequest{
		ClusterId: clusterID,
		UserSpec:  userSpec,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL user create request: %+v", request)
		return config.sdk.MDB().PostgreSQL().User().Create(ctx, request)
	})

	userID := constructResourceId(clusterID, userSpec.Name)
	d.SetId(userID)

	if err != nil {
		return fmt.Errorf("error while requesting API to create user for PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while creating user for PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating user for PostgreSQL Cluster %q failed: %s", clusterID, err)
	}

	return resourceYandexMDBPostgreSQLUserRead(d, meta)
}

func expandPgUserSpec(d *schema.ResourceData) (*postgresql.UserSpec, error) {
	user := &postgresql.UserSpec{}
	if v, ok := d.GetOkExists("name"); ok {
		user.Name = v.(string)
	}

	if v, ok := d.GetOkExists("password"); ok {
		user.Password = v.(string)
	}

	if v, ok := d.GetOkExists("login"); ok {
		user.Login = &wrappers.BoolValue{Value: v.(bool)}
	}

	if v, ok := d.GetOkExists("conn_limit"); ok {
		user.ConnLimit = &wrappers.Int64Value{Value: int64(v.(int))}
	}

	if v, ok := d.GetOkExists("permission"); ok {
		permissions, err := expandPGUserPermissions(v.(*schema.Set))
		if err != nil {
			return nil, err
		}
		user.Permissions = permissions
	}

	if v, ok := d.GetOkExists("grants"); ok {
		gs, err := expandPGUserGrants(v.([]interface{}))
		if err != nil {
			return nil, err
		}
		user.Grants = gs
	}

	if _, ok := d.GetOkExists("settings"); ok {
		if user.Settings == nil {
			user.Settings = &postgresql.UserSettings{}
		}

		err := expandResourceGenerate(mdbPGUserSettingsFieldsInfo, d, user.Settings, "settings.", true)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := d.GetOk("deletion_protection"); ok {
		user.DeletionProtection = mdbPGTristateBooleanName[v.(string)]
	}

	if v, ok := d.GetOk("generate_password"); ok {
		user.GeneratePassword = wrapperspb.Bool(v.(bool))
	}

	return user, nil
}

func resourceYandexMDBPostgreSQLUserRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	clusterID, username, err := deconstructResourceId(d.Id())
	if err != nil {
		return err
	}

	apiUser, err := config.sdk.MDB().PostgreSQL().User().Get(ctx, &postgresql.GetUserRequest{
		ClusterId: clusterID,
		UserName:  username,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("User %q", username))
	}

	stateUser, err := expandPgUserSpec(d)
	if err != nil {
		return err
	}

	userPermissions, err := removePgUserOwnerPermissions(meta, clusterID, apiUser.Name, apiUser.Permissions, stateUser.Permissions)
	if err != nil {
		return fmt.Errorf("error while removing owner permissions from user in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	d.Set("cluster_id", clusterID)
	d.Set("name", apiUser.Name)
	d.Set("login", apiUser.Login.GetValue())
	d.Set("grants", apiUser.Grants)
	d.Set("permission", flattenPGUserPermissions(userPermissions))
	d.Set("conn_limit", apiUser.ConnLimit)
	knownDefault := map[string]struct{}{
		"log_min_duration_statement": {},
	}
	settings, err := flattenResourceGenerateMapS(apiUser.Settings, false, mdbPGUserSettingsFieldsInfo, false, true, knownDefault)
	if err != nil {
		return err
	}

	d.Set("settings", settings)
	d.Set("deletion_protection", mdbPGResolveTristateBoolean(apiUser.DeletionProtection))
	d.Set("connection_manager", flattenPGUserConnectionManager(apiUser.ConnectionManager))

	return nil
}

func resourceYandexMDBPostgreSQLUserUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	user, err := expandPgUserSpec(d)
	if err != nil {
		return err
	}

	if !isValidPGPasswordConfiguration(user) {
		return fmt.Errorf("must specify either password or generate_password")
	}

	updatePath := []string{}
	changeMask := map[string]string{
		"password":                                     "password",
		"permission":                                   "permissions",
		"login":                                        "login",
		"grants":                                       "grants",
		"conn_limit":                                   "conn_limit",
		"settings.default_transaction_isolation":       "settings.default_transaction_isolation",
		"settings.lock_timeout":                        "settings.lock_timeout",
		"settings.log_min_duration_statement":          "settings.log_min_duration_statement",
		"settings.synchronous_commit":                  "settings.synchronous_commit",
		"settings.temp_file_limit":                     "settings.temp_file_limit",
		"settings.log_statement":                       "settings.log_statement",
		"settings.pool_mode":                           "settings.pool_mode",
		"settings.prepared_statements_pooling":         "settings.prepared_statements_pooling",
		"settings.catchup_timeout":                     "settings.catchup_timeout",
		"settings.wal_sender_timeout":                  "settings.wal_sender_timeout",
		"settings.idle_in_transaction_session_timeout": "settings.idle_in_transaction_session_timeout",
		"settings.statement_timeout":                   "settings.statement_timeout",
		"settings.pgaudit":                             "settings.pgaudit",
	}

	for field, mask := range changeMask {
		if d.HasChange(field) {
			updatePath = append(updatePath, mask)
		}
	}

	if user.DeletionProtection != nil {
		updatePath = append(updatePath, "deletion_protection")
	}

	if len(updatePath) == 0 && user.DeletionProtection == nil {
		updatePath = []string{"name"}
	}

	if len(updatePath) == 0 {
		return nil
	}

	clusterID := d.Get("cluster_id").(string)
	userPermissions, err := addPgUserOwnerPermissions(meta, clusterID, user.Name, user.Permissions)
	if err != nil {
		return fmt.Errorf("error while adding owner permissions to user in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	request := &postgresql.UpdateUserRequest{
		ClusterId:          clusterID,
		UserName:           user.Name,
		Password:           user.Password,
		Permissions:        userPermissions,
		ConnLimit:          user.ConnLimit.GetValue(),
		Login:              user.Login,
		Grants:             user.Grants,
		Settings:           user.Settings,
		DeletionProtection: user.DeletionProtection,
		UpdateMask:         &fieldmaskpb.FieldMask{Paths: updatePath},
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL user update request: %+v", request)
		return config.sdk.MDB().PostgreSQL().User().Update(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to update user in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while updating user in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating user for PostgreSQL Cluster %q failed: %s", clusterID, err)
	}
	return resourceYandexMDBPostgreSQLUserRead(d, meta)
}

func resourceYandexMDBPostgreSQLUserDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	username := d.Get("name").(string)

	request := &postgresql.DeleteUserRequest{
		ClusterId: clusterID,
		UserName:  username,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL user delete request: %+v", request)
		return config.sdk.MDB().PostgreSQL().User().Delete(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to delete user from PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err = op.Wait(ctx); err != nil {
		return fmt.Errorf("error while deleting user from PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting user from PostgreSQL Cluster %q failed: %s", clusterID, err)
	}

	return nil
}

// If the user is the database owner, it is assumed that they have permissions on that database (and the API reflects this as well).
// Therefore, it is redundant to declare explicit permissions for a user on a database they already own.
// However, if the user specifies both fields - owner and permissions - we simply ignore those permissions in the Read
// function to avoid constant changes in the Terraform plan.
func removePgUserOwnerPermissions(
	meta interface{},
	clusterID string,
	username string,
	apiUserPermissions []*postgresql.Permission,
	stateUserPermissions []*postgresql.Permission,
) ([]*postgresql.Permission, error) {
	config := meta.(*Config)
	responses := make([]*postgresql.ListDatabasesResponse, 0)

	nextPageToken := ""
	for {
		req := &postgresql.ListDatabasesRequest{
			ClusterId: clusterID,
			PageSize:  100,
		}
		if nextPageToken != "" {
			req.SetPageToken(nextPageToken)
		}
		resp, _ := config.sdk.MDB().PostgreSQL().Database().List(context.Background(), req)
		responses = append(responses, resp)

		if resp.GetNextPageToken() == "" {
			break
		}
		nextPageToken = resp.GetNextPageToken()
	}

	dbMap := make(map[string]*postgresql.Database)
	for _, resp := range responses {
		for _, db := range resp.Databases {
			dbMap[db.Name] = db
		}
	}

	statePermissionsMap := make(map[string]*postgresql.Permission)
	for _, permission := range stateUserPermissions {
		statePermissionsMap[permission.DatabaseName] = permission
	}

	newPerms := []*postgresql.Permission{}
	for _, p := range apiUserPermissions {
		if db, ok := dbMap[p.DatabaseName]; ok && !isOwnerWithoutPermissions(db, statePermissionsMap, username) {
			newPerms = append(newPerms, p)
		}
	}

	return newPerms, nil
}

func isOwnerWithoutPermissions(db *postgresql.Database, statePermissionsMap map[string]*postgresql.Permission, username string) bool {
	return db.Owner == username && statePermissionsMap[db.Name] == nil
}

// Add permissions for databases where user is owner
func addPgUserOwnerPermissions(meta interface{}, clusterID string, name string, permissions []*postgresql.Permission) ([]*postgresql.Permission, error) {
	config := meta.(*Config)

	resp, err := config.sdk.MDB().PostgreSQL().Database().List(context.Background(), &postgresql.ListDatabasesRequest{
		ClusterId: clusterID,
	})
	if err != nil {
		return nil, err
	}

	for _, db := range resp.Databases {
		if db.Owner == name {
			permissions = append(permissions, &postgresql.Permission{
				DatabaseName: db.Name,
			})
		}
	}

	return permissions, nil
}
