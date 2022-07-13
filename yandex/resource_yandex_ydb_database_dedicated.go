package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/ydb/v1"
	"google.golang.org/genproto/protobuf/field_mask"
)

const yandexYDBDedicatedDefaultTimeout = 10 * time.Minute

func resourceYandexYDBDatabaseDedicated() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexYDBDatabaseDedicatedCreate,
		Read:   resourceYandexYDBDatabaseDedicatedRead,
		Update: resourceYandexYDBDatabaseDedicatedUpdate,
		Delete: performYandexYDBDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(yandexYDBDedicatedDefaultTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"network_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"resource_preset_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"scale_policy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fixed_scale": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"size": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
					},
				},
			},

			"storage_config": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage_type_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.NoZeroValues,
						},
						"group_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},

			"location": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"region": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.NoZeroValues,
									},
								},
							},
						},
					},
				},
			},

			"location_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"assign_public_ips": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"folder_id": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
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

			"ydb_full_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ydb_api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"database_path": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tls_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceYandexYDBDatabaseDedicatedCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	folderID, err := getFolderID(d, config)
	if err != nil {
		return fmt.Errorf("Error getting folder ID while creating database: %s", err)
	}

	labels, err := expandLabels(d.Get("labels"))
	if err != nil {
		return fmt.Errorf("Error expanding labels while creating database: %s", err)
	}

	subnetIDs := convertStringSet(d.Get("subnet_ids").(*schema.Set))
	if len(subnetIDs) == 0 {
		return fmt.Errorf("Error expanding subnet IDs while creating database: %s", err)
	}
	for _, subnetID := range subnetIDs {
		if len(subnetID) == 0 {
			return fmt.Errorf("Error checking subnet IDs while creating database: %s", err)
		}
	}

	storageConfig, err := expandYDBStorageConfigSpec(d)
	if err != nil {
		return fmt.Errorf("Error expanding storage configuration while creating database: %s", err)
	}

	scalePolicy, err := expandYDBScalePolicySpec(d)
	if err != nil {
		return fmt.Errorf("Error expanding scale policy while creating database: %s", err)
	}

	dbType, err := expandYDBLocationSpec(d)
	if err != nil {
		return fmt.Errorf("Error expanding database type while creating database: %s", err)
	}
	req := ydb.CreateDatabaseRequest{
		FolderId:         folderID,
		Name:             d.Get("name").(string),
		DatabaseType:     dbType,
		Description:      d.Get("description").(string),
		ResourcePresetId: d.Get("resource_preset_id").(string),
		StorageConfig:    storageConfig,
		ScalePolicy:      scalePolicy,
		NetworkId:        d.Get("network_id").(string),
		SubnetIds:        subnetIDs,
		AssignPublicIps:  d.Get("assign_public_ips").(bool),
		LocationId:       d.Get("location_id").(string),
		Labels:           labels,
	}

	if err := performYandexYDBDatabaseCreate(d, config, &req); err != nil {
		return err
	}

	return resourceYandexYDBDatabaseDedicatedRead(d, meta)
}

func performYandexYDBDatabaseCreate(d *schema.ResourceData, config *Config, req *ydb.CreateDatabaseRequest) error {
	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.YDB().Database().Create(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to create database: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("Error while get database create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*ydb.CreateDatabaseMetadata)
	if !ok {
		return fmt.Errorf("could not get database ID from create operation metadata")
	}

	d.SetId(md.DatabaseId)

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error while waiting operation to create database: %s", err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("Database creation failed: %s", err)
	}
	return nil
}

func resourceYandexYDBDatabaseDedicatedUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	req := ydb.UpdateDatabaseRequest{
		DatabaseId: d.Id(),
		UpdateMask: &field_mask.FieldMask{},
	}

	if d.HasChange("assign_public_ips") {
		req.AssignPublicIps = d.Get("assign_public_ips").(bool)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "assign_public_ips")
	}

	if d.HasChange("resource_preset_id") {
		req.ResourcePresetId = d.Get("resource_preset_id").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "resource_preset_id")
	}

	if d.HasChange("storage_config") {
		storageConfig, err := expandYDBStorageConfigSpec(d)
		if err != nil {
			return fmt.Errorf("Error expanding storage configuration while updating database: %s", err)
		}
		req.StorageConfig = storageConfig
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "storage_config")
	}

	if d.HasChange("scale_policy") {
		scalePolicy, err := expandYDBScalePolicySpec(d)
		if err != nil {
			return fmt.Errorf("Error expanding scale policy while updating database: %s", err)
		}
		req.ScalePolicy = scalePolicy
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "scale_policy")
	}

	if err := performYandexYDBDatabaseUpdate(d, config, &req); err != nil {
		return err
	}

	return resourceYandexYDBDatabaseDedicatedRead(d, meta)
}

func performYandexYDBDatabaseUpdate(d *schema.ResourceData, config *Config, req *ydb.UpdateDatabaseRequest) error {
	d.Partial(true)
	// common parameters
	if d.HasChange("labels") {
		labelsProp, err := expandLabels(d.Get("labels"))
		if err != nil {
			return err
		}

		req.Labels = labelsProp
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "labels")
	}

	if d.HasChange("name") {
		req.Name = d.Get("name").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "name")
	}

	if d.HasChange("description") {
		req.Description = d.Get("description").(string)
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "description")
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	op, err := config.sdk.WrapOperation(config.sdk.YDB().Database().Update(ctx, req))
	if err != nil {
		return fmt.Errorf("Error while requesting API to update database: %s", err)
	}

	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("Error updating database %q: %s", d.Id(), err)
	}

	d.Partial(false)

	return nil
}

func performYandexYDBDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	op, err := config.sdk.YDB().Database().Delete(ctx, &ydb.DeleteDatabaseRequest{DatabaseId: d.Id()})
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("YDB Database %q", d.Id()))
	}

	return nil
}

func resourceYandexYDBDatabaseDedicatedRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	database, err := performYandexYDBDatabaseRead(d, config)
	if err != nil {
		return err
	}

	return flattenYandexYDBDatabaseDedicated(d, database)
}

func performYandexYDBDatabaseRead(d *schema.ResourceData, config *Config) (*ydb.Database, error) {
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	database, err := config.sdk.YDB().Database().Get(ctx, &ydb.GetDatabaseRequest{
		DatabaseId: d.Id(),
	})
	if err != nil {
		return nil, handleNotFoundError(err, d, fmt.Sprintf("YDB Database %q", d.Get("name").(string)))
	}

	return database, nil
}

func flattenYandexYDBDatabaseDedicated(d *schema.ResourceData, database *ydb.Database) error {
	if database == nil {
		// NOTE(shmel1k@): database existed before but was removed outside of terraform.
		d.SetId("")
		return nil
	}

	switch database.DatabaseType.(type) {
	case *ydb.Database_RegionalDatabase,
		*ydb.Database_ZonalDatabase,
		*ydb.Database_DedicatedDatabase: // we actually expect it
	case *ydb.Database_ServerlessDatabase:
		return fmt.Errorf("expect dedicated database, got serverless")
	default:
		return fmt.Errorf("unknown database type")
	}

	location, err := flattenYDBLocation(database)
	if err != nil {
		return err
	}
	d.Set("location", location)

	d.Set("assign_public_ips", database.AssignPublicIps)
	d.Set("resource_preset_id", database.ResourcePresetId)

	d.Set("network_id", database.NetworkId)
	if err := d.Set("subnet_ids", database.SubnetIds); err != nil {
		return err
	}

	storageConfig, err := flattenYDBStorageConfig(database.StorageConfig)
	if err != nil {
		return err
	}

	if err := d.Set("storage_config", storageConfig); err != nil {
		return err
	}

	scalePolicy, err := flattenYDBScalePolicy(database)
	if err != nil {
		return err
	}
	if err := d.Set("scale_policy", scalePolicy); err != nil {
		return err
	}

	return flattenYandexYDBDatabase(d, database)
}

func flattenYandexYDBDatabase(d *schema.ResourceData, database *ydb.Database) error {
	baseEP, dbPath, useTLS, err := parseYandexYDBDatabaseEndpoint(database.Endpoint)
	if err != nil {
		return err
	}

	d.Set("name", database.Name)
	d.Set("folder_id", database.FolderId)
	d.Set("description", database.Description)
	d.Set("created_at", getTimestamp(database.CreatedAt))
	if err := d.Set("labels", database.Labels); err != nil {
		return err
	}
	d.Set("location_id", database.LocationId)
	d.Set("ydb_full_endpoint", database.Endpoint)
	d.Set("ydb_api_endpoint", baseEP)
	d.Set("database_path", dbPath)
	d.Set("tls_enabled", useTLS)

	return d.Set("status", database.Status.String())
}

func parseYandexYDBDatabaseEndpoint(endpoint string) (baseEP, databasePath string, useTLS bool, err error) {
	dbSplit := strings.Split(endpoint, "/?database=")
	if len(dbSplit) != 2 {
		return "", "", false, fmt.Errorf("cannot parse endpoint %q", endpoint)
	}
	parts := strings.SplitN(dbSplit[0], "/", 3)
	if len(parts) < 3 {
		return "", "", false, fmt.Errorf("cannot parse endpoint schema %q", dbSplit[0])
	}

	const (
		protocolGRPCS = "grpcs:"
		protocolGRPC  = "grpc:"
	)

	switch protocol := parts[0]; protocol {
	case protocolGRPCS:
		useTLS = true
	case protocolGRPC:
		useTLS = false
	default:
		return "", "", false, fmt.Errorf("unknown protocol %q", protocol)
	}
	return parts[2], dbSplit[1], useTLS, nil
}
