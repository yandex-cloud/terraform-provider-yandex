package yandex

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

const (
	yandexMDBMySQLDatabaseCreateTimeout = 10 * time.Minute
	yandexMDBMySQLDatabaseReadTimeout   = 1 * time.Minute
	yandexMDBMySQLDatabaseUpdateTimeout = 10 * time.Minute
	yandexMDBMySQLDatabaseDeleteTimeout = 10 * time.Minute
)

func resourceYandexMDBMySQLDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexMDBMySQLDatabaseCreate,
		Read:   resourceYandexMDBMySQLDatabaseRead,
		Update: resourceYandexMDBMySQLDatabaseUpdate,
		Delete: resourceYandexMDBMySQLDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexMDBMySQLDatabaseCreateTimeout),
			Read:   schema.DefaultTimeout(yandexMDBMySQLDatabaseReadTimeout),
			Update: schema.DefaultTimeout(yandexMDBMySQLDatabaseUpdateTimeout),
			Delete: schema.DefaultTimeout(yandexMDBMySQLDatabaseDeleteTimeout),
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
		},
	}
}

func resourceYandexMDBMySQLDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	clusterID := d.Get("cluster_id").(string)
	request := &mysql.CreateDatabaseRequest{
		ClusterId: clusterID,
		DatabaseSpec: &mysql.DatabaseSpec{
			Name: d.Get("name").(string),
		},
	}

	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending MySQL database create request: %+v", request)
		return config.sdk.MDB().MySQL().Database().Create(ctx, request)
	})

	databaseID := constructResourceId(request.ClusterId, request.DatabaseSpec.Name)
	d.SetId(databaseID)

	if err != nil {
		return fmt.Errorf("error while requesting API to create database in MySQL Cluster %q: %s", clusterID, err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while adding database to MySQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("creating database for MySQL Cluster %q failed: %s", clusterID, err)
	}

	return resourceYandexMDBMySQLDatabaseRead(d, meta)
}

func resourceYandexMDBMySQLDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	clusterID, dbname, err := deconstructResourceId(d.Id())
	if err != nil {
		return err
	}

	db, err := config.sdk.MDB().MySQL().Database().Get(ctx, &mysql.GetDatabaseRequest{
		ClusterId:    clusterID,
		DatabaseName: dbname,
	})
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("Database %q", dbname))
	}

	d.Set("cluster_id", clusterID)
	d.Set("name", db.Name)
	return nil
}

func resourceYandexMDBMySQLDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	return fmt.Errorf("changing resource_yandex_mdb_mysql_database is not supported")
}

func resourceYandexMDBMySQLDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutDelete))
	defer cancel()

	dbname := d.Get("name").(string)
	clusterID := d.Get("cluster_id").(string)

	request := &mysql.DeleteDatabaseRequest{
		ClusterId:    clusterID,
		DatabaseName: dbname,
	}
	op, err := retryConflictingOperation(ctx, config, func() (*operation.Operation, error) {
		log.Printf("[DEBUG] Sending MySQL database delete request: %+v", request)
		return config.sdk.MDB().MySQL().Database().Delete(ctx, request)
	})

	if err != nil {
		return fmt.Errorf("error while requesting API to delete database from MySQL Cluster %q: %s", clusterID, err)
	}

	if err := op.Wait(ctx); err != nil {
		return fmt.Errorf("error while deleting database from MySQL Cluster %q: %s", clusterID, err)
	}

	if _, err := op.Response(); err != nil {
		return fmt.Errorf("deleting database from MySQL Cluster %q failed: %s", clusterID, err)
	}

	return nil
}
