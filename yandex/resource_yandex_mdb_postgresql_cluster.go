package yandex

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
						"backup_retain_period_days": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"performance_diagnostics": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"sessions_sampling_interval": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"statements_sampling_interval": {
										Type:     schema.TypeInt,
										Required: true,
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
									"web_sql": {
										Type:     schema.TypeBool,
										Optional: true,
										Computed: true,
									},
									"serverless": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
								},
							},
						},
						"postgresql_config": {
							Type:             schema.TypeMap,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGSettingsFieldsInfo),
							ValidateFunc:     generateMapSchemaValidateFunc(mdbPGSettingsFieldsInfo),
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
							Computed: true,
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
						"conn_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"settings": {
							Type:             schema.TypeMap,
							Optional:         true,
							Computed:         true,
							DiffSuppressFunc: generateMapSchemaDiffSuppressFunc(mdbPGUserSettingsFieldsInfo),
							ValidateFunc:     generateMapSchemaValidateFunc(mdbPGUserSettingsFieldsInfo),
							Elem: &schema.Schema{
								Type: schema.TypeString,
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
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"replication_source": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
							Computed: true,
						},
						"replication_source_name": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"folder_id": {
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
				Computed: true,
			},
			"host_master_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"restore": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backup_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"time_inclusive": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
						"time": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: stringToTimeValidateFunc,
						},
					},
				},
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
							ValidateFunc: pgMaintenanceWindowSchemaValidateFunc,
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

	d.Set("created_at", getTimestamp(cluster.CreatedAt))
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

	pgClusterConf, err := flattenPGClusterConfig(cluster.Config, d)
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

	fUsers, err := flattenPGUsers(users, passwords, mdbPGUserSettingsFieldsInfo)
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

	fHosts, hostMasterName, err := flattenPGHosts(d, hosts, false)
	if err != nil {
		return err
	}

	if err := d.Set("host", fHosts); err != nil {
		return err
	}
	if err := d.Set("host_master_name", hostMasterName); err != nil {
		return err
	}

	maintenanceWindow, err := flattenPGMaintenanceWindow(cluster.MaintenanceWindow)
	if err != nil {
		return err
	}

	if err := d.Set("maintenance_window", maintenanceWindow); err != nil {
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

	if err := d.Set("security_group_ids", cluster.SecurityGroupIds); err != nil {
		return err
	}

	d.Set("deletion_protection", cluster.DeletionProtection)

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

	if backupID, ok := d.GetOk("restore.0.backup_id"); ok && backupID != "" {
		return resourceYandexMDBPostgreSQLClusterRestore(d, meta, req, backupID.(string))
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

	if err := createPGClusterHosts(ctx, config, d); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts creation failed: %s", d.Id(), err)
	}

	if err := updateMasterPGClusterHosts(d, meta); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts set master failed: %s", d.Id(), err)
	}

	if err := updatePGClusterAfterCreate(d, meta); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v update params failed: %s", d.Id(), err)
	}

	return resourceYandexMDBPostgreSQLClusterRead(d, meta)
}

func resourceYandexMDBPostgreSQLClusterRestore(d *schema.ResourceData, meta interface{}, createClusterRequest *postgresql.CreateClusterRequest, backupID string) error {
	config := meta.(*Config)

	timeBackup := time.Now()
	timeInclusive := false

	if backupTime, ok := d.GetOk("restore.0.time"); ok {
		var err error
		timeBackup, err = parseStringToTime(backupTime.(string))
		if err != nil {
			return fmt.Errorf("Error while parsing restore.0.time to create PostgreSQL Cluster from backup %v, value: %v error: %s", backupID, backupTime, err)
		}
	}

	if timeInclusiveData, ok := d.GetOk("restore.0.time_inclusive"); ok {
		timeInclusive = timeInclusiveData.(bool)
	}

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().PostgreSQL().Cluster().Restore(ctx, &postgresql.RestoreClusterRequest{
		BackupId: backupID,
		Time: &timestamp.Timestamp{
			Seconds: timeBackup.Unix(),
		},
		TimeInclusive:    timeInclusive,
		Name:             createClusterRequest.Name,
		Description:      createClusterRequest.Description,
		Labels:           createClusterRequest.Labels,
		Environment:      createClusterRequest.Environment,
		ConfigSpec:       createClusterRequest.ConfigSpec,
		HostSpecs:        createClusterRequest.HostSpecs,
		NetworkId:        createClusterRequest.NetworkId,
		FolderId:         createClusterRequest.FolderId,
		SecurityGroupIds: createClusterRequest.SecurityGroupIds,
	}))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create PostgreSQL Cluster from backup %v: %s", backupID, err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get PostgreSQL Cluster create from backup %v operation metadata: %s", backupID, err)
	}

	md, ok := protoMetadata.(*postgresql.RestoreClusterMetadata)
	if !ok {
		return fmt.Errorf("Could not get PostgreSQL Cluster ID from create from backup %v operation metadata", backupID)
	}

	d.SetId(md.ClusterId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting for operation to create PostgreSQL Cluster from backup %v: %s", backupID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("PostgreSQL Cluster creation from backup %v failed: %s", backupID, err)
	}

	if err := createPGClusterHosts(ctx, config, d); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts creation from backup %v failed: %s", d.Id(), backupID, err)
	}

	if err := updateMasterPGClusterHosts(d, meta); err != nil {
		return fmt.Errorf("PostgreSQL Cluster %v hosts set master failed: %s", d.Id(), err)
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

	confSpec, _, err := expandPGConfigSpec(d)
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
		if host.HostSpec.ReplicationSource == "" {
			hostSpecs = append(hostSpecs, host.HostSpec)
		}
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	networkID, err := expandAndValidateNetworkId(d, meta)
	if err != nil {
		return nil, fmt.Errorf("Error while expanding network id on PostgreSQL Cluster create: %s", err)
	}

	req := &postgresql.CreateClusterRequest{
		FolderId:           folderID,
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		NetworkId:          networkID,
		Labels:             labels,
		Environment:        env,
		ConfigSpec:         confSpec,
		UserSpecs:          userSpecs,
		DatabaseSpecs:      databaseSpecs,
		HostSpecs:          hostSpecs,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	return req, nil
}

func updatePGClusterAfterCreate(d *schema.ResourceData, meta interface{}) error {

	maintenanceWindow, err := expandPGMaintenanceWindow(d)
	if err != nil {
		return fmt.Errorf("error expanding maintenance_window while updating PostgreSQL after creation: %s", err)
	}

	if maintenanceWindow == nil {
		return nil
	}
	updatePath := []string{"maintenance_window"}
	req := &postgresql.UpdateClusterRequest{
		ClusterId:         d.Id(),
		MaintenanceWindow: maintenanceWindow,
		UpdateMask:        &field_mask.FieldMask{Paths: updatePath},
	}

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.MDB().PostgreSQL().Cluster().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("error while requesting API to update PostgreSQL Cluster after creation %q: %s", d.Id(), err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while waiting for operation to update PostgreSQL Cluster after creation %q: %s", d.Id(), err)
	}

	return nil
}

func resourceYandexMDBPostgreSQLClusterUpdate(d *schema.ResourceData, meta interface{}) error {

	d.Partial(true)

	if err := setPGFolderID(d, meta); err != nil {
		return err
	}

	if err := updatePGClusterParams(d, meta); err != nil {
		return err
	}

	if d.HasChange("user") {
		if err := updatePGClusterUsersAdd(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("database") {
		if err := updatePGClusterDatabases(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("user") {
		if err := updatePGClusterUsersUpdateAndDrop(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host") {
		if err := updatePGClusterHosts(d, meta); err != nil {
			return err
		}
	}

	if d.HasChange("host_master_name") {

		if err := updateMasterPGClusterHosts(d, meta); err != nil {
			return err
		}
	}

	d.Partial(false)

	return resourceYandexMDBPostgreSQLClusterRead(d, meta)
}

func updatePGClusterParams(d *schema.ResourceData, meta interface{}) error {
	req, updateFieldConfigName, err := getPGClusterUpdateRequest(d)
	if err != nil {
		return err
	}

	mdbPGUpdateFieldsMap := map[string]string{
		"name":                               "name",
		"description":                        "description",
		"labels":                             "labels",
		"config.0.version":                   "config_spec.version",
		"config.0.autofailover":              "config_spec.autofailover",
		"config.0.pooler_config":             "config_spec.pooler_config",
		"config.0.access":                    "config_spec.access",
		"config.0.performance_diagnostics":   "config_spec.performance_diagnostics",
		"config.0.backup_window_start":       "config_spec.backup_window_start",
		"config.0.resources":                 "config_spec.resources",
		"config.0.backup_retain_period_days": "config_spec.backup_retain_period_days",
		"security_group_ids":                 "security_group_ids",
		"maintenance_window":                 "maintenance_window",
		"deletion_protection":                "deletion_protection",
	}

	if updateFieldConfigName != "" {
		mdbPGUpdateFieldsMap["config.0.postgresql_config"] = "config_spec." + updateFieldConfigName
	}

	onDone := []func(){}
	updatePath := []string{}
	for field, path := range mdbPGUpdateFieldsMap {
		if d.HasChange(field) {
			updatePath = append(updatePath, path)
			onDone = append(onDone, func() {

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

func getPGClusterUpdateRequest(d *schema.ResourceData) (ucr *postgresql.UpdateClusterRequest, updateFieldConfigName string, err error) {
	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return nil, updateFieldConfigName, fmt.Errorf("error expanding labels while updating PostgreSQL Cluster: %s", err)
	}

	configSpec, updateFieldConfigName, err := expandPGConfigSpec(d)
	if err != nil {
		return nil, updateFieldConfigName, fmt.Errorf("error expanding config while updating PostgreSQL Cluster: %s", err)
	}

	securityGroupIds := expandSecurityGroupIds(d.Get("security_group_ids"))

	maintenanceWindow, err := expandPGMaintenanceWindow(d)
	if err != nil {
		return nil, updateFieldConfigName, fmt.Errorf("error expanding maintenance_window while updating PostgreSQL cluster: %s", err)
	}

	req := &postgresql.UpdateClusterRequest{
		ClusterId:          d.Id(),
		Name:               d.Get("name").(string),
		Description:        d.Get("description").(string),
		Labels:             labels,
		ConfigSpec:         configSpec,
		MaintenanceWindow:  maintenanceWindow,
		SecurityGroupIds:   securityGroupIds,
		DeletionProtection: d.Get("deletion_protection").(bool),
	}

	return req, updateFieldConfigName, nil
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

	err = validateNoUpdatingOwner(currDBs, targetDBs)
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

func validateNoUpdatingOwner(currentDatabases []*postgresql.Database, targetDatabases []*postgresql.DatabaseSpec) error {
	for _, currentDatabase := range currentDatabases {
		for _, targetDatabase := range targetDatabases {
			if currentDatabase.Name == targetDatabase.Name &&
				(currentDatabase.Owner != targetDatabase.Owner) {
				return fmt.Errorf("impossible to change owner for PostgreSQL Cluster database %s", currentDatabase.Name)
			}
		}
	}
	return nil
}

func updatePGClusterUsersAdd(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currUsers, err := listPGUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	usersForCreate, err := pgUserForCreate(d, currUsers)
	if err != nil {
		return err
	}
	for _, u := range usersForCreate {
		err := createPGUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePGClusterUsersUpdateAndDrop(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currUsers, err := listPGUsers(ctx, config, d.Id())
	if err != nil {
		return err
	}

	dUser := make(map[string]string)
	cnt := d.Get("user.#").(int)
	for i := 0; i < cnt; i++ {
		dUser[d.Get(fmt.Sprintf("user.%v.name", i)).(string)] = fmt.Sprintf("user.%v.", i)
	}

	deleteNames := make([]string, 0)

	for _, v := range currUsers {
		path, ok := dUser[v.Name]
		if !ok {
			deleteNames = append(deleteNames, v.Name)
		} else {
			err := updatePGUser(ctx, config, d, v, path)
			if err != nil {
				return err
			}
		}
	}

	for _, u := range deleteNames {
		err := deletePGUser(ctx, config, d, u)
		if err != nil {
			return err
		}
	}

	return nil
}

func updatePGClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	err := createPGClusterHosts(ctx, config, d)
	if err != nil {
		return err
	}

	currHosts, err := listPGHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	compareHostsInfo, err := comparePGHostsInfo(d, currHosts, true)
	if err != nil {
		return err
	}

	hostsToDelete := []string{}

	for _, hostInfo := range compareHostsInfo.hostsInfo {
		if !hostInfo.isNew {
			hostsToDelete = append(hostsToDelete, hostInfo.fqdn)
		} else if compareHostsInfo.haveHostWithName {
			var maskPaths []string
			if hostInfo.oldPriority != hostInfo.newPriority {
				maskPaths = append(maskPaths, "priority")
			}
			if hostInfo.oldReplicationSource != hostInfo.newReplicationSource {
				maskPaths = append(maskPaths, "replication_source")
			}
			if hostInfo.oldAssignPublicIP != hostInfo.newAssignPublicIP {
				maskPaths = append(maskPaths, "assign_public_ip")
			}
			if len(maskPaths) > 0 {
				if err := updatePGHost(ctx, config, d, &postgresql.UpdateHostSpec{
					HostName:          hostInfo.fqdn,
					ReplicationSource: hostInfo.newReplicationSource,
					Priority:          &wrappers.Int64Value{Value: int64(hostInfo.newPriority)},
					AssignPublicIp:    hostInfo.newAssignPublicIP,
					UpdateMask:        &field_mask.FieldMask{Paths: maskPaths},
				}); err != nil {
					return err
				}
			}
		}
	}

	if err := deletePGHosts(ctx, config, d, hostsToDelete); err != nil {
		return err
	}

	return nil
}

func createPGClusterHosts(ctx context.Context, config *Config, d *schema.ResourceData) error {

	currHosts, err := listPGHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}

	compareHostsInfo, err := comparePGHostsInfo(d, currHosts, true)

	if err != nil {
		return err
	}

	if compareHostsInfo.hierarchyExists && len(compareHostsInfo.createHostsInfo) == 0 {
		return fmt.Errorf("Create cluster hosts error. Exists host with replication source, which can't be created. Possibly there is a loop")
	}

	if compareHostsInfo.haveHostWithName {
		for _, newHostInfo := range compareHostsInfo.createHostsInfo {
			err := addPGHost(ctx, config, d, &postgresql.HostSpec{
				ZoneId:            newHostInfo.zone,
				SubnetId:          newHostInfo.subnetID,
				AssignPublicIp:    newHostInfo.oldAssignPublicIP,
				ReplicationSource: newHostInfo.newReplicationSource,
				Priority:          &wrappers.Int64Value{Value: int64(newHostInfo.newPriority)},
			})
			if err != nil {
				return err
			}
		}
	} else {
		for _, newHostInfo := range compareHostsInfo.createHostsInfo {
			err := addPGHost(ctx, config, d, &postgresql.HostSpec{
				ZoneId:         newHostInfo.zone,
				SubnetId:       newHostInfo.subnetID,
				AssignPublicIp: newHostInfo.oldAssignPublicIP,
			})
			if err != nil {
				return err
			}
		}
	}

	if compareHostsInfo.hierarchyExists {
		return createPGClusterHosts(ctx, config, d)
	}

	return nil
}

func updateMasterPGClusterHosts(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	currHosts, err := listPGHosts(ctx, config, d.Id())
	if err != nil {
		return err
	}
	compareHostsInfo, err := comparePGHostsInfo(d, currHosts, true)
	if err != nil {
		return err
	}

	if !compareHostsInfo.haveHostWithName {
		return nil
	}

	for _, hostInfo := range compareHostsInfo.hostsInfo {
		if compareHostsInfo.hostMasterName == hostInfo.name && hostInfo.role != postgresql.Host_MASTER {
			err = startPGFailover(ctx, config, d, hostInfo.fqdn)
			if err != nil {
				return err
			}
			break
		}
	}

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

func updatePGUser(ctx context.Context, config *Config, d *schema.ResourceData, user *postgresql.User, path string) (err error) {

	us, err := expandPGUser(d, &postgresql.UserSpec{
		Name:        user.Name,
		Permissions: user.Permissions,
		ConnLimit:   &wrappers.Int64Value{Value: user.ConnLimit},
		Settings:    user.Settings,
		Login:       user.Login,
		Grants:      user.Grants,
	}, path)
	if err != nil {
		return err
	}

	changeMask := map[string]string{
		"password":   "password",
		"permission": "permissions",
		"login":      "login",
		"grants":     "grants",
		"conn_limit": "conn_limit",
		"settings":   "settings",
	}

	updatePath := []string{}
	onDone := make([]func(), 0)

	for field, mask := range changeMask {
		if d.HasChange(path + field) {
			updatePath = append(updatePath, mask)
			onDone = append(onDone, func() {

			})
		}
	}

	if len(updatePath) == 0 {
		return nil
	}

	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().User().Update(ctx, &postgresql.UpdateUserRequest{
			ClusterId:   d.Id(),
			UserName:    us.Name,
			Password:    us.Password,
			Permissions: us.Permissions,
			ConnLimit:   us.ConnLimit.GetValue(),
			Login:       us.Login,
			Grants:      us.Grants,
			Settings:    us.Settings,
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

func deletePGHosts(ctx context.Context, config *Config, d *schema.ResourceData, hostNamesToDelete []string) error {
	if len(hostNamesToDelete) == 0 {
		return nil
	}
	for _, hostToDelete := range hostNamesToDelete {
		if err := deletePGHost(ctx, config, d, hostToDelete); err != nil {
			return err
		}
	}

	return nil
}

func deletePGHost(ctx context.Context, config *Config, d *schema.ResourceData, name string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Cluster().DeleteHosts(ctx, &postgresql.DeleteClusterHostsRequest{
			ClusterId: d.Id(),
			HostNames: []string{name},
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

func startPGFailover(ctx context.Context, config *Config, d *schema.ResourceData, hostName string) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Cluster().StartFailover(ctx, &postgresql.StartClusterFailoverRequest{
			ClusterId: d.Id(),
			HostName:  hostName,
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to start failover host in PostgreSQL Cluster %q - host %v: %s", d.Id(), hostName, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while start failover host in PostgreSQL Cluster %q - host %v: %s", d.Id(), hostName, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("start failover host in PostgreSQL Cluster %q - host %v failed: %s", d.Id(), hostName, err)
	}

	return nil
}

func updatePGHost(ctx context.Context, config *Config, d *schema.ResourceData, host *postgresql.UpdateHostSpec) error {
	op, err := config.sdk.WrapOperation(
		config.sdk.MDB().PostgreSQL().Cluster().UpdateHosts(ctx, &postgresql.UpdateClusterHostsRequest{
			ClusterId:       d.Id(),
			UpdateHostSpecs: []*postgresql.UpdateHostSpec{host},
		}),
	)
	if err != nil {
		return fmt.Errorf("error while requesting API to update host for PostgreSQL Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error while updating host for PostgreSQL Cluster %q - host %v: %s", d.Id(), host.HostName, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating host for PostgreSQL Cluster %q - host %v failed: %s", d.Id(), host.HostName, err)
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

func setPGFolderID(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	cluster, err := config.sdk.MDB().PostgreSQL().Cluster().Get(ctx, &postgresql.GetClusterRequest{
		ClusterId: d.Id(),
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Cluster %q", d.Id()))
	}

	folderID, ok := d.GetOk("folder_id")
	if !ok {
		return nil
	}
	if folderID == "" {
		return nil
	}

	if cluster.FolderId != folderID {

		op, err := config.sdk.WrapOperation(
			config.sdk.MDB().PostgreSQL().Cluster().Move(ctx, &postgresql.MoveClusterRequest{
				ClusterId:           d.Id(),
				DestinationFolderId: folderID.(string),
			}),
		)
		if err != nil {
			return fmt.Errorf("error while requesting API to move PostgreSQL Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error while moving PostgreSQL Cluster %q to folder %v: %s", d.Id(), folderID, err)
		}

		if _, err := op.Response(); err != nil {
			return fmt.Errorf("moving PostgreSQL Cluster %q to folder %v failed: %s", d.Id(), folderID, err)
		}

	}

	return nil
}
