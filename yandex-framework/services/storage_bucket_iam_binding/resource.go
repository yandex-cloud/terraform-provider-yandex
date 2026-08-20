package storage_bucket_iam_binding

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	storage "github.com/yandex-cloud/go-genproto/yandex/cloud/storage/v1"
	storagesdk "github.com/yandex-cloud/go-sdk/services/storage/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/accessbinding"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
)

type bucketGetter interface {
	Get(context.Context, *storage.GetBucketRequest, ...grpc.CallOption) (*storage.Bucket, error)
}

type BucketIAMUpdater struct {
	Bucket         string
	ResourceId     string
	ProviderConfig *provider_config.Config
}

func NewIamBinding() resource.Resource {
	return accessbinding.NewIamBinding(newBucketIamUpdater())
}

func newBucketIamUpdater() accessbinding.ResourceIamUpdater {
	return &BucketIAMUpdater{}
}

func (u *BucketIAMUpdater) GetNameSuffix() string {
	return "storage_bucket_iam_binding"
}

func (u *BucketIAMUpdater) GetSchemaAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		u.GetIdAlias(): schema.StringAttribute{
			MarkdownDescription: "The name of the Object Storage (S3) bucket to attach the policy to. This resource should be used for managing [Service roles](https://yandex.cloud/docs/storage/security/#service-roles) only.",
			Required:            true,
		},
	}
}

func (u *BucketIAMUpdater) GetIdAlias() string {
	return "bucket"
}

func (u *BucketIAMUpdater) GetId() string {
	return u.Bucket
}

func (u *BucketIAMUpdater) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. "+
				"Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	u.ProviderConfig = providerConfig
}

func (u *BucketIAMUpdater) Initialize(ctx context.Context, state accessbinding.Extractable, diag *diag.Diagnostics) {
	var id types.String
	diag.Append(state.GetAttribute(ctx, path.Root("bucket"), &id)...)
	u.Bucket = id.ValueString()
}

// Access bindings attach to Bucket.resource_id (the owning cloud entity), not to the bucket
// name, and only the API knows the mapping.
func (u *BucketIAMUpdater) resourceID(ctx context.Context) (string, error) {
	if u.ResourceId != "" {
		return u.ResourceId, nil
	}

	resourceID, err := resolveBucketResourceID(ctx, storagesdk.NewBucketClient(u.ProviderConfig.SDKv2), u.Bucket)
	if err != nil {
		return "", err
	}
	u.ResourceId = resourceID

	return u.ResourceId, nil
}

func resolveBucketResourceID(ctx context.Context, client bucketGetter, bucketName string) (string, error) {
	bucket, err := client.Get(ctx, &storage.GetBucketRequest{
		Name: bucketName,
		View: storage.GetBucketRequest_VIEW_BASIC,
	})
	if err != nil {
		return "", err
	}
	return bucket.GetResourceId(), nil
}

func (u *BucketIAMUpdater) GetResourceIamPolicy(ctx context.Context) (*accessbinding.Policy, error) {
	id, err := u.resourceID(ctx)
	if err != nil {
		return nil, err
	}

	bindings, err := u.GetAccessBindings(ctx, id)
	if err != nil {
		return nil, err
	}
	return &accessbinding.Policy{Bindings: bindings}, nil
}

func (u *BucketIAMUpdater) SetResourceIamPolicy(ctx context.Context, policy *accessbinding.Policy) error {
	id, err := u.resourceID(ctx)
	if err != nil {
		return err
	}

	req := &access.SetAccessBindingsRequest{
		ResourceId:     id,
		AccessBindings: policy.Bindings,
	}

	ctx, cancel := context.WithTimeout(ctx, provider_config.DefaultTimeout)
	defer cancel()

	op, err := storagesdk.NewBucketClient(u.ProviderConfig.SDKv2).SetAccessBindings(ctx, req)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("error setting access bindings of %s: %w", u.DescribeResource(), err)
	}

	return nil
}

func (u *BucketIAMUpdater) UpdateResourceIamPolicy(ctx context.Context, policy *accessbinding.PolicyDelta) error {
	id, err := u.resourceID(ctx)
	if err != nil {
		return err
	}

	bSize := 1000
	deltas := policy.Deltas
	dLen := len(deltas)

	for i := 0; i < accessbinding.CountBatches(dLen, bSize); i++ {
		req := &access.UpdateAccessBindingsRequest{
			ResourceId:          id,
			AccessBindingDeltas: deltas[i*bSize : min((i+1)*bSize, dLen)],
		}
		op, err := storagesdk.NewBucketClient(u.ProviderConfig.SDKv2).UpdateAccessBindings(ctx, req)
		if err != nil {
			return fmt.Errorf("error updating access bindings of %s: %w", u.DescribeResource(), err)
		}

		_, err = op.Wait(ctx)
		if err != nil {
			return fmt.Errorf("error updating access bindings of %s: %w", u.DescribeResource(), err)
		}
	}

	return nil
}

func (u *BucketIAMUpdater) GetMutexKey() string {
	return fmt.Sprintf("iam-storage-bucket-%s", u.Bucket)
}

func (u *BucketIAMUpdater) DescribeResource() string {
	return fmt.Sprintf("storage-bucket '%s'", u.Bucket)
}

func (u *BucketIAMUpdater) GetAccessBindings(ctx context.Context, id string) ([]*access.AccessBinding, error) {
	var bindings []*access.AccessBinding
	pageToken := ""

	for {
		resp, err := storagesdk.NewBucketClient(u.ProviderConfig.SDKv2).ListAccessBindings(ctx, &access.ListAccessBindingsRequest{
			ResourceId: id,
			PageSize:   accessbinding.DefaultPageSize,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, fmt.Errorf("error retrieving access bindings of bucket %s: %w", id, err)
		}

		bindings = append(bindings, resp.AccessBindings...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	return bindings, nil
}
