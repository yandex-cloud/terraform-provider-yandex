package yandex

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/go-getter/helper/url"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
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

			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
			if ruleObject.AllowedOrigins != nil {
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
			return err
		}
		if err := d.Set("website_domain", websiteEndpoint.Domain); err != nil {
			return err
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

	var indexDocument, errorDocument string
	if v, ok := website["index_document"]; ok {
		indexDocument = v.(string)
	}
	if v, ok := website["error_document"]; ok {
		errorDocument = v.(string)
	}
	if indexDocument == "" {
		return fmt.Errorf("\"index_document\" field must be specified")
	}

	websiteConfiguration := &s3.WebsiteConfiguration{}

	if indexDocument != "" {
		websiteConfiguration.IndexDocument = &s3.IndexDocument{Suffix: aws.String(indexDocument)}
	}

	if errorDocument != "" {
		websiteConfiguration.ErrorDocument = &s3.ErrorDocument{Key: aws.String(errorDocument)}
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
