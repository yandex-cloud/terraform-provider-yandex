package object_storage_connection

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AttributeConnectionName: &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeServiceAccountID: &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		AttributeBucket: &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeDescription: &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		AttributeVisibility: &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
	}
}
