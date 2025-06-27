package storage_bucket_grant

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	storage "github.com/yandex-cloud/terraform-provider-yandex/pkg/storage/s3"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

var (
	_ resource.Resource                = &storageBucketGrantResource{}
	_ resource.ResourceWithConfigure   = &storageBucketGrantResource{}
	_ resource.ResourceWithImportState = &storageBucketGrantResource{}
)

type storageBucketGrantResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &storageBucketGrantResource{}
}

func (r *storageBucketGrantResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_bucket_grant"
}

func (r *storageBucketGrantResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = ResourceSchema(ctx)
}

func (r *storageBucketGrantResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *storageBucketGrantResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("bucket"), req, resp)
}

func (r *storageBucketGrantResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StorageBucketGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateBucketGrant(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readBucketGrant(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageBucketGrantResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StorageBucketGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readBucketGrant(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *storageBucketGrantResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StorageBucketGrantResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.updateBucketGrant(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	r.readBucketGrant(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *storageBucketGrantResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StorageBucketGrantResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Clear ACL and grants by setting to private
	state.ACL = types.StringValue(storage.BucketACLPrivate)
	state.Grants = []StorageBucketGrantModel{}

	r.updateBucketGrant(ctx, &state, &resp.Diagnostics)
}

func (r *storageBucketGrantResource) updateBucketGrant(ctx context.Context, model *StorageBucketGrantResourceModel, diags *diag.Diagnostics) {
	s3Client, err := r.getS3Client(ctx, model)
	if err != nil {
		diags.AddError("Error getting storage client", err.Error())
		return
	}

	bucket := model.Bucket.ValueString()

	// Update grants if specified
	if len(model.Grants) > 0 {
		// Convert model grants to S3 grants
		awsGrants, diags := r.modelGrantsToS3Grants(model.Grants)
		if diags.HasError() {
			return
		}

		tflog.Debug(ctx, "Updating bucket grants", map[string]interface{}{
			"bucket":       bucket,
			"grants_count": len(awsGrants),
		})

		err = s3Client.UpdateBucketGrants(ctx, bucket, awsGrants)
		if err != nil {
			diags.AddError("Error updating bucket grants", err.Error())
			return
		}
	} else if !model.ACL.IsNull() && !model.ACL.IsUnknown() {
		// Fallback to ACL if no grants specified
		acl := model.ACL.ValueString()
		if acl == "" {
			acl = storage.BucketACLPrivate
		}

		tflog.Debug(ctx, "Updating bucket ACL", map[string]interface{}{
			"bucket": bucket,
			"acl":    acl,
		})

		input := &s3.PutBucketAclInput{
			Bucket: &bucket,
			ACL:    &acl,
		}

		err := s3Client.UpdateBucketACL(ctx, input)

		if err != nil {
			diags.AddError("Error updating bucket ACL", err.Error())
			return
		}
	}
}

func (r *storageBucketGrantResource) readBucketGrant(ctx context.Context, model *StorageBucketGrantResourceModel, diags *diag.Diagnostics) {
	s3Client, err := r.getS3Client(ctx, model)
	if err != nil {
		diags.AddError("Error getting storage client", err.Error())
		return
	}

	bucketName := model.Bucket.ValueString()

	aclOutput, err := s3Client.GetBucketACL(ctx, bucketName)
	if err != nil {
		diags.AddError("Unable to read Storage Bucket ACL", err.Error())
		return
	}

	grants, d := r.s3GrantsToModelGrants(aclOutput.Grants)
	diags.Append(d...)
	if diags.HasError() {
		return
	}
	model.Grants = grants
}

func (r *storageBucketGrantResource) getS3Client(ctx context.Context, model *StorageBucketGrantResourceModel) (*storage.Client, error) {
	accessKey := ""
	secretKey := ""

	if !model.AccessKey.IsNull() && !model.AccessKey.IsUnknown() {
		accessKey = model.AccessKey.ValueString()
	}
	if !model.SecretKey.IsNull() && !model.SecretKey.IsUnknown() {
		secretKey = model.SecretKey.ValueString()
	}

	return storage.GetS3Client(ctx, accessKey, secretKey, r.providerConfig)
}

// Convert Terraform model grants to S3 grant structures
func (r *storageBucketGrantResource) modelGrantsToS3Grants(modelGrants []StorageBucketGrantModel) ([]*s3.Grant, diag.Diagnostics) {
	var diags diag.Diagnostics
	grants := make([]*s3.Grant, 0)

	for _, mg := range modelGrants {
		granteeType := mg.Type.ValueString()
		granteeID := mg.Id.ValueString()
		granteeURI := mg.Uri.ValueString()

		// Get permissions
		var permissions []string
		if !mg.Permissions.IsNull() && !mg.Permissions.IsUnknown() {
			diags.Append(mg.Permissions.ElementsAs(context.Background(), &permissions, false)...)
			if diags.HasError() {
				return nil, diags
			}
		}

		if len(permissions) == 0 {
			diags.AddError("Missing permissions", "Grant must have at least one permission")
			return nil, diags
		}

		grantee := &s3.Grantee{
			Type: &granteeType,
		}

		if granteeID != "" {
			grantee.ID = &granteeID
		}
		if granteeURI != "" {
			grantee.URI = &granteeURI
		}

		for _, perm := range permissions {
			grants = append(grants, &s3.Grant{
				Grantee:    grantee,
				Permission: &perm,
			})
		}
	}

	return grants, diags
}

// Convert S3 grants to Terraform model grants
func (r *storageBucketGrantResource) s3GrantsToModelGrants(s3Grants []*s3.Grant) ([]StorageBucketGrantModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Create a map to group grants by unique Grantee (Type + Id + Uri)
	granteeMap := make(map[string]*StorageBucketGrantModel)

	for _, grant := range s3Grants {
		if grant.Grantee == nil || grant.Permission == nil {
			continue
		}

		// Create unique key for grantee (Type + Id + Uri)
		granteeType := ""
		if grant.Grantee.Type != nil {
			granteeType = *grant.Grantee.Type
		}
		granteeID := ""
		if grant.Grantee.ID != nil {
			granteeID = *grant.Grantee.ID
		}
		granteeURI := ""
		if grant.Grantee.URI != nil {
			granteeURI = *grant.Grantee.URI
		}

		key := fmt.Sprintf("%s|%s|%s", granteeType, granteeID, granteeURI)

		// Get or create model for this grantee
		model, exists := granteeMap[key]
		if !exists {
			model = &StorageBucketGrantModel{
				Type: types.StringValue(granteeType),
			}
			if granteeID != "" {
				model.Id = types.StringValue(granteeID)
			} else {
				model.Id = types.StringNull()
			}
			if granteeURI != "" {
				model.Uri = types.StringValue(granteeURI)
			} else {
				model.Uri = types.StringNull()
			}
			model.Permissions = types.ListNull(types.StringType)
			granteeMap[key] = model
		}

		// Add permission to existing permissions
		var permissions []string
		if !model.Permissions.IsNull() {
			diags.Append(model.Permissions.ElementsAs(context.Background(), &permissions, false)...)
			if diags.HasError() {
				return nil, diags
			}
		}

		// Check if permission already exists to avoid duplicates
		permissionExists := slices.Contains(permissions, *grant.Permission)
		if !permissionExists {
			permissions = append(permissions, *grant.Permission)
		}

		permList, d := types.ListValueFrom(context.Background(), types.StringType, permissions)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		model.Permissions = permList
	}

	modelGrants := make([]StorageBucketGrantModel, 0, len(granteeMap))
	for _, model := range granteeMap {
		modelGrants = append(modelGrants, *model)
	}

	return modelGrants, diags
}
