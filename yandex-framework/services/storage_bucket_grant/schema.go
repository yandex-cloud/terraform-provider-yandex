package storage_bucket_grant

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	storage "github.com/yandex-cloud/terraform-provider-yandex/pkg/storage/s3"
)

var bucketACLAllowedValues = []string{
	storage.BucketOwnerFullControl,
	storage.BucketCannedACLPublicRead,
	storage.BucketCannedACLPublicReadWrite,
	storage.BucketCannedACLAuthenticatedRead,
	storage.BucketCannedACLPrivate,
}

func ResourceSchema(ctx context.Context) schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Allows management of grants of an existing [Yandex Cloud Storage Bucket](https://yandex.cloud/docs/storage/concepts/bucket).\n\n~> By default, for authentication, you need to use [IAM token](https://yandex.cloud/docs/iam/concepts/authorization/iam-token) with the necessary permissions.\n\n~> Alternatively, you can provide [static access keys](https://yandex.cloud/docs/iam/concepts/authorization/access-key) (Access and Secret). To generate these keys, you will need a Service Account with the appropriate permissions.\n\nThis resource *must not* be used in conjunction with `yandex_storage_bucket` or they will conflict over what your policy should be.\n\nThis resource should be used for managing [Primitive roles](https://yandex.cloud/en-ru/docs/storage/security/#primitive-roles) only.\n",
		Attributes: map[string]schema.Attribute{
			"bucket": schema.StringAttribute{
				MarkdownDescription: "The name of the bucket.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "The access key to use when applying changes. This value can also be provided as `storage_access_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
				Optional:            true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "The secret key to use when applying changes. This value can also be provided as `storage_secret_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
				Optional:            true,
				Sensitive:           true,
			},
			"acl": schema.StringAttribute{
				MarkdownDescription: "The [predefined ACL](https://yandex.cloud/docs/storage/concepts/acl#predefined_acls) to apply. Defaults to `private`. Conflicts with `grant`.\n\n~> To change ACL after creation, service account with `storage.admin` role should be used, though this role is not necessary to create a bucket with any ACL.\n",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(bucketACLAllowedValues...),
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("grant")),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"grant": schema.ListNestedBlock{
				MarkdownDescription: "An [ACL policy grant](https://yandex.cloud/docs/storage/concepts/acl#permissions-types). Conflicts with `acl`.\nAll permissions for a single grantee must be specified in a single `grant` block.\n\n~> To manage `grant` argument, service account with `storage.admin` role should be used.\n",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Canonical user id to grant for. Used only when type is `CanonicalUser`.",
							Optional:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Type of grantee to apply for. Valid values are `CanonicalUser` and `Group`.",
							Required:            true,
							Validators: []validator.String{
								stringvalidator.OneOf(storage.TypeCanonicalUser, storage.TypeGroup),
							},
						},
						"uri": schema.StringAttribute{
							MarkdownDescription: "URI address to grant for. Used only when type is Group.",
							Optional:            true,
						},
						"permissions": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "List of permissions to apply for grantee. Valid values are `READ`, `WRITE`, `FULL_CONTROL`.",
							Required:            true,
							Validators: []validator.List{
								listvalidator.ValueStringsAre(
									stringvalidator.OneOf(
										storage.PermissionFullControl,
										storage.PermissionRead,
										storage.PermissionWrite,
									),
								),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("acl")),
				},
			},
		},
	}
}
