package storage_bucket_grant

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StorageBucketGrantResourceModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
	ACL       types.String `tfsdk:"acl"`
	Grants    types.Set    `tfsdk:"grant"`
}

type StorageBucketGrantModel struct {
	Id          types.String `tfsdk:"id"`
	Uri         types.String `tfsdk:"uri"`
	Permissions types.Set    `tfsdk:"permissions"`
	Type        types.String `tfsdk:"type"`
}
