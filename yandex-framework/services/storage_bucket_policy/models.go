package storage_bucket_policy

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StorageBucketPolicyResourceModel struct {
	Bucket    types.String `tfsdk:"bucket"`
	Policy    PolicyValue  `tfsdk:"policy"`
	AccessKey types.String `tfsdk:"access_key"`
	SecretKey types.String `tfsdk:"secret_key"`
}
