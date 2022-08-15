package yandex

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

const (
	yandexMDBPostgreSQLDatabaseCreateTimeout = 10 * time.Minute
	yandexMDBPostgreSQLDatabaseReadTimeout   = 1 * time.Minute
	yandexMDBPostgreSQLDatabaseUpdateTimeout = 10 * time.Minute
	yandexMDBPostgreSQLDatabaseDeleteTimeout = 10 * time.Minute
)

func resourceYandexMDBPostgreSQLDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBPostgreSQLDatabaseCreate,
		Read:   resourceYandexMDBPostgreSQLDatabaseRead,
		Update: resourceYandexMDBPostgreSQLDatabaseUpdate,
		Delete: resourceYandexMDBPostgreSQLDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBPostgreSQLDatabaseCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBPostgreSQLDatabaseReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBPostgreSQLDatabaseUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBPostgreSQLDatabaseDeleteTimeout),
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"owner": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"lc_collate": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "C",
			},
			"lc_type": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
				Default:  "C",
			},
			"template_db": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
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
	}
}

func resourceYandexMDBPostgreSQLDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	databaseSpec, err := expandPgDatabaseSpec(d)
	if err != nil {
		return err
	}
	request := &postgresql.CreateDatabaseRequest{
		ClusterId:    clusterID,
		DatabaseSpec: databaseSpec,
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL database create request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Database().Create(ctx, request)
	})

	databaseID := constructResourceId(request.ClusterId, request.DatabaseSpec.Name)
	d.SetId(databaseID)

	if err != nil {
		return fmt.Errorf("error while requesting API to create database in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while adding database to PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating database for PostgreSQL Cluster %q failed: %s", clusterID, err)
	}

	return resourceYandexMDBPostgreSQLDatabaseRead(d, meta)
}

func expandPgDatabaseSpec(d *schema.ResourceData) (*postgresql.DatabaseSpec, error) {
	out := &postgresql.DatabaseSpec{}

	if v, ok := d.GetOk("name"); ok {
		out.Name = v.(string)
	}

	if v, ok := d.GetOk("owner"); ok {
		out.Owner = v.(string)
	}

	if v, ok := d.GetOk("lc_collate"); ok {
		out.LcCollate = v.(string)
	}

	if v, ok := d.GetOk("template_db"); ok {
		out.TemplateDb = v.(string)
	}

	if v, ok := d.GetOk("lc_type"); ok {
		out.LcCtype = v.(string)
	}

	if v, ok := d.GetOk("extension"); ok {
		es := v.(*schema.Set).List()
		out.Extensions = expandPGExtensions(es)
	}

	return out, nil
}

func resourceYandexMDBPostgreSQLDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	clusterID, dbname, err := deconstructResourceId(d.Id())
	if err != nil {
		return err
	}

	db, err := config.sdk.MDB().PostgreSQL().Database().Get(ctx, &postgresql.GetDatabaseRequest{
		ClusterId:    clusterID,
		DatabaseName: dbname,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Database %q", dbname))
	}

	d.Set("cluster_id", clusterID)
	d.Set("name", db.Name)
	d.Set("owner", db.Owner)
	d.Set("lc_collate", db.LcCollate)
	d.Set("lc_type", db.LcCtype)
	d.Set("template_db", db.TemplateDb)
	d.Set("extension", flattenPGExtensions(db.Extensions))
	return nil
}

func resourceYandexMDBPostgreSQLDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	extensions := []*postgresql.Extension{}
	if v, ok := d.GetOk("extension"); ok {
		es := v.(*schema.Set).List()
		extensions = expandPGExtensions(es)
	}

	request := &postgresql.UpdateDatabaseRequest{
		ClusterId:    clusterID,
		DatabaseName: d.Get("name").(string),
		Extensions:   extensions,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL database update request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Database().Update(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to update database in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while updating database in PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("updating database for PostgreSQL Cluster %q failed: %s", clusterID, err)
	}

	return nil
}

func resourceYandexMDBPostgreSQLDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	dbName := d.Get("name").(string)
	clusterID := d.Get("cluster_id").(string)

	request := &postgresql.DeleteDatabaseRequest{
		ClusterId:    clusterID,
		DatabaseName: dbName,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending PostgreSQL database delete request: %+v", request)
		return config.sdk.MDB().PostgreSQL().Database().Delete(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while deleting database from PostgreSQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting database from PostgreSQL Cluster %q failed: %s", clusterID, err)
	}

	return nil
}
