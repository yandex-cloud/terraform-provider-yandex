package yandex

import (
	"bytes"
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/storage/s3"
)

func resourceYandexStorageBucketV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:          schema.TypeString,
				Description:   "The name of the bucket.",
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
				Description: "The access key to use when applying changes.",
				Optional:    true,
			},

			"secret_key": {
				Type:        schema.TypeString,
				Description: "The secret key to use when applying changes.",
				Optional:    true,
				Sensitive:   true,
			},

			"acl": {
				Type:          schema.TypeString,
				Description:   "The predefined ACL to apply.",
				Optional:      true,
				ConflictsWith: []string{"grant"},
				ValidateFunc:  validation.StringInSlice(bucketACLAllowedValues, false),
			},

			"grant": {
				Type:          schema.TypeSet,
				Description:   "An ACL policy grant.",
				Optional:      true,
				Set:           grantHash,
				ConflictsWith: []string{"acl"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Description: "Canonical user id to grant for.",
							Optional:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "Type of grantee to apply for.",
							Required:    true,
							ValidateFunc: validation.StringInSlice([]string{
								s3.TypeCanonicalUser,
								s3.TypeGroup,
							}, false),
						},
						"uri": {
							Type:        schema.TypeString,
							Description: "URI address to grant for.",
							Optional:    true,
						},

						"permissions": {
							Type:        schema.TypeSet,
							Description: "List of permissions to apply for grantee.",
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
				Description:      "The policy object should contain the only field with the text of the policy.",
				Optional:         true,
				ValidateFunc:     validateStringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"cors_rule": {
				Type:        schema.TypeList,
				Description: "A rule of Cross-Origin Resource Sharing (CORS object).",
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
							Description: "Specifies which methods are allowed.",
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
				Description: "A Website object.",
				Optional:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_document": {
							Type:        schema.TypeString,
							Description: "Storage returns this index document when requests are made to the root domain or any of the subfolders.",
							Optional:    true,
						},

						"error_document": {
							Type:        schema.TypeString,
							Description: "An absolute path to the document to return in case of a 4XX error.",
							Optional:    true,
						},

						"redirect_all_requests_to": {
							Type:        schema.TypeString,
							Description: "A hostname to redirect all website requests for this bucket to.",
							ConflictsWith: []string{
								"website.0.index_document",
								"website.0.error_document",
								"website.0.routing_rules",
							},
							Optional: true,
						},

						"routing_rules": {
							Type:         schema.TypeString,
							Description:  "A JSON array containing routing rules describing redirect behavior and when redirects are applied.",
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
				Description: "The website endpoint, if the bucket is configured with a website.",
				Optional:    true,
				Computed:    true,
			},
			"website_domain": {
				Type:        schema.TypeString,
				Description: "The domain of the website endpoint, if the bucket is configured with a website.",
				Optional:    true,
				Computed:    true,
			},

			"versioning": {
				Type:        schema.TypeList,
				Description: "A state of versioning.",
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Description: "Enable versioning. Once you version-enable a bucket, it can never return to an unversioned state.",
							Optional:    true,
							Default:     false,
						},
					},
				},
			},

			"object_lock_configuration": {
				Type:        schema.TypeList,
				Description: "A configuration of object lock management.",
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
										Type:        schema.TypeList,
										Description: "Default retention object.",
										Required:    true,
										MaxItems:    1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"mode": {
													Type:        schema.TypeString,
													Description: "Specifies a type of object lock.",
													Required:    true,
													ValidateFunc: validation.StringInSlice(
														s3.ObjectLockRetentionModeValues,
														false,
													),
												},
												"days": {
													Type:        schema.TypeInt,
													Description: "Specifies a retention period in days after uploading an object version.",
													Optional:    true,
													ExactlyOneOf: []string{
														"object_lock_configuration.0.rule.0.default_retention.0.days",
														"object_lock_configuration.0.rule.0.default_retention.0.years",
													},
												},
												"years": {
													Type:        schema.TypeInt,
													Description: "Specifies a retention period in years after uploading an object version.",
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
				Description: "A settings of bucket logging.",
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
				Description: "A configuration of object lifecycle management.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Description:  "Unique identifier for the rule.",
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"prefix": {
							Type:        schema.TypeString,
							Description: "Object key prefix identifying one or more objects to which the rule applies.",
							Optional:    true,
						},
						"tags": tagsSchema(),
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
										Description: "In a versioned bucket, you can add this element in the lifecycle configuration to direct Object Storage to delete expired object delete markers.",
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
										Description:  "Specifies the storage class to which you want the object to transition.",
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
										Description:  "Specifies the storage class to which you want the noncurrent object versions to transition.",
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
				Description: "A boolean that indicates all objects should be deleted from the bucket so that the bucket can be destroyed without error.",
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
										Description: "A single object for the default encryption to apply.",
										MaxItems:    1,
										Required:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"kms_master_key_id": {
													Type:        schema.TypeString,
													Description: "The AWS KMS master key ID used for the SSE-KMS encryption.",
													Required:    true,
												},
												"sse_algorithm": {
													Type:        schema.TypeString,
													Description: "The server-side encryption algorithm to use.",
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
				Description: "Storage class which is used for storing objects by default.",
				Optional:    true,
				Computed:    true,
			},

			"folder_id": {
				Type:        schema.TypeString,
				Description: "The ID of the folder to create the resource in.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},

			"max_size": {
				Type:        schema.TypeInt,
				Description: "The size of bucket, in bytes.",
				Optional:    true,
			},

			"anonymous_access_flags": {
				Type:        schema.TypeSet,
				Description: "Provides various access to objects.",
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
				Description: "Manages https certificates for bucket.",
				Optional:    true,
				MaxItems:    1,
				Set:         storageBucketS3SetFunc("certificate_id"),
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_id": {
							Type:        schema.TypeString,
							Description: "ID of the SSL certificate.",
							Required:    true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceYandexStorageBucketStateUpgradeV0(
	ctx context.Context,
	rawState map[string]any,
	meta any,
) (map[string]any, error) {
	if rawState == nil {
		return nil, nil
	}

	if _, ok := rawState["lifecycle_rule"]; ok {
		switch rawState["lifecycle_rule"].(type) {
		case []map[string]interface{}:
			rawLifecycleRules := rawState["lifecycle_rule"].([]map[string]interface{})
			updatedLifecycleRules := make([]map[string]interface{}, len(rawLifecycleRules))
			for i, rule := range rawLifecycleRules {
				newRule := make(map[string]interface{})
				for k, v := range rule {
					if k != "tags" {
						newRule[k] = v
					}
				}
				updatedLifecycleRules[i] = newRule
			}
			rawState["lifecycle_rule"] = updatedLifecycleRules
		}
	}

	return rawState, nil
}
