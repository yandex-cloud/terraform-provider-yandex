package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceYandexYDBDatabaseServerless() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceYandexYDBDatabaseServerlessRead,

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"database_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"folder_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"location_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"labels": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"document_api_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
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

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceYandexYDBDatabaseServerlessRead(d *schema.ResourceData, meta interface{}) error {
	database, err := dataSourceYandexYDBDatabaseRead(d, meta)
	if err != nil {
		return err
	}

	return flattenYandexYDBDatabaseServerless(d, database)
}
