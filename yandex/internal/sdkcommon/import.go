package sdkcommon

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

func AddPassthroughImport(r *schema.Resource) *schema.Resource {
	r.Importer = &schema.ResourceImporter{
		State: schema.ImportStatePassthrough,
	}
	return r
}
