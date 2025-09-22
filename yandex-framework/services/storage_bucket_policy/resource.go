package storage_bucket_policy

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	storage "github.com/yandex-cloud/terraform-provider-yandex/pkg/storage/s3"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	_ resource.Resource                = &storageBucketPolicyResource{}
	_ resource.ResourceWithConfigure   = &storageBucketPolicyResource{}
	_ resource.ResourceWithImportState = &storageBucketPolicyResource{}
)

type storageBucketPolicyResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &storageBucketPolicyResource{}
}

func (r *storageBucketPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_bucket_policy"
}

func (r *storageBucketPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *storageBucketPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func (r *storageBucketPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}

func (r *storageBucketPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StorageBucketPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateBucketPolicy(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readBucketPolicy(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageBucketPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StorageBucketPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readBucketPolicy(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageBucketPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StorageBucketPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateBucketPolicy(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageBucketPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StorageBucketPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Clear policy by setting empty policy
	state.Policy = PolicyValue(types.StringNull())

	r.updateBucketPolicy(ctx, &state, &resp.Diagnostics)
}

func (r *storageBucketPolicyResource) updateBucketPolicy(ctx context.Context, model *StorageBucketPolicyResourceModel, diags *diag.Diagnostics) {
	s3Client, err := r.getS3Client(ctx, model)
	if err != nil {
		diags.AddError("Error getting storage client", err.Error())
		return
	}

	bucket := model.Bucket.ValueString()
	policy := model.Policy.ValueString()

	err = s3Client.UpdateBucketPolicy(ctx, bucket, policy)
	if err != nil {
		diags.AddError("Error updating bucket policy", err.Error())
		return
	}
}

func (r *storageBucketPolicyResource) readBucketPolicy(ctx context.Context, model *StorageBucketPolicyResourceModel, diags *diag.Diagnostics) {
	s3Client, err := r.getS3Client(ctx, model)
	if err != nil {
		diags.AddError("Error getting storage client", err.Error())
		return
	}

	bucketName := model.Bucket.ValueString()

	policy, err := s3Client.GetBucketPolicy(ctx, bucketName)
	if err != nil {
		diags.AddError("Unable to read Storage Bucket Policy", err.Error())
		return
	}

	model.Policy = PolicyValue(types.StringValue(policy))
}

func (r *storageBucketPolicyResource) getS3Client(ctx context.Context, model *StorageBucketPolicyResourceModel) (*storage.Client, error) {
	return r.providerConfig.GetS3Client(ctx, model.AccessKey.ValueString(), model.SecretKey.ValueString())
}
