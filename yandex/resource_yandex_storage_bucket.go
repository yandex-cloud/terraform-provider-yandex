package yandex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	awspolicy "github.com/jen20/awspolicyequivalence"
)

func resourceYandexStorageBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceYandexStorageBucketCreate,
		Read:   resourceYandexStorageBucketRead,
		Update: resourceYandexStorageBucketUpdate,
		Delete: resourceYandexStorageBucketDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		SchemaVersion: 0,

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket_prefix"},
			},
			"bucket_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"bucket"},
			},
			"bucket_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_key": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"secret_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},

			"acl": {
				Type:          schema.TypeString,
				Default:       "private",
				Optional:      true,
				ConflictsWith: []string{"grant"},
			},

			"grant": {
				Type:          schema.TypeSet,
				Optional:      true,
				Set:           grantHash,
				ConflictsWith: []string{"acl"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								s3.TypeCanonicalUser,
								s3.TypeGroup,
							}, false),
						},
						"uri": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"permissions": {
							Type:     schema.TypeSet,
							Required: true,
							Set:      schema.HashString,
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
				Optional:         true,
				ValidateFunc:     validateStringIsJSON,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"cors_rule": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_origins": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"website": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index_document": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"error_document": {
							Type:     schema.TypeString,
							Optional: true,
						},

						"redirect_all_requests_to": {
							Type: schema.TypeString,
							ConflictsWith: []string{
								"website.0.index_document",
								"website.0.error_document",
								"website.0.routing_rules",
							},
							Optional: true,
						},

						"routing_rules": {
							Type:         schema.TypeString,
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
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"website_domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"versioning": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"logging": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"target_bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_prefix": {
							Type:     schema.TypeString,
							Optional: true,
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
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"abort_incomplete_multipart_upload_days": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"expiration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateS3BucketLifecycleTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"expired_object_delete_marker": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"noncurrent_version_expiration": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
						"transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      transitionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"date": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validateS3BucketLifecycleTimestamp,
									},
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											s3.StorageClassStandardIa,
											"COLD",
										}, false),
									},
								},
							},
						},
						"noncurrent_version_transition": {
							Type:     schema.TypeSet,
							Optional: true,
							Set:      transitionHash,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"days": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
									"storage_class": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											s3.StorageClassStandardIa,
											"COLD",
										}, false),
									},
								},
							},
						},
					},
				},
			},

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"server_side_encryption_configuration": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"apply_server_side_encryption_by_default": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"kms_master_key_id": {
													Type:     schema.TypeString,
													Required: true,
												},
												"sse_algorithm": {
													Type:     schema.TypeString,
													Required: true,
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
		},
	}
}

func resourceYandexStorageBucketCreate(d *schema.ResourceData, meta interface{}) error {
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
		return fmt.Errorf("error validating Storage Bucket name: %s", err)
	}

	d.Set("bucket", bucket)
	acl := d.Get("acl").(string)

	config := meta.(*Config)
	s3Client, err := getS3Client(d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		log.Printf("[DEBUG] Trying to create new Storage Bucket: %q, ACL: %q", bucket, acl)
		_, err := s3Client.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucket),
			ACL:    aws.String(acl),
		})
		if awsErr, ok := err.(awserr.Error); ok && (awsErr.Code() == "OperationAborted" ||
			awsErr.Code() == "AccessDenied" || awsErr.Code() == "Forbidden") {
			log.Printf("[WARN] Got an error while trying to create Storage Bucket %s: %s", bucket, err)
			return resource.RetryableError(
				fmt.Errorf("error creating Storage Bucket %s, retrying: %s", bucket, err))
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error creating Storage Bucket: %s", err)
	}

	d.SetId(bucket)

	return resourceYandexStorageBucketUpdate(d, meta)
}

func resourceYandexStorageBucketUpdate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3Client, err := getS3Client(d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	if d.HasChange("policy") {
		if err := resourceYandexStorageBucketPolicyUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("cors_rule") {
		if err := resourceYandexStorageBucketCORSUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("website") {
		if err := resourceYandexStorageBucketWebsiteUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("versioning") {
		if err := resourceYandexStorageBucketVersioningUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("acl") && !d.IsNewResource() {
		if err := resourceYandexStorageBucketACLUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("grant") {
		if err := resourceYandexStorageBucketGrantsUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("logging") {
		if err := resourceYandexStorageBucketLoggingUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("lifecycle_rule") {
		if err := resourceYandexStorageBucketLifecycleUpdate(s3Client, d); err != nil {
			return err
		}
	}

	if d.HasChange("server_side_encryption_configuration") {
		if err := resourceYandexStorageBucketServerSideEncryptionConfigurationUpdate(s3Client, d); err != nil {
			return err
		}
	}

	return resourceYandexStorageBucketRead(d, meta)
}

func resourceYandexStorageBucketRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3Client, err := getS3Client(d, config)

	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	resp, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.HeadBucket(&s3.HeadBucketInput{
			Bucket: aws.String(d.Id()),
		})
	})
	if err != nil {
		if handleS3BucketNotFoundError(d, err) {
			return nil
		}
		return fmt.Errorf("error reading Storage Bucket (%s): %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Storage head bucket output: %#v", resp)

	if _, ok := d.GetOk("bucket"); !ok {
		d.Set("bucket", d.Id())
	}

	domainName, err := bucketDomainName(d.Get("bucket").(string), config.StorageEndpoint)
	if err != nil {
		return fmt.Errorf("error getting bucket domain name: %s", err)
	}
	d.Set("bucket_domain_name", domainName)

	// Read the policy
	pol, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(d.Id()),
		})
	})
	log.Printf("[DEBUG] S3 bucket: %s, read policy: %v", d.Id(), pol)
	if err != nil {
		if isAWSErr(err, "NoSuchBucketPolicy", "") {
			if err := d.Set("policy", ""); err != nil {
				return fmt.Errorf("error setting policy: %s", err)
			}
		} else {
			return fmt.Errorf("error getting current policy: %s", err)
		}
	} else {
		v := pol.(*s3.GetBucketPolicyOutput).Policy
		if v == nil {
			if err := d.Set("policy", ""); err != nil {
				return fmt.Errorf("error setting policy: %s", err)
			}
		} else {
			policy, err := NormalizeJsonString(aws.StringValue(v))
			if err != nil {
				return fmt.Errorf("policy contains an invalid JSON: %s", err)
			}
			if err := d.Set("policy", policy); err != nil {
				return fmt.Errorf("error setting policy: %s", err)
			}
		}
	}

	corsResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.GetBucketCors(&s3.GetBucketCorsInput{
			Bucket: aws.String(d.Id()),
		})
	})
	if err != nil && !isAWSErr(err, "NoSuchCORSConfiguration", "") {
		if handleS3BucketNotFoundError(d, err) {
			return nil
		}
		return fmt.Errorf("error getting Storage Bucket CORS configuration: %s", err)
	}

	corsRules := make([]map[string]interface{}, 0)
	if cors, ok := corsResponse.(*s3.GetBucketCorsOutput); ok && len(cors.CORSRules) > 0 {
		log.Printf("[DEBUG] Storage get bucket CORS output: %#v", cors)

		corsRules = make([]map[string]interface{}, 0, len(cors.CORSRules))
		for _, ruleObject := range cors.CORSRules {
			rule := make(map[string]interface{})
			rule["allowed_headers"] = flattenStringList(ruleObject.AllowedHeaders)
			rule["allowed_methods"] = flattenStringList(ruleObject.AllowedMethods)
			rule["allowed_origins"] = flattenStringList(ruleObject.AllowedOrigins)
			// Both the "ExposeHeaders" and "MaxAgeSeconds" might not be set.
			if ruleObject.ExposeHeaders != nil {
				rule["expose_headers"] = flattenStringList(ruleObject.ExposeHeaders)
			}
			if ruleObject.MaxAgeSeconds != nil {
				rule["max_age_seconds"] = int(*ruleObject.MaxAgeSeconds)
			}
			corsRules = append(corsRules, rule)
		}
	}
	if err := d.Set("cors_rule", corsRules); err != nil {
		return fmt.Errorf("error setting cors_rule: %s", err)
	}

	// Read the website configuration
	wsResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.GetBucketWebsite(&s3.GetBucketWebsiteInput{
			Bucket: aws.String(d.Id()),
		})
	})
	if err != nil && !isAWSErr(err, "NotImplemented", "") && !isAWSErr(err, "NoSuchWebsiteConfiguration", "") {
		if handleS3BucketNotFoundError(d, err) {
			return nil
		}
		return fmt.Errorf("error getting Storage Bucket website configuration: %s", err)
	}

	websites := make([]map[string]interface{}, 0, 1)
	if ws, ok := wsResponse.(*s3.GetBucketWebsiteOutput); ok {
		log.Printf("[DEBUG] Storage get bucket website output: %#v", ws)

		w := make(map[string]interface{})

		if v := ws.IndexDocument; v != nil {
			w["index_document"] = *v.Suffix
		}

		if v := ws.ErrorDocument; v != nil {
			w["error_document"] = *v.Key
		}

		if v := ws.RedirectAllRequestsTo; v != nil {
			if v.Protocol == nil {
				w["redirect_all_requests_to"] = aws.StringValue(v.HostName)
			} else {
				var host string
				var path string
				var query string
				parsedHostName, err := url.Parse(aws.StringValue(v.HostName))
				if err == nil {
					host = parsedHostName.Host
					path = parsedHostName.Path
					query = parsedHostName.RawQuery
				} else {
					host = aws.StringValue(v.HostName)
					path = ""
				}

				w["redirect_all_requests_to"] = (&url.URL{
					Host:     host,
					Path:     path,
					Scheme:   aws.StringValue(v.Protocol),
					RawQuery: query,
				}).String()
			}
		}

		if v := ws.RoutingRules; v != nil {
			rr, err := normalizeRoutingRules(v)
			if err != nil {
				return fmt.Errorf("Error while marshaling routing rules: %s", err)
			}
			w["routing_rules"] = rr
		}

		// We have special handling for the website configuration,
		// so only add the configuration if there is any
		if len(w) > 0 {
			websites = append(websites, w)
		}
	}
	if err := d.Set("website", websites); err != nil {
		return fmt.Errorf("error setting website: %s", err)
	}

	// Add website_endpoint as an attribute
	websiteEndpoint, err := websiteEndpoint(s3Client, d)
	if err != nil {
		return err
	}
	if websiteEndpoint != nil {
		if err := d.Set("website_endpoint", websiteEndpoint.Endpoint); err != nil {
			return fmt.Errorf("error setting website_endpoint: %s", err)
		}
		if err := d.Set("website_domain", websiteEndpoint.Domain); err != nil {
			return fmt.Errorf("error setting website_domain: %s", err)
		}
	}

	//Read the Grant ACL. Reset if `acl` (canned ACL) is set.
	if acl, ok := d.GetOk("acl"); ok && acl.(string) != "private" {
		if err := d.Set("grant", nil); err != nil {
			return fmt.Errorf("error resetting Storage Bucket `grant` %s", err)
		}
	} else {
		apResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
			return s3Client.GetBucketAcl(&s3.GetBucketAclInput{
				Bucket: aws.String(d.Id()),
			})
		})

		if err != nil {
			//Ignore access denied error, when reading ACL for bucket.
			if awsErr, ok := err.(awserr.Error); ok && (awsErr.Code() == "AccessDenied" || awsErr.Code() == "Forbidden") {
				log.Printf("[WARN] Got an error while trying to read Storage Bucket (%s) ACL: %s", d.Id(), err)

				if err := d.Set("grant", nil); err != nil {
					return fmt.Errorf("error resetting Storage Bucket `grant` %s", err)
				}

				return nil
			}

			return fmt.Errorf("error getting Storage Bucket (%s) ACL: %s", d.Id(), err)
		} else {
			log.Printf("[DEBUG] getting storage: %s, read ACL grants policy: %+v", d.Id(), apResponse)
			grants := flattenGrants(apResponse.(*s3.GetBucketAclOutput))
			if err := d.Set("grant", schema.NewSet(grantHash, grants)); err != nil {
				return fmt.Errorf("error setting Storage Bucket `grant` %s", err)
			}
		}
	}

	// Read the versioning configuration

	versioningResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: aws.String(d.Id()),
		})
	})
	if err != nil {
		return err
	}

	vcl := make([]map[string]interface{}, 0, 1)
	if versioning, ok := versioningResponse.(*s3.GetBucketVersioningOutput); ok {
		vc := make(map[string]interface{})
		if versioning.Status != nil && aws.StringValue(versioning.Status) == s3.BucketVersioningStatusEnabled {
			vc["enabled"] = true
		} else {
			vc["enabled"] = false
		}

		vcl = append(vcl, vc)
	}
	if err := d.Set("versioning", vcl); err != nil {
		return fmt.Errorf("error setting versioning: %s", err)
	}

	// Read the logging configuration
	loggingResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.GetBucketLogging(&s3.GetBucketLoggingInput{
			Bucket: aws.String(d.Id()),
		})
	})

	if err != nil {
		return fmt.Errorf("error getting S3 Bucket logging: %s", err)
	}

	lcl := make([]map[string]interface{}, 0, 1)
	if logging, ok := loggingResponse.(*s3.GetBucketLoggingOutput); ok && logging.LoggingEnabled != nil {
		v := logging.LoggingEnabled
		lc := make(map[string]interface{})
		if aws.StringValue(v.TargetBucket) != "" {
			lc["target_bucket"] = aws.StringValue(v.TargetBucket)
		}
		if aws.StringValue(v.TargetPrefix) != "" {
			lc["target_prefix"] = aws.StringValue(v.TargetPrefix)
		}
		lcl = append(lcl, lc)
	}
	if err := d.Set("logging", lcl); err != nil {
		return fmt.Errorf("error setting logging: %s", err)
	}

	// Read the lifecycle configuration

	lifecycleResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(d.Id()),
		})
	})
	if err != nil && !isAWSErr(err, "NoSuchLifecycleConfiguration", "") {
		return err
	}

	lifecycleRules := make([]map[string]interface{}, 0)
	if lifecycle, ok := lifecycleResponse.(*s3.GetBucketLifecycleConfigurationOutput); ok && len(lifecycle.Rules) > 0 {
		lifecycleRules = make([]map[string]interface{}, 0, len(lifecycle.Rules))

		for _, lifecycleRule := range lifecycle.Rules {
			log.Printf("[DEBUG] S3 bucket: %s, read lifecycle rule: %v", d.Id(), lifecycleRule)
			rule := make(map[string]interface{})

			// ID
			if lifecycleRule.ID != nil && aws.StringValue(lifecycleRule.ID) != "" {
				rule["id"] = aws.StringValue(lifecycleRule.ID)
			}
			filter := lifecycleRule.Filter
			if filter != nil {
				if filter.And != nil {
					// Prefix
					if filter.And.Prefix != nil && aws.StringValue(filter.And.Prefix) != "" {
						rule["prefix"] = aws.StringValue(filter.And.Prefix)
					}
				} else {
					// Prefix
					if filter.Prefix != nil && aws.StringValue(filter.Prefix) != "" {
						rule["prefix"] = aws.StringValue(filter.Prefix)
					}
				}
			}

			// Enabled
			if lifecycleRule.Status != nil {
				if aws.StringValue(lifecycleRule.Status) == s3.ExpirationStatusEnabled {
					rule["enabled"] = true
				} else {
					rule["enabled"] = false
				}
			}

			// AbortIncompleteMultipartUploadDays
			if lifecycleRule.AbortIncompleteMultipartUpload != nil {
				if lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation != nil {
					rule["abort_incomplete_multipart_upload_days"] = int(aws.Int64Value(lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation))
				}
			}

			// expiration
			if lifecycleRule.Expiration != nil {
				e := make(map[string]interface{})
				if lifecycleRule.Expiration.Date != nil {
					e["date"] = (aws.TimeValue(lifecycleRule.Expiration.Date)).Format("2006-01-02")
				}
				if lifecycleRule.Expiration.Days != nil {
					e["days"] = int(aws.Int64Value(lifecycleRule.Expiration.Days))
				}
				if lifecycleRule.Expiration.ExpiredObjectDeleteMarker != nil {
					e["expired_object_delete_marker"] = aws.BoolValue(lifecycleRule.Expiration.ExpiredObjectDeleteMarker)
				}
				rule["expiration"] = []interface{}{e}
			}
			// noncurrent_version_expiration
			if lifecycleRule.NoncurrentVersionExpiration != nil {
				e := make(map[string]interface{})
				if lifecycleRule.NoncurrentVersionExpiration.NoncurrentDays != nil {
					e["days"] = int(aws.Int64Value(lifecycleRule.NoncurrentVersionExpiration.NoncurrentDays))
				}
				rule["noncurrent_version_expiration"] = []interface{}{e}
			}
			//// transition
			if len(lifecycleRule.Transitions) > 0 {
				transitions := make([]interface{}, 0, len(lifecycleRule.Transitions))
				for _, v := range lifecycleRule.Transitions {
					t := make(map[string]interface{})
					if v.Date != nil {
						t["date"] = (aws.TimeValue(v.Date)).Format("2006-01-02")
					}
					if v.Days != nil {
						t["days"] = int(aws.Int64Value(v.Days))
					}
					if v.StorageClass != nil {
						t["storage_class"] = aws.StringValue(v.StorageClass)
					}
					transitions = append(transitions, t)
				}
				rule["transition"] = schema.NewSet(transitionHash, transitions)
			}
			// noncurrent_version_transition
			if len(lifecycleRule.NoncurrentVersionTransitions) > 0 {
				transitions := make([]interface{}, 0, len(lifecycleRule.NoncurrentVersionTransitions))
				for _, v := range lifecycleRule.NoncurrentVersionTransitions {
					t := make(map[string]interface{})
					if v.NoncurrentDays != nil {
						t["days"] = int(aws.Int64Value(v.NoncurrentDays))
					}
					if v.StorageClass != nil {
						t["storage_class"] = aws.StringValue(v.StorageClass)
					}
					transitions = append(transitions, t)
				}
				rule["noncurrent_version_transition"] = schema.NewSet(transitionHash, transitions)
			}

			lifecycleRules = append(lifecycleRules, rule)
		}
	}
	if err := d.Set("lifecycle_rule", lifecycleRules); err != nil {
		return fmt.Errorf("error setting lifecycle_rule: %s", err)
	}

	// Read the bucket server side encryption configuration

	encryptionResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.GetBucketEncryption(&s3.GetBucketEncryptionInput{
			Bucket: aws.String(d.Id()),
		})
	})
	if err != nil && !isAWSErr(err, "ServerSideEncryptionConfigurationNotFoundError", "encryption configuration was not found") {
		return fmt.Errorf("error getting S3 Bucket encryption: %s", err)
	}

	serverSideEncryptionConfiguration := make([]map[string]interface{}, 0)
	if encryption, ok := encryptionResponse.(*s3.GetBucketEncryptionOutput); ok && encryption.ServerSideEncryptionConfiguration != nil {
		serverSideEncryptionConfiguration = flattenS3ServerSideEncryptionConfiguration(encryption.ServerSideEncryptionConfiguration)
	}
	if err := d.Set("server_side_encryption_configuration", serverSideEncryptionConfiguration); err != nil {
		return fmt.Errorf("error setting server_side_encryption_configuration: %s", err)
	}

	return nil
}

func resourceYandexStorageBucketDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	s3Client, err := getS3Client(d, config)
	if err != nil {
		return fmt.Errorf("error getting storage client: %s", err)
	}

	log.Printf("[DEBUG] Storage Delete Bucket: %s", d.Id())

	_, err = retryOnAwsCodes([]string{"AccessDenied", "Forbidden"}, func() (interface{}, error) {
		return s3Client.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(d.Id()),
		})
	})

	if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
		return nil
	}

	if isAWSErr(err, "BucketNotEmpty", "") {
		if d.Get("force_destroy").(bool) {
			// bucket may have things delete them
			log.Printf("[DEBUG] Storage Bucket attempting to forceDestroy %+v", err)

			bucket := d.Get("bucket").(string)
			resp, err := s3Client.ListObjectVersions(
				&s3.ListObjectVersionsInput{
					Bucket: aws.String(bucket),
				},
			)

			if err != nil {
				return fmt.Errorf("error listing Storage Bucket object versions: %s", err)
			}

			objectsToDelete := make([]*s3.ObjectIdentifier, 0)

			if len(resp.DeleteMarkers) != 0 {
				for _, v := range resp.DeleteMarkers {
					objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
						Key:       v.Key,
						VersionId: v.VersionId,
					})
				}
			}

			if len(resp.Versions) != 0 {
				for _, v := range resp.Versions {
					objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
						Key:       v.Key,
						VersionId: v.VersionId,
					})
				}
			}

			params := &s3.DeleteObjectsInput{
				Bucket: aws.String(bucket),
				Delete: &s3.Delete{
					Objects: objectsToDelete,
				},
			}

			_, err = s3Client.DeleteObjects(params)

			if err != nil {
				return fmt.Errorf("error force_destroy deleting Storage Bucket (%s): %s", d.Id(), err)
			}

			// this line recurses until all objects are deleted or an error is returned
			return resourceYandexStorageBucketDelete(d, meta)
		}
	}

	if err == nil {
		req := &s3.HeadBucketInput{
			Bucket: aws.String(d.Id()),
		}
		err = waitConditionStable(func() (bool, error) {
			_, err := s3Client.HeadBucket(req)
			if awsError, ok := err.(awserr.RequestFailure); ok && awsError.StatusCode() == 404 {
				return true, nil
			}
			return false, err
		})
	}

	if err != nil {
		return fmt.Errorf("error deleting Storage Bucket (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceYandexStorageBucketCORSUpdate(s3Client *s3.S3, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	rawCors := d.Get("cors_rule").([]interface{})

	if len(rawCors) == 0 {
		// Delete CORS
		log.Printf("[DEBUG] Storage Bucket: %s, delete CORS", bucket)

		_, err := retryFlakyS3Responses(func() (interface{}, error) {
			return s3Client.DeleteBucketCors(&s3.DeleteBucketCorsInput{
				Bucket: aws.String(bucket),
			})
		})
		if err == nil {
			err = waitCorsDeleted(s3Client, bucket)
		}
		if err != nil {
			return fmt.Errorf("error deleting storage CORS: %s", err)
		}
	} else {
		// Put CORS
		rules := make([]*s3.CORSRule, 0, len(rawCors))
		for _, cors := range rawCors {
			corsMap := cors.(map[string]interface{})
			r := &s3.CORSRule{}
			for k, v := range corsMap {
				log.Printf("[DEBUG] Storage Bucket: %s, put CORS: %#v, %#v", bucket, k, v)
				if k == "max_age_seconds" {
					r.MaxAgeSeconds = aws.Int64(int64(v.(int)))
				} else {
					vMap := make([]*string, len(v.([]interface{})))
					for i, vv := range v.([]interface{}) {
						var value string
						if str, ok := vv.(string); ok {
							value = str
						}
						vMap[i] = aws.String(value)
					}
					switch k {
					case "allowed_headers":
						r.AllowedHeaders = vMap
					case "allowed_methods":
						r.AllowedMethods = vMap
					case "allowed_origins":
						r.AllowedOrigins = vMap
					case "expose_headers":
						r.ExposeHeaders = vMap
					}
				}
			}
			rules = append(rules, r)
		}
		corsConfiguration := &s3.CORSConfiguration{
			CORSRules: rules,
		}
		corsInput := &s3.PutBucketCorsInput{
			Bucket:            aws.String(bucket),
			CORSConfiguration: corsConfiguration,
		}
		log.Printf("[DEBUG] Storage Bucket: %s, put CORS: %#v", bucket, corsInput)

		_, err := retryFlakyS3Responses(func() (interface{}, error) {
			return s3Client.PutBucketCors(corsInput)
		})
		if err == nil {
			err = waitCorsPut(s3Client, bucket, corsConfiguration)
		}
		if err != nil {
			return fmt.Errorf("error putting bucket CORS: %s", err)
		}
	}

	return nil
}

func resourceYandexStorageBucketWebsiteUpdate(s3Client *s3.S3, d *schema.ResourceData) error {
	ws := d.Get("website").([]interface{})

	if len(ws) == 0 {
		return resourceYandexStorageBucketWebsiteDelete(s3Client, d)
	}

	var w map[string]interface{}
	if ws[0] != nil {
		w = ws[0].(map[string]interface{})
	} else {
		w = make(map[string]interface{})
	}

	return resourceYandexStorageBucketWebsitePut(s3Client, d, w)
}

func resourceYandexStorageBucketWebsitePut(s3Client *s3.S3, d *schema.ResourceData, website map[string]interface{}) error {
	bucket := d.Get("bucket").(string)

	var indexDocument, errorDocument, redirectAllRequestsTo, routingRules string
	if v, ok := website["index_document"]; ok {
		indexDocument = v.(string)
	}
	if v, ok := website["error_document"]; ok {
		errorDocument = v.(string)
	}

	if v, ok := website["redirect_all_requests_to"]; ok {
		redirectAllRequestsTo = v.(string)
	}
	if v, ok := website["routing_rules"]; ok {
		routingRules = v.(string)
	}

	if indexDocument == "" && redirectAllRequestsTo == "" {
		return fmt.Errorf("Must specify either index_document or redirect_all_requests_to.")
	}

	websiteConfiguration := &s3.WebsiteConfiguration{}

	if indexDocument != "" {
		websiteConfiguration.IndexDocument = &s3.IndexDocument{Suffix: aws.String(indexDocument)}
	}

	if errorDocument != "" {
		websiteConfiguration.ErrorDocument = &s3.ErrorDocument{Key: aws.String(errorDocument)}
	}

	if redirectAllRequestsTo != "" {
		redirect, err := url.Parse(redirectAllRequestsTo)
		if err == nil && redirect.Scheme != "" {
			var redirectHostBuf bytes.Buffer
			redirectHostBuf.WriteString(redirect.Host)
			if redirect.Path != "" {
				redirectHostBuf.WriteString(redirect.Path)
			}
			if redirect.RawQuery != "" {
				redirectHostBuf.WriteString("?")
				redirectHostBuf.WriteString(redirect.RawQuery)
			}
			websiteConfiguration.RedirectAllRequestsTo = &s3.RedirectAllRequestsTo{HostName: aws.String(redirectHostBuf.String()), Protocol: aws.String(redirect.Scheme)}
		} else {
			websiteConfiguration.RedirectAllRequestsTo = &s3.RedirectAllRequestsTo{HostName: aws.String(redirectAllRequestsTo)}
		}
	}

	if routingRules != "" {
		var unmarshaledRules []*s3.RoutingRule
		if err := json.Unmarshal([]byte(routingRules), &unmarshaledRules); err != nil {
			return err
		}
		websiteConfiguration.RoutingRules = unmarshaledRules
	}

	putInput := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfiguration,
	}

	log.Printf("[DEBUG] Storage put bucket website: %#v", putInput)

	_, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.PutBucketWebsite(putInput)
	})
	if err == nil {
		err = waitWebsitePut(s3Client, bucket, websiteConfiguration)
	}
	if err != nil {
		return fmt.Errorf("error putting storage website: %s", err)
	}

	return nil
}

func resourceYandexStorageBucketWebsiteDelete(s3Client *s3.S3, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	deleteInput := &s3.DeleteBucketWebsiteInput{Bucket: aws.String(bucket)}

	log.Printf("[DEBUG] Storage delete bucket website: %#v", deleteInput)

	_, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.DeleteBucketWebsite(deleteInput)
	})
	if err == nil {
		err = waitWebsiteDeleted(s3Client, bucket)
	}
	if err != nil {
		return fmt.Errorf("error deleting storage website: %s", err)
	}

	d.Set("website_endpoint", "")
	d.Set("website_domain", "")

	return nil
}

func websiteEndpoint(s3Client *s3.S3, d *schema.ResourceData) (*S3Website, error) {
	// If the bucket doesn't have a website configuration, return an empty
	// endpoint
	if _, ok := d.GetOk("website"); !ok {
		return nil, nil
	}

	bucket := d.Get("bucket").(string)

	return WebsiteEndpoint(bucket), nil
}

func WebsiteEndpoint(bucket string) *S3Website {
	domain := WebsiteDomainURL()
	return &S3Website{Endpoint: fmt.Sprintf("%s.%s", bucket, domain), Domain: domain}
}

func WebsiteDomainURL() string {
	return "website.yandexcloud.net"
}

func resourceYandexStorageBucketACLUpdate(s3Client *s3.S3, d *schema.ResourceData) error {
	acl := d.Get("acl").(string)
	bucket := d.Get("bucket").(string)

	i := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(acl),
	}
	log.Printf("[DEBUG] Storage put bucket ACL: %#v", i)

	_, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3Client.PutBucketAcl(i)
	})
	if err != nil {
		return fmt.Errorf("error putting Storage Bucket ACL: %s", err)
	}

	return nil
}

func resourceYandexStorageBucketVersioningUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	v := d.Get("versioning").([]interface{})
	bucket := d.Get("bucket").(string)
	vc := &s3.VersioningConfiguration{}

	if len(v) > 0 {
		c := v[0].(map[string]interface{})

		if c["enabled"].(bool) {
			vc.Status = aws.String(s3.BucketVersioningStatusEnabled)
		} else {
			vc.Status = aws.String(s3.BucketVersioningStatusSuspended)
		}

	} else {
		vc.Status = aws.String(s3.BucketVersioningStatusSuspended)
	}

	i := &s3.PutBucketVersioningInput{
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: vc,
	}
	log.Printf("[DEBUG] S3 put bucket versioning: %#v", i)

	_, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3conn.PutBucketVersioning(i)
	})
	if err != nil {
		return fmt.Errorf("Error putting S3 versioning: %s", err)
	}

	return nil
}

func resourceYandexStorageBucketLoggingUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	logging := d.Get("logging").(*schema.Set).List()
	bucket := d.Get("bucket").(string)
	loggingStatus := &s3.BucketLoggingStatus{}

	if len(logging) > 0 {
		c := logging[0].(map[string]interface{})

		loggingEnabled := &s3.LoggingEnabled{}
		if val, ok := c["target_bucket"]; ok {
			loggingEnabled.TargetBucket = aws.String(val.(string))
		}
		if val, ok := c["target_prefix"]; ok {
			loggingEnabled.TargetPrefix = aws.String(val.(string))
		}

		loggingStatus.LoggingEnabled = loggingEnabled
	}

	i := &s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: loggingStatus,
	}
	log.Printf("[DEBUG] S3 put bucket logging: %#v", i)

	_, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3conn.PutBucketLogging(i)
	})
	if err != nil {
		return fmt.Errorf("Error putting S3 logging: %s", err)
	}

	return nil
}

func bucketDomainName(bucket string, endpointURL string) (string, error) {
	// Without a scheme the url will not be parsed as we expect
	// See https://github.com/golang/go/issues/19779
	if !strings.Contains(endpointURL, "//") {
		endpointURL = "//" + endpointURL
	}

	parse, err := url.Parse(endpointURL)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", bucket, parse.Hostname()), nil
}

type S3Website struct {
	Endpoint, Domain string
}

func retryOnAwsCodes(codes []string, f func() (interface{}, error)) (interface{}, error) {
	var resp interface{}
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			awsErr, ok := err.(awserr.Error)
			if ok {
				for _, code := range codes {
					if awsErr.Code() == code {
						return resource.RetryableError(err)
					}
				}
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	return resp, err
}

func retryFlakyS3Responses(f func() (interface{}, error)) (interface{}, error) {
	return retryOnAwsCodes([]string{"NoSuchBucket", "AccessDenied", "Forbidden"}, f)
}

func waitConditionStable(check func() (bool, error)) error {
	for checks := 0; checks < 12; checks++ {
		allOk := true
		for subchecks := 0; allOk && subchecks < 10; subchecks++ {
			ok, err := check()
			if err != nil {
				return err
			}
			allOk = allOk && ok
			if ok {
				time.Sleep(time.Second)
			}
		}
		if allOk {
			return nil
		}
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("timeout exceeded")
}

func waitWebsitePut(s3Client *s3.S3, bucket string, configuration *s3.WebsiteConfiguration) error {
	input := &s3.GetBucketWebsiteInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		output, err := s3Client.GetBucketWebsite(input)
		if err != nil && !isAWSErr(err, "NoSuchWebsiteConfiguration", "") {
			return false, err
		}
		outputConfiguration := &s3.WebsiteConfiguration{
			ErrorDocument:         output.ErrorDocument,
			IndexDocument:         output.IndexDocument,
			RedirectAllRequestsTo: output.RedirectAllRequestsTo,
			RoutingRules:          output.RoutingRules,
		}
		if reflect.DeepEqual(outputConfiguration, configuration) {
			return true, nil
		}
		return false, nil
	}

	err := waitConditionStable(check)
	if err != nil {
		return fmt.Errorf("error assuring bucket %q website updated: %s", bucket, err)
	}
	return nil
}

func waitWebsiteDeleted(s3Client *s3.S3, bucket string) error {
	input := &s3.GetBucketWebsiteInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		_, err := s3Client.GetBucketWebsite(input)
		if isAWSErr(err, "NoSuchWebsiteConfiguration", "") {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}

	err := waitConditionStable(check)
	if err != nil {
		return fmt.Errorf("error assuring bucket %q website deleted: %s", bucket, err)
	}
	return nil
}

func waitCorsPut(s3Client *s3.S3, bucket string, configuration *s3.CORSConfiguration) error {
	input := &s3.GetBucketCorsInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		output, err := s3Client.GetBucketCors(input)
		if err != nil && !isAWSErr(err, "NoSuchCORSConfiguration", "") {
			return false, err
		}
		empty := len(output.CORSRules) == 0 && len(configuration.CORSRules) == 0
		for _, rule := range output.CORSRules {
			if rule.ExposeHeaders == nil {
				rule.ExposeHeaders = make([]*string, 0)
			}
			if rule.AllowedHeaders == nil {
				rule.AllowedHeaders = make([]*string, 0)
			}
		}
		if empty || reflect.DeepEqual(output.CORSRules, configuration.CORSRules) {
			return true, nil
		}
		return false, nil
	}

	err := waitConditionStable(check)
	if err != nil {
		return fmt.Errorf("error assuring bucket %q CORS updated: %s", bucket, err)
	}
	return nil
}

func waitCorsDeleted(s3Client *s3.S3, bucket string) error {
	input := &s3.GetBucketCorsInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		_, err := s3Client.GetBucketCors(input)
		if isAWSErr(err, "NoSuchCORSConfiguration", "") {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}

	err := waitConditionStable(check)
	if err != nil {
		return fmt.Errorf("error assuring bucket %q CORS deleted: %s", bucket, err)
	}
	return nil
}

// Returns true if the error matches all these conditions:
//  * err is of type awserr.Error
//  * Error.Code() matches code
//  * Error.Message() contains message
func isAWSErr(err error, code string, message string) bool {
	if err, ok := err.(awserr.Error); ok {
		return err.Code() == code && strings.Contains(err.Message(), message)
	}
	return false
}

func handleS3BucketNotFoundError(d *schema.ResourceData, err error) bool {
	if awsError, ok := err.(awserr.RequestFailure); ok && awsError.StatusCode() == 404 {
		log.Printf("[WARN] Storage Bucket (%s) not found, error code (404)", d.Id())
		d.SetId("")
		return true
	}
	return false
}

// Takes list of pointers to strings. Expand to an array
// of raw strings and returns a []interface{}
// to keep compatibility w/ schema.NewSetschema.NewSet
func flattenStringList(list []*string) []interface{} {
	vs := make([]interface{}, 0, len(list))
	for _, v := range list {
		vs = append(vs, *v)
	}
	return vs
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

func transitionHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})

	if !ok {
		return 0
	}

	if v, ok := m["date"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["days"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["storage_class"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}

func resourceYandexStorageBucketPolicyUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	policy := d.Get("policy").(string)

	if policy == "" {
		log.Printf("[DEBUG] S3 bucket: %s, delete policy: %s", bucket, policy)
		_, err := retryFlakyS3Responses(func() (interface{}, error) {
			return s3conn.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
				Bucket: aws.String(bucket),
			})
		})

		if err != nil {
			return fmt.Errorf("Error deleting S3 policy: %s", err)
		}
		return nil
	}
	log.Printf("[DEBUG] S3 bucket: %s, put policy: %s", bucket, policy)

	params := &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := s3conn.PutBucketPolicy(params)
		if isAWSErr(err, "MalformedPolicy", "") || isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
			return resource.RetryableError(err)
		}
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error putting S3 policy: %s", err)
	}

	return nil
}

func resourceYandexStorageBucketGrantsUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	rawGrants := d.Get("grant").(*schema.Set).List()

	if len(rawGrants) == 0 {
		log.Printf("[DEBUG] Storage Bucket: %s, Grants fallback to canned ACL", bucket)
		if err := resourceYandexStorageBucketACLUpdate(s3conn, d); err != nil {
			return fmt.Errorf("error fallback to canned ACL, %s", err)
		}
	} else {
		apResponse, err := retryFlakyS3Responses(func() (interface{}, error) {
			return s3conn.GetBucketAcl(&s3.GetBucketAclInput{
				Bucket: aws.String(bucket),
			})
		})

		if err != nil {
			return fmt.Errorf("error getting Storage Bucket (%s) ACL: %s", bucket, err)
		}

		ap := apResponse.(*s3.GetBucketAclOutput)
		log.Printf("[DEBUG] Storage Bucket: %s, read ACL grants policy: %+v", bucket, ap)

		grants := make([]*s3.Grant, 0, len(rawGrants))
		for _, rawGrant := range rawGrants {
			log.Printf("[DEBUG] Storage Bucket: %s, put grant: %#v", bucket, rawGrant)
			grantMap := rawGrant.(map[string]interface{})
			permissions := grantMap["permissions"].(*schema.Set).List()
			if err := validateBucketPermissions(permissions); err != nil {
				return err
			}
			for _, rawPermission := range permissions {
				ge := &s3.Grantee{}
				if i, ok := grantMap["id"].(string); ok && i != "" {
					ge.SetID(i)
				}
				if t, ok := grantMap["type"].(string); ok && t != "" {
					ge.SetType(t)
				}
				if u, ok := grantMap["uri"].(string); ok && u != "" {
					ge.SetURI(u)
				}

				g := &s3.Grant{
					Grantee:    ge,
					Permission: aws.String(rawPermission.(string)),
				}
				grants = append(grants, g)
			}
		}

		grantsInput := &s3.PutBucketAclInput{
			Bucket: aws.String(bucket),
			AccessControlPolicy: &s3.AccessControlPolicy{
				Grants: grants,
				Owner:  ap.Owner,
			},
		}

		log.Printf("[DEBUG] Bucket: %s, put Grants: %#v", bucket, grantsInput)

		_, err = retryFlakyS3Responses(func() (interface{}, error) {
			return s3conn.PutBucketAcl(grantsInput)
		})

		if err != nil {
			return fmt.Errorf("error putting Storage Bucket (%s) ACL: %s", bucket, err)
		}
	}
	return nil
}

func resourceYandexStorageBucketLifecycleUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)

	lifecycleRules := d.Get("lifecycle_rule").([]interface{})

	if len(lifecycleRules) == 0 || lifecycleRules[0] == nil {
		i := &s3.DeleteBucketLifecycleInput{
			Bucket: aws.String(bucket),
		}

		_, err := s3conn.DeleteBucketLifecycle(i)
		if err != nil {
			return fmt.Errorf("Error removing S3 lifecycle: %s", err)
		}
		return nil
	}

	rules := make([]*s3.LifecycleRule, 0, len(lifecycleRules))

	for i, lifecycleRule := range lifecycleRules {
		r := lifecycleRule.(map[string]interface{})

		rule := &s3.LifecycleRule{}

		// Filter
		filter := &s3.LifecycleRuleFilter{}
		filter.SetPrefix(r["prefix"].(string))
		rule.SetFilter(filter)

		// ID
		if val, ok := r["id"].(string); ok && val != "" {
			rule.ID = aws.String(val)
		} else {
			rule.ID = aws.String(resource.PrefixedUniqueId("tf-s3-lifecycle-"))
		}

		// Enabled
		if val, ok := r["enabled"].(bool); ok && val {
			rule.Status = aws.String(s3.ExpirationStatusEnabled)
		} else {
			rule.Status = aws.String(s3.ExpirationStatusDisabled)
		}

		// AbortIncompleteMultipartUpload
		if val, ok := r["abort_incomplete_multipart_upload_days"].(int); ok && val > 0 {
			rule.AbortIncompleteMultipartUpload = &s3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int64(int64(val)),
			}
		}

		// Expiration
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).([]interface{})
		if len(expiration) > 0 && expiration[0] != nil {
			e := expiration[0].(map[string]interface{})
			i := &s3.LifecycleExpiration{}
			if val, ok := e["date"].(string); ok && val != "" {
				t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", val))
				if err != nil {
					return fmt.Errorf("Error Parsing AWS S3 Bucket Lifecycle Expiration Date: %s", err.Error())
				}
				i.Date = aws.Time(t)
			} else if val, ok := e["days"].(int); ok && val > 0 {
				i.Days = aws.Int64(int64(val))
			} else if val, ok := e["expired_object_delete_marker"].(bool); ok {
				i.ExpiredObjectDeleteMarker = aws.Bool(val)
			}
			rule.Expiration = i
		}

		// NoncurrentVersionExpiration
		nc_expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_expiration", i)).([]interface{})
		if len(nc_expiration) > 0 && nc_expiration[0] != nil {
			e := nc_expiration[0].(map[string]interface{})

			if val, ok := e["days"].(int); ok && val > 0 {
				rule.NoncurrentVersionExpiration = &s3.NoncurrentVersionExpiration{
					NoncurrentDays: aws.Int64(int64(val)),
				}
			}
		}

		// Transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transition", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			rule.Transitions = make([]*s3.Transition, 0, len(transitions))
			for _, transition := range transitions {
				transition := transition.(map[string]interface{})
				i := &s3.Transition{}
				if val, ok := transition["date"].(string); ok && val != "" {
					t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", val))
					if err != nil {
						return fmt.Errorf("Error Parsing AWS S3 Bucket Lifecycle Expiration Date: %s", err.Error())
					}
					i.Date = aws.Time(t)
				} else if val, ok := transition["days"].(int); ok && val >= 0 {
					i.Days = aws.Int64(int64(val))
				}
				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = aws.String(val)
				}

				rule.Transitions = append(rule.Transitions, i)
			}
		}
		// NoncurrentVersionTransitions
		nc_transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_transition", i)).(*schema.Set).List()
		if len(nc_transitions) > 0 {
			rule.NoncurrentVersionTransitions = make([]*s3.NoncurrentVersionTransition, 0, len(nc_transitions))
			for _, transition := range nc_transitions {
				transition := transition.(map[string]interface{})
				i := &s3.NoncurrentVersionTransition{}
				if val, ok := transition["days"].(int); ok && val >= 0 {
					i.NoncurrentDays = aws.Int64(int64(val))
				}
				if val, ok := transition["storage_class"].(string); ok && val != "" {
					i.StorageClass = aws.String(val)
				}

				rule.NoncurrentVersionTransitions = append(rule.NoncurrentVersionTransitions, i)
			}
		}

		// As a lifecycle rule requires 1 or more transition/expiration actions,
		// we explicitly pass a default ExpiredObjectDeleteMarker value to be able to create
		// the rule while keeping the policy unaffected if the conditions are not met.
		if rule.Expiration == nil && rule.NoncurrentVersionExpiration == nil &&
			rule.Transitions == nil && rule.NoncurrentVersionTransitions == nil &&
			rule.AbortIncompleteMultipartUpload == nil {
			rule.Expiration = &s3.LifecycleExpiration{ExpiredObjectDeleteMarker: aws.Bool(false)}
		}

		rules = append(rules, rule)
	}

	i := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: rules,
		},
	}

	_, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3conn.PutBucketLifecycleConfiguration(i)
	})
	if err != nil {
		return fmt.Errorf("Error putting S3 lifecycle: %s", err)
	}

	return nil
}

func resourceYandexStorageBucketServerSideEncryptionConfigurationUpdate(s3conn *s3.S3, d *schema.ResourceData) error {
	bucket := d.Get("bucket").(string)
	serverSideEncryptionConfiguration := d.Get("server_side_encryption_configuration").([]interface{})
	if len(serverSideEncryptionConfiguration) == 0 {
		log.Printf("[DEBUG] Delete server side encryption configuration: %#v", serverSideEncryptionConfiguration)
		i := &s3.DeleteBucketEncryptionInput{
			Bucket: aws.String(bucket),
		}

		_, err := s3conn.DeleteBucketEncryption(i)
		if err != nil {
			return fmt.Errorf("error removing S3 bucket server side encryption: %s", err)
		}
		return nil
	}

	c := serverSideEncryptionConfiguration[0].(map[string]interface{})

	rc := &s3.ServerSideEncryptionConfiguration{}

	rcRules := c["rule"].([]interface{})
	var rules []*s3.ServerSideEncryptionRule
	for _, v := range rcRules {
		rr := v.(map[string]interface{})
		rrDefault := rr["apply_server_side_encryption_by_default"].([]interface{})
		sseAlgorithm := rrDefault[0].(map[string]interface{})["sse_algorithm"].(string)
		kmsMasterKeyId := rrDefault[0].(map[string]interface{})["kms_master_key_id"].(string)
		rcDefaultRule := &s3.ServerSideEncryptionByDefault{
			SSEAlgorithm: aws.String(sseAlgorithm),
		}
		if kmsMasterKeyId != "" {
			rcDefaultRule.KMSMasterKeyID = aws.String(kmsMasterKeyId)
		}
		rcRule := &s3.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: rcDefaultRule,
		}

		rules = append(rules, rcRule)
	}

	rc.Rules = rules
	i := &s3.PutBucketEncryptionInput{
		Bucket:                            aws.String(bucket),
		ServerSideEncryptionConfiguration: rc,
	}
	log.Printf("[DEBUG] S3 put bucket replication configuration: %#v", i)

	_, err := retryFlakyS3Responses(func() (interface{}, error) {
		return s3conn.PutBucketEncryption(i)
	})
	if err != nil {
		return fmt.Errorf("error putting S3 server side encryption configuration: %s", err)
	}

	return nil
}

func flattenGrants(ap *s3.GetBucketAclOutput) []interface{} {
	//if ACL grants contains bucket owner FULL_CONTROL only - it is default "private" acl
	if len(ap.Grants) == 1 && aws.StringValue(ap.Grants[0].Grantee.ID) == aws.StringValue(ap.Owner.ID) &&
		aws.StringValue(ap.Grants[0].Permission) == s3.PermissionFullControl {
		return nil
	}

	getGrant := func(grants []interface{}, grantee map[string]interface{}) (interface{}, bool) {
		for _, pg := range grants {
			pgt := pg.(map[string]interface{})
			if pgt["type"] == grantee["type"] && pgt["id"] == grantee["id"] && pgt["uri"] == grantee["uri"] &&
				pgt["permissions"].(*schema.Set).Len() > 0 {
				return pg, true
			}
		}
		return nil, false
	}

	grants := make([]interface{}, 0, len(ap.Grants))
	for _, granteeObject := range ap.Grants {
		grantee := make(map[string]interface{})
		grantee["type"] = aws.StringValue(granteeObject.Grantee.Type)

		if granteeObject.Grantee.ID != nil {
			grantee["id"] = aws.StringValue(granteeObject.Grantee.ID)
		}
		if granteeObject.Grantee.URI != nil {
			grantee["uri"] = aws.StringValue(granteeObject.Grantee.URI)
		}
		if pg, ok := getGrant(grants, grantee); ok {
			pg.(map[string]interface{})["permissions"].(*schema.Set).Add(aws.StringValue(granteeObject.Permission))
		} else {
			grantee["permissions"] = schema.NewSet(schema.HashString, []interface{}{aws.StringValue(granteeObject.Permission)})
			grants = append(grants, grantee)
		}
	}

	return grants
}

func flattenS3ServerSideEncryptionConfiguration(c *s3.ServerSideEncryptionConfiguration) []map[string]interface{} {
	var encryptionConfiguration []map[string]interface{}
	rules := make([]interface{}, 0, len(c.Rules))
	for _, v := range c.Rules {
		if v.ApplyServerSideEncryptionByDefault != nil {
			r := make(map[string]interface{})
			d := make(map[string]interface{})
			d["kms_master_key_id"] = aws.StringValue(v.ApplyServerSideEncryptionByDefault.KMSMasterKeyID)
			d["sse_algorithm"] = aws.StringValue(v.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
			r["apply_server_side_encryption_by_default"] = []map[string]interface{}{d}
			rules = append(rules, r)
		}
	}
	encryptionConfiguration = append(encryptionConfiguration, map[string]interface{}{
		"rule": rules,
	})
	return encryptionConfiguration
}

func validateBucketPermissions(permissions []interface{}) error {
	var (
		fullControl     bool
		permissionRead  bool
		permissionWrite bool
	)

	for _, p := range permissions {
		s := p.(string)
		switch s {
		case s3.PermissionFullControl:
			fullControl = true
		case s3.PermissionRead:
			permissionRead = true
		case s3.PermissionWrite:
			permissionWrite = true
		}
	}

	if fullControl && len(permissions) > 1 {
		return fmt.Errorf("do not use other ACP permissions along with `FULL_CONTROL` permission for Storage Bucket")
	}

	if permissionWrite && !permissionRead {
		return fmt.Errorf("should always provide `READ` permission, when granting `WRITE` for Storage Bucket")
	}

	return nil
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

func normalizeRoutingRules(w []*s3.RoutingRule) (string, error) {
	withNulls, err := json.Marshal(w)
	if err != nil {
		return "", err
	}

	var rules []map[string]interface{}
	if err := json.Unmarshal(withNulls, &rules); err != nil {
		return "", err
	}

	var cleanRules []map[string]interface{}
	for _, rule := range rules {
		cleanRules = append(cleanRules, removeNil(rule))
	}

	withoutNulls, err := json.Marshal(cleanRules)
	if err != nil {
		return "", err
	}

	return string(withoutNulls), nil
}

func removeNil(data map[string]interface{}) map[string]interface{} {
	withoutNil := make(map[string]interface{})

	for k, v := range data {
		if v == nil {
			continue
		}

		switch v := v.(type) {
		case map[string]interface{}:
			withoutNil[k] = removeNil(v)
		default:
			withoutNil[k] = v
		}
	}

	return withoutNil
}

func suppressEquivalentAwsPolicyDiffs(k, old, new string, d *schema.ResourceData) bool {
	equivalent, err := awspolicy.PoliciesAreEquivalent(old, new)
	if err != nil {
		return false
	}

	return equivalent
}
