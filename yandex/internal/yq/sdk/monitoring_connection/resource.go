package monitoring_connection

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		AttributeName: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeServiceAccountID: {
			Type:     schema.TypeString,
			Optional: true,
		},
		AttributeProject: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeCluster: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeDescription: {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
