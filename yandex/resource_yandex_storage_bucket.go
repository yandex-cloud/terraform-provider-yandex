package yandex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/storage/s3"

	awspolicy "github.com/jen20/awspolicyequivalence"
	storagepb "github.com/yandex-cloud/go-genproto/yandex/cloud/storage/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var storageClassSet = []string{
	s3.StorageClassStandard,
	s3.StorageClassCold,
	s3.StorageClassIce,
}

func resourceYandexStorageBucket() *schema.Resource {
	return &schema.Resource{
		Description:   "Allows management of [Yandex Cloud Storage Bucket](https://yandex.cloud/docs/storage/concepts/bucket).\n\n~> By default, for authentication, you need to use [IAM token](https://yandex.cloud/docs/iam/concepts/authorization/iam-token) with the necessary permissions.\n\n~> Alternatively, you can provide [static access keys](https://yandex.cloud/docs/iam/concepts/authorization/access-key) (Access and Secret). To generate these keys, you will need a Service Account with the appropriate permissions.\n\n~> For extended API usage, such as setting the `max_size`, `folder_id`, `anonymous_access_flags`, `default_storage_class`, and `https` parameters for a bucket, only the default authorization method will be used. This means the `IAM` token from the `provider` block will be applied.\nThis can be confusing in cases where a separate service account is used for managing buckets because, in such scenarios,buckets may be accessed by two different accounts, each with potentially different permissions for the buckets.\n\n~> In case you are using IAM token from UserAccount, you are needed to explicitly specify `folder_id` in the resource, as it cannot be identified from such type of account. In case you are using IAM token from ServiceAccount or static access keys, `folder_id` does not need to be specified unless you want to create the resource in a different folder than the account folder.\n\n~> Terraform will import this resource with `force_destroy` set to `false` in state. If you've set it to `true` in config, run `terraform apply` to update the value set in state. If you delete this resource before updating the value, objects in the bucket will not be destroyed.\n",
		CreateContext: resourceYandexStorageBucketCreate,
		ReadContext:   resourceYandexStorageBucketRead,
		UpdateContext: resourceYandexStorageBucketUpdate,
		DeleteContext: resourceYandexStorageBucketDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceYandexStorageBucketV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceYandexStorageBucketStateUpgradeV0,
				Version: 0,
			},
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:          schema.TypeString,
				Description:   "The name of the bucket. If omitted, Terraform will assign a random, unique name.",
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket_prefix"},
			},
			"bucket_prefix": {
				Type:          schema.TypeString,
				Description:   "Creates a unique bucket name beginning with the specified prefix. Conflicts with `bucket`.",
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket"},
			},
			"bucket_domain_name": {
				Type:        schema.TypeString,
				Description: "The bucket domain name.",
				Computed:    true,
			},

			"access_key": {
				Type:        schema.TypeString,
				Description: "The access key to use when applying changes. This value can also be provided as `storage_access_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
				Optional:    true,
				Sensitive:   true,
			},

			"secret_key": {
				Type:        schema.TypeString,
				Description: "The secret key to use when applying changes. This value can also be provided as `storage_secret_key` specified in provider config (explicitly or within `shared_credentials_file`) is used.",
				Optional:    true,
				Sensitive:   true,
			},

			"acl": {
				Type:          schema.TypeString,
				Description:   "The [predefined ACL](https://yandex.cloud/docs/storage/concepts/acl#predefined_acls) to apply. Defaults to `private`. Conflicts with `grant`.\n\n~> To change ACL after creation, service account with `storage.admin` role should be used, though this role is not necessary to create a bucket with any ACL.\n",
				Optional:      true,
				Deprecated:    "Use `yandex_storage_bucket_grant` instead.",
				Computed:      true,
				ConflictsWith: []string{"grant"},
				ValidateFunc:  validation.StringInSlice(bucketACLAllowedValues, false),
			},

			"grant": {
				Type:          schema.TypeSet,
				Description:   "An [ACL policy grant](https://yandex.cloud/docs/storage/concepts/acl#permissions-types). Conflicts with `acl`.\n\n~> To manage `grant` argument, service account with `storage.admin` role should be used.\n",
				Optional:      true,
				Deprecated:    "Use `yandex_storage_bucket_grant` instead.",
				Computed:      true,
				Set:           grantHash,
				ConflictsWith: []string{"acl"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Canonical user id to grant for. Used only when type is `CanonicalUser`.",
							Optional:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Type of grantee to apply for. Valid values are `CanonicalUser` and `Group`.",
							Required:    true,
							ValidateFunc: validation.StringInSlice([]string{
								s3.TypeCanonicalUser,
								s3.TypeGroup,
							}, false),
						},
						"uri": {
							Type:        schema.TypeString,
							Description: "URI address to grant for. Used only when type is Group.",
							Optional:    true,
						},

						"permissions": {
							Type:        schema.TypeSet,
							Description: "List of permissions to apply for grantee. Valid values are `READ`, `WRITE`, `FULL_CONTROL`.",
							Required:    true,
							Set:         schema.HashString,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									s3.PermissionFullControl,
									s3.PermissionRead,
									s3.PermissionWrite,
								}, false),
							},
						},
					},
				},
			},

			"policy": {
				Type:             schema.TypeString,
				Description:      "The `policy` object should contain the only field with the text of the policy. See [policy documentation](https://yandex.cloud/docs/storage/concepts/policy) for more information on policy format.",
				Optional:         true,
				ValidateFunc:     validateStringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
				Deprecated:       "Use `yandex_storage_bucket_policy` resource instead.",
				Computed:         true,
			},

			"cors_rule": {
				Type:        schema.TypeList,
				Description: "A rule of [Cross-Origin Resource Sharing](https://yandex.cloud/docs/storage/concepts/cors) (CORS object).",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:        schema.TypeList,
							Description: "Specifies which headers are allowed.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:        schema.TypeList,
							Description: "Specifies which methods are allowed. Can be `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.",
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"allowed_origins": {
							Type:        schema.TypeList,
							Description: "Specifies which origins are allowed.",
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:        schema.TypeList,
							Description: "Specifies expose header in the response.",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"max_age_seconds": {
							Type:        schema.TypeInt,
							Description: "Specifies time in seconds that browser can cache the response for a preflight request.",
							Optional:    true,
						},
					},
				},
			},

			"website": {
				Type:        schema.TypeList,
				Description: "A [Website Object](https://yandex.cloud/docs/storage/concepts/hosting)",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_document": {
							Type:        schema.TypeString,
							Description: "Storage returns this index document when requests are made to the root domain or any of the subfolders (unless using `redirect_all_requests_to`).",
							Optional:    true,
						},

						"error_document": {
							Type:        schema.TypeString,
							Description: "An absolute path to the document to return in case of a 4XX error.",
							Optional:    true,
						},

						"redirect_all_requests_to": {
							Description: "A hostname to redirect all website requests for this bucket to. Hostname can optionally be prefixed with a protocol (`http://` or `https://`) to use when redirecting requests. The default is the protocol that is used in the original request.",
							Type:        schema.TypeString,
							ConflictsWith: []string{
								"website.0.index_document",
								"website.0.error_document",
								"website.0.routing_rules",
							},
							Optional: true,
						},

						"routing_rules": {
							Type:         schema.TypeString,
							Description:  "A JSON array containing [routing rules](https://yandex.cloud/docs/storage/s3/api-ref/hosting/upload#request-scheme) describing redirect behavior and when redirects are applied.",
							Optional:     true,
							ValidateFunc: validateStringIsJSON,
							StateFunc: func(v interface{}) string {
								json, _ := NormalizeJsonString(v)
								return json
							},
						},
					},
				},
			},
			"website_endpoint": {
				Type:        schema.TypeString,
				Description: "The website endpoint, if the bucket is configured with a website. If not, this will be an empty string.",
				Optional:    true,
				Computed:    true,
			},
			"website_domain": {
				Type:        schema.TypeString,
				Description: "The domain of the website endpoint, if the bucket is configured with a website. If not, this will be an empty string.",
				Optional:    true,
				Computed:    true,
			},

			"versioning": {
				Type:        schema.TypeList,
				Description: "A state of [versioning](https://yandex.cloud/docs/storage/concepts/versioning).\n\n~> To manage `versioning` argument, service account with `storage.admin` role should be used.\n",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enable versioning. Once you version-enable a bucket, it can never return to an unversioned state. You can, however, suspend versioning on that bucket.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},

			"object_lock_configuration": {
				Type:        schema.TypeList,
				Description: "A configuration of [object lock management](https://yandex.cloud/docs/storage/concepts/object-lock).",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"object_lock_enabled": {
							Type:         schema.TypeString,
							Description:  "Enable object locking in a bucket. Require versioning to be enabled.",
							Optional:     true,
							Default:      s3.ObjectLockEnabled,
							ValidateFunc: validation.StringInSlice(s3.ObjectLockEnabledValues, false),
						},
						"rule": {
							Type:        schema.TypeList,
							Description: "Specifies a default locking configuration for added objects. Require object_lock_enabled to be enabled.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"default_retention": {
										Description: "Default retention object.",
										Type:        schema.TypeList,
										Required:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mode": {
													Type:        schema.TypeString,
													Description: "Specifies a type of object lock. One of `[\"GOVERNANCE\", \"COMPLIANCE\"]`.",
													Required:    true,
													ValidateFunc: validation.StringInSlice(
														s3.ObjectLockRetentionModeValues,
														false,
													),
												},
												"days": {
													Type:        schema.TypeInt,
													Description: "Specifies a retention period in days after uploading an object version. It must be a positive integer. You can't set it simultaneously with `years`.",
													Optional:    true,
													ExactlyOneOf: []string{
														"object_lock_configuration.0.rule.0.default_retention.0.days",
														"object_lock_configuration.0.rule.0.default_retention.0.years",
													},
												},
												"years": {
													Type:        schema.TypeInt,
													Description: "Specifies a retention period in years after uploading an object version. It must be a positive integer. You can't set it simultaneously with `days`.",
													Optional:    true,
													ExactlyOneOf: []string{
														"object_lock_configuration.0.rule.0.default_retention.0.days",
														"object_lock_configuration.0.rule.0.default_retention.0.years",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			"logging": {
				Type:        schema.TypeSet,
				Description: "A settings of [bucket logging](https://yandex.cloud/docs/storage/concepts/server-logs).",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_bucket": {
							Type:        schema.TypeString,
							Description: "The name of the bucket that will receive the log objects.",
							Required:    true,
						},
						"target_prefix": {
							Type:        schema.TypeString,
							Description: "To specify a key prefix for log objects.",
							Optional:    true,
						},
					},
				},
				Set: func(v interface{}) int {
					var buf bytes.Buffer
					m := v.(map[string]interface{})
					buf.WriteString(fmt.Sprintf("%s-", m["target_bucket"]))
					buf.WriteString(fmt.Sprintf("%s-", m["target_prefix"]))
					return hashcode.String(buf.String())
				},
			},

			"lifecycle_rule": {
				Type:        schema.TypeList,
				Description: "A configuration of [object lifecycle management](https://yandex.cloud/docs/storage/concepts/lifecycles).",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Description:  "Unique identifier for the rule. Must be less than or equal to 255 characters in length.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"prefix": {
							Type:             schema.TypeString,
							Description:      "Object key prefix identifying one or more objects to which the rule applies.",
							Optional:         true,
							Deprecated:       "Use filter instead",
							DiffSuppressFunc: suppressPrefixDiffIfFilterPrefixSet,
						},
						"filter": {
							Type:             schema.TypeList,
							Description:      "Filter block identifies one or more objects to which the rule applies. A Filter must have exactly one of Prefix, Tag, or And specified. The filter supports options listed below.\n\nAt least one of `abort_incomplete_multipart_upload_days`, `expiration`, `transition`, `noncurrent_version_expiration`, `noncurrent_version_transition` must be specified.",
							Optional:         true,
							MaxItems:         1,
							DiffSuppressFunc: suppressFilterIfPrefixEqualsFilterPrefix,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"and": {
										Type:        schema.TypeList,
										Description: "A logical `and` operator applied to one or more filter parameters. It should be used when two or more of the above parameters are used.",
										Optional:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object_size_greater_than": {
													Type:         schema.TypeInt,
													Description:  "Minimum object size to which the rule applies.",
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(0),
												},
												"object_size_less_than": {
													Type:         schema.TypeInt,
													Description:  "Maximum object size to which the rule applies.",
													Optional:     true,
													ValidateFunc: validation.IntAtLeast(1),
												},
												"prefix": {
													Type:        schema.TypeString,
													Description: "Object key prefix identifying one or more objects to which the rule applies.",
													Optional:    true,
												},
												"tags": tagsSchema(),
											},
										},
									},
									"object_size_greater_than": {
										Type:        schema.TypeInt,
										Description: "Minimum object size to which the rule applies.",
										Optional:    true,
									},
									"object_size_less_than": {
										Type:        schema.TypeInt,
										Description: "Maximum object size to which the rule applies.",
										Optional:    true,
									},
									"prefix": {
										Type:             schema.TypeString,
										Description:      "Object key prefix identifying one or more objects to which the rule applies.",
										Optional:         true,
										DiffSuppressFunc: suppressFilterPrefixDiffIfPrefixSet,
									},
									"tag": {
										Type:        schema.TypeList,
										Description: "A key and value pair for filtering objects. E.g.: `key=key1, value=value1`.",
										MaxItems:    1,
										Optional:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key": {
													Description: "A key.",
													Type:        schema.TypeString,
													Required:    true,
												},
												"value": {
													Description: "A value.",
													Type:        schema.TypeString,
													Required:    true,
												},
											},
										},
									},
								},
							},
						},
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Specifies lifecycle rule status.",
							Required:    true,
						},
						"abort_incomplete_multipart_upload_days": {
							Type:        schema.TypeInt,
							Description: "Specifies the number of days after initiating a multipart upload when the multipart upload must be completed.",
							Optional:    true,
						},
						"expiration": {
							Type:        schema.TypeList,
							Description: "Specifies a period in the object's expire.",
							Optional:    true,
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Description:  "Specifies the date after which you want the corresponding action to take effect.",
										Optional:     true,
										ValidateFunc: validateS3BucketLifecycleTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Description:  "Specifies the number of days after object creation when the specific rule action takes effect.",
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"expired_object_delete_marker": {
										Type:        schema.TypeBool,
										Description: "n a versioned bucket (versioning-enabled or versioning-suspended bucket), you can add this element in the lifecycle configuration to direct Object Storage to delete expired object delete markers.",
										Optional:    true,
									},
								},
							},
						},
						"noncurrent_version_expiration": {
							Type:        schema.TypeList,
							Description: "Specifies when noncurrent object versions expire.",
							MaxItems:    1,
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:         schema.TypeInt,
										Description:  "Specifies the number of days noncurrent object versions expire.",
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"transition": {
							Type:        schema.TypeSet,
							Description: "Specifies a period in the object's transitions.",
							Optional:    true,
							Set:         s3.TransitionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Description:  "Specifies the date after which you want the corresponding action to take effect.",
										Optional:     true,
										ValidateFunc: validateS3BucketLifecycleTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Description:  "Specifies the number of days after object creation when the specific rule action takes effect.",
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:         schema.TypeString,
										Description:  "Specifies the storage class to which you want the object to transition. Supported values: [`STANDARD_IA`, `COLD`, `ICE`].",
										Required:     true,
										ValidateFunc: validation.StringInSlice(storageClassSet, false),
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:        schema.TypeSet,
							Description: "Specifies when noncurrent object versions transitions.",
							Optional:    true,
							Set:         s3.TransitionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:         schema.TypeInt,
										Description:  "Specifies the number of days noncurrent object versions transition.",
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:         schema.TypeString,
										Description:  "Specifies the storage class to which you want the noncurrent object versions to transition. Supported values: [`STANDARD_IA`, `COLD`, `ICE`].",
										Required:     true,
										ValidateFunc: validation.StringInSlice(storageClassSet, false),
									},
								},
							},
						},
					},
				},
			},

			"force_destroy": {
				Type:        schema.TypeBool,
				Description: " A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error. These objects are *not* recoverable. Default is `false`.",
				Optional:    true,
				Default:     false,
			},

			"server_side_encryption_configuration": {
				Type:        schema.TypeList,
				Description: "A configuration of server-side encryption for the bucket.",
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule": {
							Type:        schema.TypeList,
							Description: "A single object for server-side encryption by default configuration.",
							MaxItems:    1,
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"apply_server_side_encryption_by_default": {
										Type:        schema.TypeList,
										Description: "A single object for setting server-side encryption by default.",
										MaxItems:    1,
										Required:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"kms_master_key_id": {
													Type:        schema.TypeString,
													Description: "The KMS master key ID used for the SSE-KMS encryption.",
													Required:    true,
												},
												"sse_algorithm": {
													Type:        schema.TypeString,
													Description: "The server-side encryption algorithm to use. Single valid value is `aws:kms`.",
													Required:    true,
													ValidateFunc: validation.StringInSlice([]string{
														s3.ServerSideEncryptionAwsKms,
													}, false),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},

			// These fields use extended API and requires IAM token
			// to be set in order to operate.
			"default_storage_class": {
				Type:        schema.TypeString,
				Description: "Storage class which is used for storing objects by default. Available values are: \"STANDARD\", \"COLD\", \"ICE\". Default is `\"STANDARD\"`. See [Storage Class](https://yandex.cloud/docs/storage/concepts/storage-class) for more information.",
				Optional:    true,
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: "Allow to create bucket in different folder. In case you are using IAM token from UserAccount, you are needed to explicitly specify folder_id in the resource, as it cannot be identified from such type of account. In case you are using IAM token from ServiceAccount or static access keys, folder_id does not need to be specified unless you want to create the resource in a different folder than the account folder.\n\n~> It will try to create bucket using `IAM-token`, not using `access keys`.\n",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"max_size": {
				Type:        schema.TypeInt,
				Description: "The size of bucket, in bytes. See [Size Limiting](https://yandex.cloud/docs/storage/operations/buckets/limit-max-volume) for more information.",
				Optional:    true,
			},

			"anonymous_access_flags": {
				Type:        schema.TypeSet,
				Description: "Provides various access to objects. See [Bucket Availability](https://yandex.cloud/docs/storage/operations/buckets/bucket-availability) for more information.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Set:         storageBucketS3SetFunc("list", "read", "config_read"),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"list": {
							Type:        schema.TypeBool,
							Description: "Allows to list object in bucket anonymously.",
							Optional:    true,
						},
						"read": {
							Type:        schema.TypeBool,
							Description: "Allows to read objects in bucket anonymously.",
							Optional:    true,
						},
						"config_read": {
							Type:        schema.TypeBool,
							Description: "Allows to read bucket configuration anonymously.",
							Optional:    true,
						},
					},
				},
			},

			"https": {
				Type:        schema.TypeSet,
				Description: "Manages https certificates for bucket. See [https](https://yandex.cloud/docs/storage/operations/hosting/certificate) for more information.",
				Optional:    true,
				MaxItems:    1,
				Set:         storageBucketS3SetFunc("certificate_id"),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_id": {
							Type:        schema.TypeString,
							Description: "Id of the certificate in Certificate Manager, that will be used for bucket.",
							Required:    true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func tagsSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Description: "The `tags` object for setting tags (or labels) for bucket. See [Tags](https://yandex.cloud/docs/storage/concepts/tags) for more information.",
		Optional:    true,
		Elem: &schema.Schema{
			Type:        schema.TypeString,
			Description: "A tag.",
		},
	}
}

var bucketACLAllowedValues = []string{
	string(s3.BucketACLOwnerFullControl),
	string(s3.BucketACLPublicRead),
	string(s3.BucketACLPublicReadWrite),
	string(s3.BucketACLAuthRead),
	string(s3.BucketACLPrivate),
}

func resourceYandexStorageBucketCreateBySDK(d *schema.ResourceData, meta interface{}) error {
	mapACL := func(acl s3.BucketACL) (*storagepb.ACL, error) {
		baseACL := &storagepb.ACL{}
		switch acl {
		case s3.BucketACLPublicRead:
			baseACL.Grants = []*storagepb.ACL_Grant{{
				Permission: storagepb.ACL_Grant_PERMISSION_READ,
				GrantType:  storagepb.ACL_Grant_GRANT_TYPE_ALL_USERS,
			}}
		case s3.BucketACLPublicReadWrite:
			baseACL.Grants = []*storagepb.ACL_Grant{{
				Permission: storagepb.ACL_Grant_PERMISSION_READ,
				GrantType:  storagepb.ACL_Grant_GRANT_TYPE_ALL_USERS,
			}, {
				Permission: storagepb.ACL_Grant_PERMISSION_READ,
				GrantType:  storagepb.ACL_Grant_GRANT_TYPE_ALL_USERS,
			}}
		case s3.BucketACLAuthRead:
			baseACL.Grants = []*storagepb.ACL_Grant{{
				Permission: storagepb.ACL_Grant_PERMISSION_READ,
				GrantType:  storagepb.ACL_Grant_GRANT_TYPE_ALL_AUTHENTICATED_USERS,
			}}
		case s3.BucketACLPrivate, s3.BucketACLOwnerFullControl:
			baseACL.Grants = []*storagepb.ACL_Grant{}
		}

		return baseACL, nil
	}

	bucket := d.Get("bucket").(string)
	folderID := d.Get("folder_id").(string)
	acl := s3.BucketACL(d.Get("acl").(string))
	maxSize := d.Get("max_size").(int)
	defaultStorageClass := d.Get("default_storage_class").(string)
	aaf := getAnonymousAccessFlagsSDK(d.Get("anonymous_access_flags"))

	request := &storagepb.CreateBucketRequest{
		Name:                 bucket,
		FolderId:             folderID,
		AnonymousAccessFlags: aaf,
		MaxSize:              int64(maxSize),
		DefaultStorageClass:  defaultStorageClass,
	}

	var err error
	request.Acl, err = mapACL(acl)
	if err != nil {
		return fmt.Errorf("mapping acl: %w", err)
	}

	config := meta.(*Config)
	ctx := config.Context()

	log.Printf("[INFO] Creating Storage S3 bucket using sdk: %s", protojson.Format(request))

	bucketAPI := config.sdk.StorageAPI().Bucket()
	op, err := bucketAPI.Create(ctx, request)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		log.Printf("[ERROR] Unable to create S3 bucket using sdk: %v", err)

		return err
	}

	responseBucket := &storagepb.Bucket{}
	err = op.GetResponse().UnmarshalTo(responseBucket)
	if err != nil {
		log.Printf("[ERROR] Returned message is not a bucket: %v", err)

		return err
	}

	log.Printf("[INFO] Created Storage S3 bucket: %s", protojson.Format(responseBucket))

	return nil
}

func resourceYandexStorageBucketCreateByS3Client(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	bucket := d.Get("bucket").(string)
	var acl s3.BucketACL
	if aclValue, ok := d.GetOk("acl"); ok {
		acl = s3.BucketACL(aclValue.(string))
	} else {
		acl = s3.BucketACLPrivate
	}

	config := meta.(*Config)

	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	log.Printf("[INFO] Trying to create new Storage S3 Bucket: %q, ACL: %q", bucket, acl)
	if err := s3Client.CreateBucket(ctx, bucket, acl); err != nil {
		log.Printf("[ERROR] Got an error while trying to create Storage Bucket %s: %s", bucket, err)
		return err
	}

	log.Printf("[INFO] Created new Storage S3 Bucket: %q, ACL: %q", bucket, acl)
	return nil
}

func resourceYandexStorageBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Get the bucket and acl
	var bucket string
	if v, ok := d.GetOk("bucket"); ok {
		bucket = v.(string)
	} else if v, ok := d.GetOk("bucket_prefix"); ok {
		bucket = resource.PrefixedUniqueId(v.(string))
	} else {
		bucket = resource.UniqueId()
	}

	if err := validateS3BucketName(bucket); err != nil {
		return diag.Errorf("error validating Storage Bucket name: %s", err)
	}

	d.Set("bucket", bucket)

	var err error
	if folderID, ok := d.Get("folder_id").(string); ok && folderID != "" {
		err = resourceYandexStorageBucketCreateBySDK(d, meta)
	} else {
		err = resourceYandexStorageBucketCreateByS3Client(ctx, d, meta)
	}
	if err != nil {
		return diag.Errorf("error creating Storage S3 Bucket: %s", err)
	}

	d.SetId(bucket)

	return resourceYandexStorageBucketUpdate(ctx, d, meta)
}

func resourceYandexStorageRequireExternalSDK(d *schema.ResourceData) bool {
	value, ok := d.GetOk("folder_id")
	if !ok {
		return false
	}

	folderID, ok := value.(string)
	if !ok {
		return false
	}

	return folderID != ""
}

func resourceYandexStorageBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := resourceYandexStorageBucketUpdateBasic(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceYandexStorageBucketUpdateExtended(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceYandexStorageBucketRead(ctx, d, meta)
}

func resourceYandexStorageBucketUpdateBasic(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	type property struct {
		name          string
		updateHandler func(context.Context, *s3.Client, *schema.ResourceData) error
	}
	resourceProperties := []property{
		{"policy", resourceYandexStorageBucketPolicyUpdate},
		{"cors_rule", resourceYandexStorageBucketCORSUpdate},
		{"website", resourceYandexStorageBucketWebsiteUpdate},
		{"versioning", resourceYandexStorageBucketVersioningUpdate},
		{"acl", resourceYandexStorageBucketACLUpdate},
		{"grant", resourceYandexStorageBucketGrantsUpdate},
		{"logging", resourceYandexStorageBucketLoggingUpdate},
		{"lifecycle_rule", resourceYandexStorageBucketLifecycleUpdate},
		{"server_side_encryption_configuration", resourceYandexStorageBucketServerSideEncryptionConfigurationUpdate},
		{"object_lock_configuration", resourceYandexStorageBucketObjectLockConfigurationUpdate},
		{"tags", resourceYandexStorageBucketTagsUpdate},
	}

	for _, property := range resourceProperties {
		if !d.HasChange(property.name) {
			continue
		}

		if property.name == "acl" && d.IsNewResource() {
			continue
		}

		err := property.updateHandler(ctx, s3Client, d)
		if err != nil {
			return fmt.Errorf("handling %s: %w", property.name, err)
		}
	}

	return nil
}

func resourceYandexStorageBucketUpdateExtended(d *schema.ResourceData, meta interface{}) (err error) {
	if d.Id() == "" {
		// bucket has been deleted, skipping
		return nil
	}

	bucket := d.Get("bucket").(string)
	bucketUpdateRequest := &storagepb.UpdateBucketRequest{
		Name: bucket,
	}
	paths := make([]string, 0, 3)

	createdBySDK := resourceYandexStorageRequireExternalSDK(d)
	handleChange := func(property string, f func(value interface{})) {
		switch {
		// If this bucket is a new resource and we created it
		// by our SDK it means we've already set all parameters
		// to its values.
		case d.IsNewResource() && createdBySDK:
			fallthrough
		case !d.HasChange(property):
			return
		}

		log.Printf("[DEBUG] property %q is going to be updated", property)

		value := d.Get(property)
		f(value)

		paths = append(paths, property)
	}

	changeHandlers := map[string]func(value interface{}){
		"default_storage_class": func(value interface{}) {
			bucketUpdateRequest.SetDefaultStorageClass(value.(string))
		},
		"max_size": func(value interface{}) {
			bucketUpdateRequest.SetMaxSize(int64(value.(int)))
		},
		"anonymous_access_flags": func(value interface{}) {
			bucketUpdateRequest.AnonymousAccessFlags = getAnonymousAccessFlagsSDK(value)
		},
	}

	for field, handler := range changeHandlers {
		handleChange(field, handler)
	}

	config := meta.(*Config)
	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutUpdate))
	defer cancel()

	bucketAPI := config.sdk.StorageAPI().Bucket()

	if len(paths) > 0 {
		bucketUpdateRequest.UpdateMask, err = fieldmaskpb.New(bucketUpdateRequest, paths...)
		if err != nil {
			return fmt.Errorf("constructing field mask: %w", err)
		}

		log.Printf("[INFO] updating S3 bucket extended parameters: %s", protojson.Format(bucketUpdateRequest))

		op, err := bucketAPI.Update(ctx, bucketUpdateRequest)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			if handleBucketNotFoundError(d, err) {
				return nil
			}

			log.Printf("[WARN] Storage api error updating S3 bucket extended parameters: %v", err)

			return err
		}

		if opErr := op.GetError(); opErr != nil {
			log.Printf("[WARN] Operation ended with error: %s", protojson.Format(opErr))

			return status.Error(codes.Code(opErr.Code), opErr.Message)
		}

		log.Printf("[INFO] updated S3 bucket extended parameters: %s", protojson.Format(op.GetResponse()))
	}

	if !d.HasChange("https") {
		return nil
	}

	log.Println("[DEBUG] updating S3 bucket https configuration")

	schemaSet := d.Get("https").(*schema.Set)
	if schemaSet.Len() > 0 {
		httpsUpdateRequest := &storagepb.SetBucketHTTPSConfigRequest{
			Name: bucket,
		}

		params := schemaSet.List()[0].(map[string]interface{})
		httpsUpdateRequest.Params = &storagepb.SetBucketHTTPSConfigRequest_CertificateManager{
			CertificateManager: &storagepb.CertificateManagerHTTPSConfigParams{
				CertificateId: params["certificate_id"].(string),
			},
		}

		log.Printf("[INFO] updating S3 bucket https config: %s", protojson.Format(httpsUpdateRequest))
		op, err := bucketAPI.SetHTTPSConfig(ctx, httpsUpdateRequest)
		err = waitOperation(ctx, config, op, err)
		if err != nil {
			if handleBucketNotFoundError(d, err) {
				return nil
			}

			log.Printf("[WARN] Storage api updating S3 bucket https config: %v", err)

			return err
		}

		if opErr := op.GetError(); opErr != nil {
			log.Printf("[WARN] Operation ended with error: %s", protojson.Format(opErr))

			return status.Error(codes.Code(opErr.Code), opErr.Message)
		}

		log.Printf("[INFO] updated S3 bucket https config: %s", protojson.Format(op.GetResponse()))

		return nil
	}

	httpsDeleteRequest := &storagepb.DeleteBucketHTTPSConfigRequest{
		Name: bucket,
	}

	log.Printf("[INFO] deleting S3 bucket https config: %s", protojson.Format(httpsDeleteRequest))
	op, err := bucketAPI.DeleteHTTPSConfig(ctx, httpsDeleteRequest)
	err = waitOperation(ctx, config, op, err)
	if err != nil {
		if handleBucketNotFoundError(d, err) {
			return nil
		}

		log.Printf("[WARN] Storage api deleting S3 bucket https config: %v", err)

		return err
	}

	if opErr := op.GetError(); opErr != nil {
		log.Printf("[WARN] Operation ended with error: %s", protojson.Format(opErr))

		return status.Error(codes.Code(opErr.Code), opErr.Message)
	}
	log.Printf("[INFO] deleted S3 bucket https config: %s", protojson.Format(op.GetResponse()))

	return nil
}

func resourceYandexStorageBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	err := resourceYandexStorageBucketReadBasic(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	err = resourceYandexStorageBucketReadExtended(d, meta)
	if err != nil {
		log.Printf("[WARN] Got an error reading Storage Bucket's extended properties: %s", err)
	}

	return nil
}

func resourceYandexStorageBucketReadBasic(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	bucketName := d.Id()
	acl := d.Get("acl").(string)
	bucket, err := s3Client.GetBucket(ctx, bucketName, config.StorageEndpoint, acl)
	if err != nil {
		if errors.Is(err, s3.ErrBucketNotFound) {
			log.Printf("[WARN] Storage Bucket (%s) not found, error code (404)", bucketName)
			d.SetId("")
			return nil
		}
		log.Printf("[ERROR] Unable to read Storage Bucket (%s): %s", bucketName, err)
		return err
	}

	if _, ok := d.GetOk("bucket"); !ok {
		d.Set("bucket", bucketName)
	}
	d.Set("bucket_domain_name", bucket.DomainName)
	if err := d.Set("policy", bucket.Policy); err != nil {
		return fmt.Errorf("error setting policy: %w", err)
	}
	if err := d.Set("cors_rule", bucket.CORSRules); err != nil {
		return fmt.Errorf("error setting cors_rule: %w", err)
	}
	if bucket.Website != nil {
		if err := d.Set("website", bucket.Website.RawData); err != nil {
			return fmt.Errorf("error setting website: %w", err)
		}
		if err := d.Set("website_endpoint", bucket.Website.Endpoint); err != nil {
			return fmt.Errorf("error setting website_endpoint: %w", err)
		}
		if err := d.Set("website_domain", bucket.Website.Domain); err != nil {
			return fmt.Errorf("error setting website_domain: %w", err)
		}
	} else {
		if err := d.Set("website", nil); err != nil {
			return fmt.Errorf("error resetting website: %w", err)
		}
	}
	if bucket.Grants != nil {
		if err := d.Set("grant", schema.NewSet(grantHash, bucket.Grants)); err != nil {
			return fmt.Errorf("error setting Storage Bucket `grant` %w", err)
		}
	} else {
		if err := d.Set("grant", nil); err != nil {
			return fmt.Errorf("error resetting Storage Bucket `grant` %w", err)
		}
	}
	if err := d.Set("versioning", bucket.Versioning); err != nil {
		return fmt.Errorf("error setting versioning: %w", err)
	}
	if err := d.Set("object_lock_configuration", bucket.ObjectLock); err != nil {
		return fmt.Errorf("error setting object lock configuration: %w", err)
	}
	if err := d.Set("logging", bucket.Logging); err != nil {
		return fmt.Errorf("error setting logging: %w", err)
	}
	if err := d.Set("lifecycle_rule", bucket.Lifecycle); err != nil {
		return fmt.Errorf("error setting lifecycle_rule: %w", err)
	}
	if err := d.Set("server_side_encryption_configuration", bucket.Encryption); err != nil {
		return fmt.Errorf("error setting server_side_encryption_configuration: %w", err)
	}
	if err := d.Set("tags", s3.TagsToRaw(bucket.Tags)); err != nil {
		return fmt.Errorf("error setting S3 Bucket tags: %w", err)
	}

	return nil
}

func resourceYandexStorageBucketReadExtended(d *schema.ResourceData, meta interface{}) error {
	if d.Id() == "" {
		// bucket has been deleted, skipping read
		return nil
	}

	config := meta.(*Config)
	bucketAPI := config.sdk.StorageAPI().Bucket()

	name := d.Get("bucket").(string)

	ctx, cancel := config.ContextWithTimeout(d.Timeout(schema.TimeoutRead))
	defer cancel()

	log.Println("[DEBUG] Getting S3 bucket extended parameters")

	bucket, err := bucketAPI.Get(ctx, &storagepb.GetBucketRequest{
		Name: name,
		View: storagepb.GetBucketRequest_VIEW_FULL,
	})
	if err != nil {
		if handleBucketNotFoundError(d, err) {
			return nil
		}

		log.Printf("[WARN] Storage api getting S3 bucket extended parameters: %v", err)

		return err
	}

	log.Printf("[DEBUG] Bucket %s", protojson.Format(bucket))

	d.Set("default_storage_class", bucket.GetDefaultStorageClass())
	d.Set("folder_id", bucket.GetFolderId())
	d.Set("max_size", bucket.GetMaxSize())

	aafValue := make([]map[string]interface{}, 0)
	if aaf := bucket.AnonymousAccessFlags; aaf != nil {
		flatten := map[string]interface{}{}
		if value := aaf.List; value != nil {
			flatten["list"] = value.Value
		}
		if value := aaf.Read; value != nil {
			flatten["read"] = value.Value
		}
		if value := aaf.ConfigRead; value != nil {
			flatten["config_read"] = value.Value
		}

		aafValue = append(aafValue, flatten)
	}

	log.Printf("[DEBUG] setting anonymous access flags: %v", aafValue)
	if len(aafValue) == 0 {
		d.Set("anonymous_access_flags", nil)
	} else {
		d.Set("anonymous_access_flags", aafValue)
	}

	log.Println("[DEBUG] trying to get S3 bucket https config")

	https, err := bucketAPI.GetHTTPSConfig(ctx, &storagepb.GetBucketHTTPSConfigRequest{
		Name: name,
	})
	switch {
	case err == nil:
		// continue
	case isStatusWithCode(err, codes.NotFound),
		isStatusWithCode(err, codes.PermissionDenied):
		log.Printf("[INFO] Storage api got minor error getting S3 bucket https config %v", err)
		d.Set("https", nil)

		return nil
	default:
		log.Printf("[WARN] Storage api error getting S3 bucket https config %v", err)

		return err
	}

	log.Printf("[DEBUG] S3 bucket https config: %s", protojson.Format(https))

	if https.SourceType == storagepb.HTTPSConfig_SOURCE_TYPE_MANAGED_BY_CERTIFICATE_MANAGER {
		flatten := map[string]interface{}{
			"certificate_id": https.CertificateId,
		}

		result := []map[string]interface{}{flatten}

		err = d.Set("https", result)
		if err != nil {
			return fmt.Errorf("updating S3 bucket https config state: %w", err)
		}
	}

	return nil
}

func resourceYandexStorageBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)
	s3Client, err := getS3Client(ctx, d, config)
	if err != nil {
		return diag.Errorf("error getting storage client: %s", err)
	}

	bucket := d.Id()
	force := d.Get("force_destroy").(bool)

	log.Printf("[DEBUG] Storage Delete Bucket: %s", bucket)
	err = s3Client.DeleteBucket(ctx, bucket, force)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceYandexStorageBucketCORSUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	rawCors := d.Get("cors_rule").([]interface{})

	rules := s3.NewCORSRules(rawCors)
	return s3Client.UpdateBucketCORS(ctx, bucket, rules)
}

func resourceYandexStorageBucketWebsiteUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	rawWebsite := d.Get("website").([]interface{})
	bucket := d.Get("bucket").(string)

	website, err := s3.NewWebsite(rawWebsite)
	if err != nil {
		return fmt.Errorf("error parsing website configuration: %s", err)
	}
	if err := s3Client.UpdateBucketWebsite(ctx, bucket, website); err != nil {
		return err
	}
	if website == nil {
		// cleanup after site deletion
		d.Set("website_endpoint", "")
		d.Set("website_domain", "")
	}

	return nil
}

func resourceYandexStorageBucketACLUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	acl := s3.BucketACL(d.Get("acl").(string))
	if acl == "" {
		acl = s3.BucketACLPrivate
	}
	bucket := d.Get("bucket").(string)

	return s3Client.UpdateBucketACL(ctx, bucket, acl)
}

func resourceYandexStorageBucketVersioningUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	rawVersioning := d.Get("versioning").([]interface{})
	bucket := d.Get("bucket").(string)

	versioningStatus := s3.NewVersioningStatus(rawVersioning)
	return s3Client.UpdateBucketVersioning(ctx, bucket, versioningStatus)
}

func resourceYandexStorageBucketTagsUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	rawTags := d.Get("tags")

	tags := s3.NewTags(rawTags)

	if err := s3Client.UpdateBucketTags(ctx, bucket, tags); err != nil {
		log.Printf("[ERROR] Unable to update Storage S3 bucket tags: %s", err)
		return err
	}

	return nil
}

func resourceYandexStorageBucketObjectLockConfigurationUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	olc := d.Get("object_lock_configuration").([]interface{})
	bucket := d.Get("bucket").(string)

	lock := s3.NewObjectLock(olc)
	return s3Client.UpdateBucketObjectLock(ctx, bucket, lock)
}

func resourceYandexStorageBucketLoggingUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	rawLogging := d.Get("logging").(*schema.Set).List()
	bucket := d.Get("bucket").(string)

	loggingStatus := s3.NewLoggingStatus(rawLogging)
	return s3Client.UpdateBucketLogging(ctx, bucket, loggingStatus)
}

type S3Website struct {
	Endpoint, Domain string
}

func handleBucketNotFoundError(d *schema.ResourceData, err error) bool {
	if isStatusWithCode(err, codes.NotFound) {
		log.Printf("[WARN] Storage Bucket (%s) not found, error code (404)", d.Id())
		d.SetId("")
		return true
	}
	return false
}

func validateS3BucketName(value string) error {
	if len(value) > 63 {
		return fmt.Errorf("%q must contain 63 characters at most", value)
	}
	if len(value) < 3 {
		return fmt.Errorf("%q must contain at least 3 characters", value)
	}
	if !regexp.MustCompile(`^[0-9a-zA-Z-.]+$`).MatchString(value) {
		return fmt.Errorf("only alphanumeric characters, hyphens, and periods allowed in %q", value)
	}

	return nil
}

func grantHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["type"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["uri"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if p, ok := m["permissions"]; ok {
		buf.WriteString(fmt.Sprintf("%v-", p.(*schema.Set).List()))
	}
	return hashcode.String(buf.String())
}

func resourceYandexStorageBucketPolicyUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	policy := d.Get("policy").(string)

	log.Printf("[DEBUG] S3 bucket: %s, put policy: %s", bucket, policy)
	if err := s3Client.UpdateBucketPolicy(ctx, bucket, policy); err != nil {
		log.Printf("[ERROR] Got an error while trying to update policy for Storage Bucket %s: %s", bucket, err)
		return err
	}

	log.Printf("[INFO] Updated policy for Storage S3 Bucket: %q", bucket)
	return nil
}

func resourceYandexStorageBucketGrantsUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	rawGrants := d.Get("grant").(*schema.Set).List()

	if len(rawGrants) == 0 {
		log.Printf("[DEBUG] Storage Bucket: %s, Grants fallback to canned ACL", bucket)
		if err := resourceYandexStorageBucketACLUpdate(ctx, s3Client, d); err != nil {
			return fmt.Errorf("error fallback to canned ACL, %s", err)
		}

		return nil
	}

	grants, err := s3.NewGrants(rawGrants)
	if err != nil {
		return fmt.Errorf("error parsing grants: %s", err)
	}

	return s3Client.UpdateBucketGrants(ctx, bucket, grants)
}

func resourceYandexStorageBucketLifecycleUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	rawLifecycleRules := d.Get("lifecycle_rule").([]interface{})

	rules, err := s3.NewLifecycleRules(rawLifecycleRules, d)
	if err != nil {
		return fmt.Errorf("error parsing lifecycle rules: %s", err)
	}

	return s3Client.UpdateBucketLifecycle(ctx, bucket, rules)
}

func resourceYandexStorageBucketServerSideEncryptionConfigurationUpdate(
	ctx context.Context,
	s3Client *s3.Client,
	d *schema.ResourceData,
) error {
	bucket := d.Get("bucket").(string)
	serverSideEncryptionConfiguration := d.Get("server_side_encryption_configuration").([]interface{})

	rules := s3.NewServerSideEncryptionRules(serverSideEncryptionConfiguration)
	return s3Client.UpdateBucketServerSideEncryption(ctx, bucket, rules)
}

func validateStringIsJSON(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if _, err := NormalizeJsonString(v); err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
	}

	return warnings, errors
}

func NormalizeJsonString(jsonString interface{}) (string, error) {
	var j interface{}

	if jsonString == nil || jsonString.(string) == "" {
		return "", nil
	}

	s := jsonString.(string)

	err := json.Unmarshal([]byte(s), &j)
	if err != nil {
		return "", err
	}

	bytes, _ := json.Marshal(j)
	return string(bytes[:]), nil
}

func suppressEquivalentAwsPolicyDiffs(_, old, new string, _ *schema.ResourceData) bool {
	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)
	if err != nil {
		return false
	}

	return equivalent
}

func storageBucketS3SetFunc(keys ...string) schema.SchemaSetFunc {
	return func(v interface{}) int {
		var buf bytes.Buffer
		m, ok := v.(map[string]interface{})

		if !ok {
			return 0
		}

		for _, key := range keys {
			if v, ok := m[key]; ok {
				value := fmt.Sprintf("%v", v)
				buf.WriteString(value + "-")
			}
		}

		return hashcode.String(buf.String())
	}
}

func getAnonymousAccessFlagsSDK(value interface{}) *storagepb.AnonymousAccessFlags {
	schemaSet, ok := value.(*schema.Set)
	if !ok || schemaSet.Len() == 0 {
		return nil
	}

	accessFlags := new(storagepb.AnonymousAccessFlags)
	flags := schemaSet.List()[0].(map[string]interface{})
	if val, ok := flags["list"].(bool); ok {
		accessFlags.List = wrapperspb.Bool(val)
	}
	if val, ok := flags["read"].(bool); ok {
		accessFlags.Read = wrapperspb.Bool(val)
	}
	if val, ok := flags["config_read"].(bool); ok {
		accessFlags.ConfigRead = wrapperspb.Bool(val)
	}

	return accessFlags
}

func suppressPrefixDiffIfFilterPrefixSet(k, old, new string, d *schema.ResourceData) bool {
	// lifecycle_rule.prefix is deprecated in favor of lifecycle_rule.filter.prefix, so we can suppress it
	// if lifecycle_rule.filter.prefix is set and equal
	if prefix, ok := d.GetOk(strings.TrimSuffix(k, "prefix") + "filter.0.prefix"); ok {
		return prefix == new && old == ""
	}
	return false
}

func suppressFilterPrefixDiffIfPrefixSet(k, old, new string, d *schema.ResourceData) bool {
	// lifecycle_rule.prefix is deprecated in favor of lifecycle_rule.filter.prefix, so we can suppress it
	// if lifecycle_rule.filter.prefix is set and equal
	if prefix, ok := d.GetOk(strings.TrimSuffix(k, "filter.0.prefix") + "prefix"); ok {
		if filterPrefix, ok := d.GetOk(k); ok {
			return prefix == filterPrefix
		}
	}
	return false
}

func suppressFilterIfPrefixEqualsFilterPrefix(k, old, new string, d *schema.ResourceData) bool {
	// lifecycle_rule.prefix is deprecated in favor of lifecycle_rule.filter.prefix, so we can suppress it
	// if lifecycle_rule.filter.prefix is set and equal
	if strings.HasSuffix(k, "filter.#") {
		key := strings.TrimSuffix(k, "filter.#")
		prefix := d.Get(key + "prefix")
		filterPrefix := d.Get(key + "filter.0.prefix")
		filterTag := d.Get(key + "filter.0.tag")
		filterAnd := d.Get(key + "filter.0.and")
		if prefix != "" && filterPrefix != "" && prefix == filterPrefix && len(filterTag.([]interface{})) == 0 &&
			len(filterAnd.([]interface{})) == 0 {
			return true
		}
	}
	return false
}
