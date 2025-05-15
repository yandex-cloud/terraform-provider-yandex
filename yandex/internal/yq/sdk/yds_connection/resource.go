package yds_connection

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AttributeName: &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeServiceAccountID: &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
		AttributeDatabaseID: &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeDescription: &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
			Optional: true,
		},
	}
}
