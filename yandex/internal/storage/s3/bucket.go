package s3

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/hashcode"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func (c *Client) CreateBucket(ctx context.Context, bucket string, acl BucketACL) error {
	_, err := RetryLongTermOperations[*s3.CreateBucketOutput](ctx, func() (*s3.CreateBucketOutput, error) {
		output, err := c.s3.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
			ACL:    aws.String(string(acl)),
		})
		if err != nil {
			return nil, err
		}
		return output, nil
	})
	if err != nil {
		if IsErr(err, BadRequest) {
			description := "This is usually due to the absence or inability to identify folder_id in which the bucket will be created." +
				" This is possible when creating a bucket using UserAccount IAM token"
			return fmt.Errorf("failed to create bucket. %s: %w", description, err)
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func (c *Client) UpdateBucketPolicy(ctx context.Context, bucket, policy string) error {
	if policy == "" {
		_, err := RetryLongTermOperations[*s3.DeleteBucketPolicyOutput](
			ctx,
			func() (*s3.DeleteBucketPolicyOutput, error) {
				return c.s3.DeleteBucketPolicyWithContext(ctx, &s3.DeleteBucketPolicyInput{
					Bucket: aws.String(bucket),
				})
			},
		)
		if err != nil {
			return fmt.Errorf("failed to delete policy: %w", err)
		}
		return nil
	}

	_, err := RetryOnCodes[*s3.PutBucketPolicyOutput](
		ctx,
		[]ErrCode{MalformedPolicy, NoSuchBucket},
		func() (*s3.PutBucketPolicyOutput, error) {
			output, err := c.s3.PutBucketPolicyWithContext(ctx, &s3.PutBucketPolicyInput{
				Bucket: aws.String(bucket),
				Policy: aws.String(policy),
			})
			if err != nil {
				return nil, err
			}
			return output, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to update policy: %w", err)
	}

	return nil
}

type CORSRule struct {
	MaxAgeSeconds  int
	AllowedHeaders []*string
	AllowedMethods []*string
	AllowedOrigins []*string
	ExposeHeaders  []*string
}

func NewCORSRules(raw []interface{}) []CORSRule {
	rules := make([]CORSRule, 0, len(raw))

	for _, cors := range raw {
		corsMap := cors.(map[string]interface{})
		c := CORSRule{}
		for k, v := range corsMap {
			if k == "max_age_seconds" {
				c.MaxAgeSeconds = v.(int)
			} else {
				vSlice := make([]*string, len(v.([]interface{})))
				for i, vv := range v.([]interface{}) {
					var value string
					if str, ok := vv.(string); ok {
						value = str
					}
					vSlice[i] = &value
				}
				switch k {
				case "allowed_headers":
					c.AllowedHeaders = vSlice
				case "allowed_methods":
					c.AllowedMethods = vSlice
				case "allowed_origins":
					c.AllowedOrigins = vSlice
				case "expose_headers":
					c.ExposeHeaders = vSlice
				}
			}
		}
		rules = append(rules, c)
	}

	return rules
}

func (c *Client) UpdateBucketCORS(ctx context.Context, bucket string, rules []CORSRule) error {
	if len(rules) == 0 {
		// Delete CORS
		log.Printf("[DEBUG] Storage Bucket: %s, delete CORS", bucket)

		_, err := RetryLongTermOperations(ctx, func() (interface{}, error) {
			return c.s3.DeleteBucketCorsWithContext(ctx, &s3.DeleteBucketCorsInput{
				Bucket: aws.String(bucket),
			})
		})
		if err == nil {
			err = c.waitCorsDeleted(ctx, bucket)
		}
		if err != nil {
			return fmt.Errorf("error deleting storage CORS: %w", err)
		}
		return nil
	}

	// Put CORS
	corsRules := make([]*s3.CORSRule, 0, len(rules))
	for _, rule := range rules {
		corsRule := &s3.CORSRule{
			MaxAgeSeconds:  aws.Int64(int64(rule.MaxAgeSeconds)),
			AllowedHeaders: rule.AllowedHeaders,
			AllowedMethods: rule.AllowedMethods,
			AllowedOrigins: rule.AllowedOrigins,
			ExposeHeaders:  rule.ExposeHeaders,
		}
		corsRules = append(corsRules, corsRule)
	}
	corsConfiguration := &s3.CORSConfiguration{
		CORSRules: corsRules,
	}
	corsInput := &s3.PutBucketCorsInput{
		Bucket:            aws.String(bucket),
		CORSConfiguration: corsConfiguration,
	}
	log.Printf("[DEBUG] Storage Bucket: %s, put CORS: %#v", bucket, corsInput)

	_, err := RetryLongTermOperations(ctx, func() (interface{}, error) {
		return c.s3.PutBucketCorsWithContext(ctx, corsInput)
	})
	if err == nil {
		err = c.waitCorsPut(ctx, bucket, corsConfiguration)
	}
	if err != nil {
		return fmt.Errorf("error putting bucket CORS: %w", err)
	}
	return nil
}

func (c *Client) waitCorsDeleted(ctx context.Context, bucket string) error {
	input := &s3.GetBucketCorsInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		_, err := c.s3.GetBucketCorsWithContext(ctx, input)
		if IsErr(err, NoSuchCORSConfiguration) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}

	err := waitConditionStable(check)
	if err != nil {
		return fmt.Errorf("error assuring bucket %q CORS deleted: %w", bucket, err)
	}
	return nil
}

func (c *Client) waitCorsPut(ctx context.Context, bucket string, configuration *s3.CORSConfiguration) error {
	input := &s3.GetBucketCorsInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		output, err := c.s3.GetBucketCorsWithContext(ctx, input)
		if err != nil && !IsErr(err, NoSuchCORSConfiguration) {
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
		return fmt.Errorf("error assuring bucket %q CORS updated: %w", bucket, err)
	}
	return nil
}

type RedirectAllRequestsTo struct {
	HostName string
	Protocol string
}

func newRedirectAllRequestsTo(s string) *RedirectAllRequestsTo {
	redirect, err := url.Parse(s)
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
		return &RedirectAllRequestsTo{
			HostName: redirectHostBuf.String(),
			Protocol: redirect.Scheme,
		}
	}
	return &RedirectAllRequestsTo{HostName: s}
}

type Website struct {
	IndexDocument         string
	ErrorDocument         string
	RedirectAllRequestsTo *RedirectAllRequestsTo
	RoutingRules          []*s3.RoutingRule
}

func NewWebsite(raw []interface{}) (*Website, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var rawWebsite map[string]interface{}
	if raw[0] != nil {
		rawWebsite = raw[0].(map[string]interface{})
	} else {
		rawWebsite = make(map[string]interface{})
	}

	var indexDocument, errorDocument, redirectAllRequestsTo, routingRules string
	if v, ok := rawWebsite["index_document"]; ok {
		indexDocument = v.(string)
	}
	if v, ok := rawWebsite["error_document"]; ok {
		errorDocument = v.(string)
	}
	if v, ok := rawWebsite["redirect_all_requests_to"]; ok {
		redirectAllRequestsTo = v.(string)
	}
	if v, ok := rawWebsite["routing_rules"]; ok {
		routingRules = v.(string)
	}
	if indexDocument == "" && redirectAllRequestsTo == "" {
		return nil, fmt.Errorf("must specify either index_document or redirect_all_requests_to")
	}

	website := &Website{}
	if indexDocument != "" {
		website.IndexDocument = indexDocument
	}
	if errorDocument != "" {
		website.ErrorDocument = errorDocument
	}
	if redirectAllRequestsTo != "" {
		website.RedirectAllRequestsTo = newRedirectAllRequestsTo(redirectAllRequestsTo)
	}
	if routingRules != "" {
		var unmarshalledRules []*s3.RoutingRule
		if err := json.Unmarshal([]byte(routingRules), &unmarshalledRules); err != nil {
			return nil, fmt.Errorf("error unmarshaling routing_rules: %w", err)
		}
		website.RoutingRules = unmarshalledRules
	}

	return website, nil
}

func (c *Client) UpdateBucketWebsite(ctx context.Context, bucket string, website *Website) error {
	if website == nil {
		// Delete website
		deleteInput := &s3.DeleteBucketWebsiteInput{Bucket: aws.String(bucket)}

		log.Printf("[DEBUG] Storage delete bucket website: %#v", deleteInput)

		_, err := RetryLongTermOperations(ctx, func() (any, error) {
			return c.s3.DeleteBucketWebsiteWithContext(ctx, deleteInput)
		})
		if err == nil {
			err = c.waitWebsiteDeleted(ctx, bucket)
		}
		if err != nil {
			return fmt.Errorf("error deleting storage website: %w", err)
		}
		return nil
	}

	// Put website
	websiteConfiguration := &s3.WebsiteConfiguration{
		RoutingRules: website.RoutingRules,
	}
	if website.IndexDocument != "" {
		websiteConfiguration.IndexDocument = &s3.IndexDocument{
			Suffix: aws.String(website.IndexDocument),
		}
	}
	if website.ErrorDocument != "" {
		websiteConfiguration.ErrorDocument = &s3.ErrorDocument{
			Key: aws.String(website.ErrorDocument),
		}
	}
	if website.RedirectAllRequestsTo != nil {
		websiteConfiguration.RedirectAllRequestsTo = &s3.RedirectAllRequestsTo{
			HostName: aws.String(website.RedirectAllRequestsTo.HostName),
			Protocol: aws.String(website.RedirectAllRequestsTo.Protocol),
		}
	}
	putInput := &s3.PutBucketWebsiteInput{
		Bucket:               aws.String(bucket),
		WebsiteConfiguration: websiteConfiguration,
	}

	log.Printf("[DEBUG] Storage put bucket website: %#v", putInput)

	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketWebsiteWithContext(ctx, putInput)
	})
	if err == nil {
		err = c.waitWebsitePut(ctx, bucket, websiteConfiguration)
	}
	if err != nil {
		return fmt.Errorf("error putting storage website: %w", err)
	}

	return nil
}

func (c *Client) waitWebsiteDeleted(ctx context.Context, bucket string) error {
	input := &s3.GetBucketWebsiteInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		_, err := c.s3.GetBucketWebsiteWithContext(ctx, input)
		if IsErr(err, NoSuchWebsiteConfiguration) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return false, nil
	}

	err := waitConditionStable(check)
	if err != nil {
		return fmt.Errorf("error assuring bucket %q website deleted: %w", bucket, err)
	}
	return nil
}

func (c *Client) waitWebsitePut(ctx context.Context, bucket string, configuration *s3.WebsiteConfiguration) error {
	input := &s3.GetBucketWebsiteInput{Bucket: aws.String(bucket)}

	check := func() (bool, error) {
		output, err := c.s3.GetBucketWebsiteWithContext(ctx, input)
		if err != nil && !IsErr(err, NoSuchWebsiteConfiguration) {
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
		return fmt.Errorf("error assuring bucket %q website updated: %w", bucket, err)
	}
	return nil
}

type VersioningStatus string

const (
	VersioningEnabled  VersioningStatus = "Enabled"
	VersioningDisabled VersioningStatus = "Suspended"
)

func NewVersioningStatus(raw []interface{}) VersioningStatus {
	if len(raw) > 0 {
		c := raw[0].(map[string]interface{})
		if c["enabled"].(bool) {
			return VersioningEnabled
		}
	}
	return VersioningDisabled
}

func (c *Client) UpdateBucketVersioning(ctx context.Context, bucket string, status VersioningStatus) error {
	var versioningStatus string
	if status == VersioningEnabled {
		versioningStatus = s3.BucketVersioningStatusEnabled
	} else {
		versioningStatus = s3.BucketVersioningStatusSuspended
	}

	i := &s3.PutBucketVersioningInput{
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{Status: aws.String(versioningStatus)},
	}
	log.Printf("[DEBUG] S3 put bucket versioning: %#v", i)

	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketVersioningWithContext(ctx, i)
	})
	if err != nil {
		return fmt.Errorf("error putting S3 versioning: %w", err)
	}

	return nil
}

type BucketACL string

const (
	BucketACLOwnerFullControl BucketACL = "bucket-owner-full-control"
	BucketACLPublicRead       BucketACL = s3.BucketCannedACLPublicRead
	BucketACLPublicReadWrite  BucketACL = s3.BucketCannedACLPublicReadWrite
	BucketACLAuthRead         BucketACL = s3.BucketCannedACLAuthenticatedRead
	BucketACLPrivate          BucketACL = s3.BucketCannedACLPrivate
)

func (c *Client) UpdateBucketACL(ctx context.Context, bucket string, acl BucketACL) error {
	i := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(string(acl)),
	}
	log.Printf("[DEBUG] Storage put bucket ACL: %#v", i)

	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketAclWithContext(ctx, i)
	})
	if err != nil {
		return fmt.Errorf("error putting Storage Bucket ACL: %w", err)
	}

	return nil
}

type Grantee struct {
	ID   *string
	Type *string
	URI  *string
}

type Grant struct {
	Grantee    *Grantee
	Permission string
}

func NewGrants(raw []interface{}) ([]Grant, error) {
	grants := make([]Grant, 0, len(raw))
	for _, rawGrant := range raw {
		grantMap := rawGrant.(map[string]interface{})
		if err := validateBucketGrant(grantMap); err != nil {
			return nil, err
		}
		id, _ := grantMap["id"].(string)
		type_, _ := grantMap["type"].(string)
		uri, _ := grantMap["uri"].(string)
		permissions := grantMap["permissions"].(*schema.Set).List()
		for _, rawPermission := range permissions {
			grantee := &Grantee{}
			if id != "" {
				grantee.ID = &id
			}
			if type_ != "" {
				grantee.Type = &type_
			}
			if uri != "" {
				grantee.URI = &uri
			}
			grant := Grant{
				Grantee:    grantee,
				Permission: rawPermission.(string),
			}
			grants = append(grants, grant)
		}
	}
	return grants, nil
}

func (c *Client) UpdateBucketGrants(ctx context.Context, bucket string, grants []Grant) error {
	acl, err := RetryLongTermOperations[*s3.GetBucketAclOutput](
		ctx,
		func() (*s3.GetBucketAclOutput, error) {
			return c.s3.GetBucketAclWithContext(ctx, &s3.GetBucketAclInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		return fmt.Errorf("error getting Storage Bucket (%s) ACL: %w", bucket, err)
	}
	log.Printf("[DEBUG] Storage Bucket: %s, read ACL grants policy: %+v", bucket, acl)

	awsGrants := make([]*s3.Grant, 0, len(grants))
	for _, grant := range grants {
		awsGrants = append(awsGrants, &s3.Grant{
			Grantee: &s3.Grantee{
				ID:   grant.Grantee.ID,
				Type: grant.Grantee.Type,
				URI:  grant.Grantee.URI,
			},
			Permission: aws.String(grant.Permission),
		})
	}
	grantsInput := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		AccessControlPolicy: &s3.AccessControlPolicy{
			Grants: awsGrants,
			Owner:  acl.Owner,
		},
	}

	log.Printf("[DEBUG] Bucket: %s, put Grants: %#v", bucket, grantsInput)
	_, err = RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketAclWithContext(ctx, grantsInput)
	})
	if err != nil {
		return fmt.Errorf("error putting Storage Bucket (%s) ACL: %w", bucket, err)
	}

	return nil
}

func validateBucketGrant(grant map[string]interface{}) error {
	switch grant["type"].(string) {
	case s3.TypeCanonicalUser:
		if grant["uri"].(string) != "" {
			return fmt.Errorf("uri can be used only for Group grant type for Storage Bucket")
		}
	case s3.TypeGroup:
		if grant["id"].(string) != "" {
			return fmt.Errorf("id can be used only for CanonicalUser grant type for Storage Bucket")
		}
	}

	permissions := grant["permissions"].(*schema.Set).List()
	return validateBucketPermissions(permissions)
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

type LoggingStatus struct {
	Enabled      bool
	TargetBucket *string
	TargetPrefix *string
}

func NewLoggingStatus(raw []interface{}) LoggingStatus {
	if len(raw) > 0 {
		c := raw[0].(map[string]interface{})
		loggingStatus := LoggingStatus{
			Enabled: true,
		}
		if val, ok := c["target_bucket"]; ok {
			loggingStatus.TargetBucket = aws.String(val.(string))
		}
		if val, ok := c["target_prefix"]; ok {
			loggingStatus.TargetPrefix = aws.String(val.(string))
		}
		return loggingStatus
	}

	return LoggingStatus{
		Enabled: false,
	}
}

func (c *Client) UpdateBucketLogging(ctx context.Context, bucket string, loggingStatus LoggingStatus) error {
	awsLoggingStatus := &s3.BucketLoggingStatus{}
	if loggingStatus.Enabled {
		awsLoggingStatus.LoggingEnabled = &s3.LoggingEnabled{
			TargetBucket: loggingStatus.TargetBucket,
			TargetPrefix: loggingStatus.TargetPrefix,
		}
	}
	i := &s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: awsLoggingStatus,
	}
	log.Printf("[DEBUG] S3 put bucket logging: %#v", i)

	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketLoggingWithContext(ctx, i)
	})
	if err != nil {
		return fmt.Errorf("error putting S3 logging: %w", err)
	}

	return nil
}

func (c *Client) UpdateBucketTags(ctx context.Context, bucket string, tags []Tag) error {
	if len(tags) == 0 {
		// Delete tags
		log.Printf("[DEBUG] Deleting Storage S3 bucket tags")
		request := &s3.DeleteBucketTaggingInput{
			Bucket: aws.String(bucket),
		}
		_, err := RetryLongTermOperations(ctx, func() (any, error) {
			return c.s3.DeleteBucketTaggingWithContext(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("failed to delete bucket tags: %w", err)
		}
		return nil
	}

	// Put tags
	log.Printf("[DEBUG] Updating Storage S3 bucket tags with %v", tags)
	request := &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucket),
		Tagging: &s3.Tagging{
			TagSet: TagsToS3(tags),
		},
	}
	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketTaggingWithContext(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("failed to update bucket tags: %w", err)
	}
	return nil
}

type LifecycleRuleAndOperator struct {
	ObjectSizeGreaterThan *int64
	ObjectSizeLessThan    *int64
	Prefix                *string
	Tags                  []Tag
}

type LifecycleRuleFilter struct {
	Prefix                *string
	ObjectSizeGreaterThan *int64
	ObjectSizeLessThan    *int64
	Tag                   *Tag
	And                   *LifecycleRuleAndOperator
}

type LifecycleExpiration struct {
	Date                      *time.Time
	Days                      *int64
	ExpiredObjectDeleteMarker *bool
}

type LifecycleAbortIncompleteMultipartUpload struct {
	DaysAfterInitiation *int64
}

type LifecycleNoncurrentVersionExpiration struct {
	NoncurrentDays *int64
}

type LifecycleTransition struct {
	Date         *time.Time
	Days         *int64
	StorageClass *string
}

type LifecycleNoncurrentVersionTransition struct {
	NoncurrentDays *int64
	StorageClass   *string
}

type LifecycleRule struct {
	ID                             *string
	Status                         *string
	Filter                         *LifecycleRuleFilter
	Expiration                     *LifecycleExpiration
	AbortIncompleteMultipartUpload *LifecycleAbortIncompleteMultipartUpload
	NoncurrentVersionExpiration    *LifecycleNoncurrentVersionExpiration
	Transitions                    []LifecycleTransition
	NoncurrentVersionTransitions   []LifecycleNoncurrentVersionTransition
}

func NewLifecycleRules(raw []interface{}, d *schema.ResourceData) ([]LifecycleRule, error) {
	rules := make([]LifecycleRule, 0, len(raw))

	for i, lifecycleRule := range raw {
		r := lifecycleRule.(map[string]interface{})

		rule := LifecycleRule{}

		// Filter
		filter := &LifecycleRuleFilter{}
		if prefix, ok := r["prefix"].(string); ok && prefix != "" {
			filter.Prefix = &prefix
		}

		if objectSize, ok := d.GetOk(fmt.Sprintf("lifecycle_rule.%d.filter.0.object_size_greater_than", i)); ok {
			if objectSizeInt, ok := objectSize.(int); ok && objectSizeInt >= 0 {
				filter.ObjectSizeGreaterThan = aws.Int64(int64(objectSizeInt))
			}
		}
		if objectSize, ok := d.GetOk(fmt.Sprintf("lifecycle_rule.%d.filter.0.object_size_less_than", i)); ok {
			if objectSizeInt, ok := objectSize.(int); ok && objectSizeInt >= 1 {
				filter.ObjectSizeLessThan = aws.Int64(int64(objectSizeInt))
			}
		}

		if prefix, ok := d.GetOk(fmt.Sprintf("lifecycle_rule.%d.filter.0.prefix", i)); ok {
			filter.Prefix = aws.String(prefix.(string))
		}

		tag := d.Get(fmt.Sprintf("lifecycle_rule.%d.filter.0.tag", i)).([]interface{})
		if len(tag) > 0 && tag[0] != nil {
			if tagFilter := newTag(tag[0]); tagFilter != nil {
				filter.Tag = tagFilter
			}
		}

		andOperator := d.Get(fmt.Sprintf("lifecycle_rule.%d.filter.0.and", i)).([]interface{})
		if len(andOperator) > 0 && andOperator[0] != nil {
			and := &LifecycleRuleAndOperator{}
			el := andOperator[0].(map[string]interface{})
			if objectSize, ok := el["object_size_greater_than"].(int); ok && objectSize >= 0 {
				and.ObjectSizeGreaterThan = aws.Int64(int64(objectSize))
			}
			if objectSize, ok := el["object_size_less_than"].(int); ok && objectSize >= 1 {
				and.ObjectSizeLessThan = aws.Int64(int64(objectSize))
			}
			if prefix, ok := el["prefix"].(string); ok {
				and.Prefix = &prefix
			}
			if tags, ok := el["tags"].(map[string]interface{}); ok && len(tags) > 0 {
				and.Tags = NewTags(tags)
			}
			filter.And = and
		}

		if filter.And == nil && filter.Tag == nil && filter.Prefix == nil {
			// For backward compatibility set "" to prefix in case any of And, Tag, Prefix is empty
			filter.Prefix = aws.String("")
		}

		rule.Filter = filter

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
			rule.AbortIncompleteMultipartUpload = &LifecycleAbortIncompleteMultipartUpload{
				DaysAfterInitiation: aws.Int64(int64(val)),
			}
		}

		// Expiration
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).([]interface{})
		if len(expiration) > 0 && expiration[0] != nil {
			e := expiration[0].(map[string]interface{})
			i := &LifecycleExpiration{}
			if val, ok := e["date"].(string); ok && val != "" {
				t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", val))
				if err != nil {
					return nil, fmt.Errorf("error Parsing AWS S3 Bucket Lifecycle Expiration Date: %w", err)
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
		ncExpiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_expiration", i)).([]interface{})
		if len(ncExpiration) > 0 && ncExpiration[0] != nil {
			e := ncExpiration[0].(map[string]interface{})

			if val, ok := e["days"].(int); ok && val > 0 {
				rule.NoncurrentVersionExpiration = &LifecycleNoncurrentVersionExpiration{
					NoncurrentDays: aws.Int64(int64(val)),
				}
			}
		}

		// Transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transition", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			rule.Transitions = make([]LifecycleTransition, 0, len(transitions))
			for _, transition := range transitions {
				transition := transition.(map[string]interface{})
				i := LifecycleTransition{}
				if val, ok := transition["date"].(string); ok && val != "" {
					t, err := time.Parse(time.RFC3339, fmt.Sprintf("%sT00:00:00Z", val))
					if err != nil {
						return nil, fmt.Errorf("error Parsing AWS S3 Bucket Lifecycle Expiration Date: %w", err)
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
		ncTransitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.noncurrent_version_transition", i)).(*schema.Set).List()
		if len(ncTransitions) > 0 {
			rule.NoncurrentVersionTransitions = make([]LifecycleNoncurrentVersionTransition, 0, len(ncTransitions))
			for _, transition := range ncTransitions {
				transition := transition.(map[string]interface{})
				i := LifecycleNoncurrentVersionTransition{}
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
			rule.Expiration = &LifecycleExpiration{ExpiredObjectDeleteMarker: aws.Bool(false)}
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func (c *Client) UpdateBucketLifecycle(ctx context.Context, bucket string, rules []LifecycleRule) error {
	if len(rules) == 0 {
		// Delete lifecycle
		i := &s3.DeleteBucketLifecycleInput{
			Bucket: aws.String(bucket),
		}
		_, err := c.s3.DeleteBucketLifecycleWithContext(ctx, i)
		if err != nil {
			return fmt.Errorf("error removing S3 lifecycle: %w", err)
		}
		return nil
	}

	// Put lifecycle
	awsRules := make([]*s3.LifecycleRule, 0, len(rules))
	for _, rule := range rules {
		awsRule := &s3.LifecycleRule{
			ID:     rule.ID,
			Status: rule.Status,
		}
		if rule.Filter != nil {
			awsRule.Filter = &s3.LifecycleRuleFilter{
				Prefix:                rule.Filter.Prefix,
				ObjectSizeLessThan:    rule.Filter.ObjectSizeLessThan,
				ObjectSizeGreaterThan: rule.Filter.ObjectSizeGreaterThan,
			}
			if rule.Filter.Tag != nil {
				awsRule.Filter.Tag = &s3.Tag{
					Key:   &rule.Filter.Tag.Key,
					Value: &rule.Filter.Tag.Value,
				}
			}
			if rule.Filter.And != nil {
				awsRule.Filter.And = &s3.LifecycleRuleAndOperator{
					ObjectSizeGreaterThan: rule.Filter.And.ObjectSizeGreaterThan,
					ObjectSizeLessThan:    rule.Filter.And.ObjectSizeLessThan,
					Prefix:                rule.Filter.And.Prefix,
				}
				awsRule.Filter.And.Tags = TagsToS3(rule.Filter.And.Tags)
			}
		}
		if rule.Expiration != nil {
			awsRule.Expiration = &s3.LifecycleExpiration{
				Date:                      rule.Expiration.Date,
				Days:                      rule.Expiration.Days,
				ExpiredObjectDeleteMarker: rule.Expiration.ExpiredObjectDeleteMarker,
			}
		}
		if rule.AbortIncompleteMultipartUpload != nil {
			awsRule.AbortIncompleteMultipartUpload = &s3.AbortIncompleteMultipartUpload{
				DaysAfterInitiation: rule.AbortIncompleteMultipartUpload.DaysAfterInitiation,
			}
		}
		if rule.NoncurrentVersionExpiration != nil {
			awsRule.NoncurrentVersionExpiration = &s3.NoncurrentVersionExpiration{
				NoncurrentDays: rule.NoncurrentVersionExpiration.NoncurrentDays,
			}
		}
		if rule.Transitions != nil {
			awsRule.Transitions = make([]*s3.Transition, 0, len(rule.Transitions))
			for _, transition := range rule.Transitions {
				awsRule.Transitions = append(awsRule.Transitions, &s3.Transition{
					Date:         transition.Date,
					Days:         transition.Days,
					StorageClass: transition.StorageClass,
				})
			}
		}
		if rule.NoncurrentVersionTransitions != nil {
			awsRule.NoncurrentVersionTransitions = make(
				[]*s3.NoncurrentVersionTransition,
				0,
				len(rule.NoncurrentVersionTransitions),
			)
			for _, transition := range rule.NoncurrentVersionTransitions {
				awsRule.NoncurrentVersionTransitions = append(
					awsRule.NoncurrentVersionTransitions,
					&s3.NoncurrentVersionTransition{
						NoncurrentDays: transition.NoncurrentDays,
						StorageClass:   transition.StorageClass,
					},
				)
			}
		}
		awsRules = append(awsRules, awsRule)
	}
	i := &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucket),
		LifecycleConfiguration: &s3.BucketLifecycleConfiguration{
			Rules: awsRules,
		},
	}
	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketLifecycleConfigurationWithContext(ctx, i)
	})
	if err != nil {
		return fmt.Errorf("error putting S3 lifecycle: %w", err)
	}

	return nil
}

type ServerSideEncryptionRule struct {
	SSEAlgorithm   string
	KMSMasterKeyID *string
}

func NewServerSideEncryptionRules(raw []interface{}) []ServerSideEncryptionRule {
	rules := make([]ServerSideEncryptionRule, 0, len(raw))
	if len(raw) == 0 {
		return rules
	}

	c := raw[0].(map[string]interface{})
	rcRules := c["rule"].([]interface{})
	for _, v := range rcRules {
		rr := v.(map[string]interface{})
		rrDefault := rr["apply_server_side_encryption_by_default"].([]interface{})
		sseAlgorithm := rrDefault[0].(map[string]interface{})["sse_algorithm"].(string)
		kmsMasterKeyId := rrDefault[0].(map[string]interface{})["kms_master_key_id"].(string)
		rule := ServerSideEncryptionRule{
			SSEAlgorithm: sseAlgorithm,
		}
		if kmsMasterKeyId != "" {
			rule.KMSMasterKeyID = &kmsMasterKeyId
		}
		rules = append(rules, rule)
	}
	return rules
}

func (c *Client) UpdateBucketServerSideEncryption(
	ctx context.Context,
	bucket string,
	rules []ServerSideEncryptionRule,
) error {
	if len(rules) == 0 {
		// Delete server side encryption
		log.Printf("[DEBUG] Delete server side encryption configuration rules: %#v", rules)
		i := &s3.DeleteBucketEncryptionInput{
			Bucket: aws.String(bucket),
		}

		_, err := c.s3.DeleteBucketEncryptionWithContext(ctx, i)
		if err != nil {
			return fmt.Errorf("error removing S3 bucket server side encryption: %w", err)
		}
		return nil
	}

	// Put server side encryption
	awsRules := make([]*s3.ServerSideEncryptionRule, 0, len(rules))
	for _, rule := range rules {
		awsRule := &s3.ServerSideEncryptionRule{
			ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
				SSEAlgorithm:   aws.String(rule.SSEAlgorithm),
				KMSMasterKeyID: rule.KMSMasterKeyID,
			},
		}
		awsRules = append(awsRules, awsRule)
	}
	i := &s3.PutBucketEncryptionInput{
		Bucket:                            aws.String(bucket),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{Rules: awsRules},
	}
	log.Printf("[DEBUG] S3 put bucket server side encryption configuration: %#v", i)

	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketEncryptionWithContext(ctx, i)
	})
	if err != nil {
		return fmt.Errorf("error putting S3 server side encryption configuration: %w", err)
	}

	return nil
}

type ObjectLockRule struct {
	Mode  string
	Days  *int64
	Years *int64
}

type ObjectLock struct {
	Enabled bool
	Rule    *ObjectLockRule
}

func NewObjectLock(raw []interface{}) ObjectLock {
	if len(raw) == 0 {
		return ObjectLock{
			Enabled: false,
		}
	}

	out := ObjectLock{
		Enabled: true,
	}
	config := raw[0].(map[string]interface{})
	rs := config["rule"].([]interface{})
	if len(rs) == 0 {
		return out
	}

	rawRule := rs[0].(map[string]interface{})
	drs := rawRule["default_retention"].([]interface{})
	retention := drs[0].(map[string]interface{})

	rule := ObjectLockRule{
		Mode: retention["mode"].(string),
	}
	if days, ok := retention["days"].(int); ok && days > 0 {
		rule.Days = aws.Int64(int64(days))
	}
	if years, ok := retention["years"].(int); ok && years > 0 {
		rule.Years = aws.Int64(int64(years))
	}
	out.Rule = &rule
	return out
}

func (c *Client) UpdateBucketObjectLock(ctx context.Context, bucket string, lock ObjectLock) error {
	olc := &s3.ObjectLockConfiguration{}
	if lock.Enabled {
		olc.ObjectLockEnabled = aws.String(s3.ObjectLockEnabledEnabled)
		if lock.Rule != nil {
			olc.Rule = &s3.ObjectLockRule{
				DefaultRetention: &s3.DefaultRetention{
					Mode:  aws.String(lock.Rule.Mode),
					Days:  lock.Rule.Days,
					Years: lock.Rule.Years,
				},
			}
		}
	}
	i := &s3.PutObjectLockConfigurationInput{
		Bucket:                  aws.String(bucket),
		ObjectLockConfiguration: olc,
	}
	log.Printf("[DEBUG] S3 put bucket object lock configuration: %#v", i)

	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutObjectLockConfigurationWithContext(ctx, i)
	})
	if err != nil {
		return fmt.Errorf("error putting S3 object lock configuration: %w", err)
	}

	return nil
}

func (c *Client) DeleteBucket(ctx context.Context, bucket string, force bool) error {
	_, err := RetryOnCodes(ctx, []ErrCode{AccessDenied, Forbidden}, func() (any, error) {
		return c.s3.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
	})
	if err != nil {
		if IsErr(err, NoSuchBucket) {
			return nil
		}
		if IsErr(err, BucketNotEmpty) {
			if force {
				// bucket may have things delete them
				log.Printf("[DEBUG] Storage Bucket attempting to forceDestroy %+v", err)

				resp, err := c.s3.ListObjectVersionsWithContext(
					ctx,
					&s3.ListObjectVersionsInput{
						Bucket: aws.String(bucket),
					},
				)

				if err != nil {
					return fmt.Errorf("error listing Storage Bucket object versions: %w", err)
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

				_, err = c.s3.DeleteObjectsWithContext(ctx, params)

				if err != nil {
					return fmt.Errorf("error force_destroy deleting Storage Bucket (%s): %w", bucket, err)
				}

				// this line recurses until all objects are deleted or an error is returned
				return c.DeleteBucket(ctx, bucket, force)
			}
		}

		return fmt.Errorf("error deleting Storage Bucket (%s): %w", bucket, err)
	}

	req := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}
	err = waitConditionStable(func() (bool, error) {
		_, err := c.s3.HeadBucketWithContext(ctx, req)
		var awsError awserr.RequestFailure
		if errors.As(err, &awsError) && awsError.StatusCode() == 404 {
			return true, nil
		}
		return false, err
	})
	if err != nil {
		return fmt.Errorf("error waiting for bucket %q to be deleted: %w", bucket, err)
	}

	return nil
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

var ErrBucketNotFound = errors.New("bucket not found")

type WebsiteInfo struct {
	RawData  []map[string]interface{}
	Domain   string
	Endpoint string
}

type Bucket struct {
	DomainName string
	Policy     string
	CORSRules  []map[string]interface{}
	Website    *WebsiteInfo
	Grants     []interface{}
	Versioning []map[string]interface{}
	ObjectLock []map[string]interface{}
	Logging    []map[string]interface{}
	Lifecycle  []map[string]interface{}
	Encryption []map[string]interface{}
	Tags       []Tag
}

func (c *Client) GetBucket(ctx context.Context, bucket, endpoint, acl string) (*Bucket, error) {
	resp, err := RetryLongTermOperations[*s3.HeadBucketOutput](ctx, func() (*s3.HeadBucketOutput, error) {
		return c.s3.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(bucket),
		})
	})
	if err != nil {
		var awsError awserr.RequestFailure
		if errors.As(err, &awsError) && awsError.StatusCode() == 404 {
			return nil, ErrBucketNotFound
		}
		return nil, fmt.Errorf("error reading Storage Bucket (%s): %w", bucket, err)
	}
	log.Printf("[DEBUG] Storage head bucket output: %#v", resp)

	domainName, err := c.getBucketDomainName(bucket, endpoint)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket domain name: %w", err)
	}
	policy, err := c.getBucketPolicy(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket policy: %w", err)
	}
	corsRules, err := c.getCORSRules(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket CORS rules: %w", err)
	}
	website, err := c.getBucketWebsite(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket website: %w", err)
	}
	grants, err := c.getBucketGrants(ctx, bucket, acl)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket grants: %w", err)
	}
	versioning, err := c.getBucketVersioning(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket versioning: %w", err)
	}
	objectLock, err := c.getBucketObjectLock(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket object lock: %w", err)
	}
	logging, err := c.getBucketLogging(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket logging: %w", err)
	}
	lifecycle, err := c.getBucketLifecycle(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket lifecycle: %w", err)
	}
	encryption, err := c.getBucketServerSideEncryption(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket server side encryption: %w", err)
	}
	tags, err := c.getBucketTags(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("error getting bucket tags: %w", err)
	}

	return &Bucket{
		DomainName: domainName,
		Policy:     policy,
		CORSRules:  corsRules,
		Website:    website,
		Grants:     grants,
		Versioning: versioning,
		ObjectLock: objectLock,
		Logging:    logging,
		Lifecycle:  lifecycle,
		Encryption: encryption,
		Tags:       tags,
	}, nil
}

func (c *Client) getBucketDomainName(bucket string, endpointURL string) (string, error) {
	// Without a scheme the url will not be parsed as we expect
	// See https://github.com/golang/go/issues/19779
	if !strings.Contains(endpointURL, "//") {
		endpointURL = "//" + endpointURL
	}
	parse, err := url.Parse(endpointURL)
	if err != nil {
		return "", fmt.Errorf("error parsing endpoint URL: %w", err)
	}
	return fmt.Sprintf("%s.%s", bucket, parse.Hostname()), nil
}

func (c *Client) getBucketPolicy(ctx context.Context, bucket string) (string, error) {
	pol, err := RetryLongTermOperations[*s3.GetBucketPolicyOutput](
		ctx,
		func() (*s3.GetBucketPolicyOutput, error) {
			return c.s3.GetBucketPolicyWithContext(ctx, &s3.GetBucketPolicyInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	log.Printf("[DEBUG] S3 bucket: %s, read policy: %v", bucket, pol)
	if err != nil {
		if IsErr(err, NoSuchBucketPolicy) {
			return "", nil
		}
		if IsErr(err, AccessDenied) {
			log.Printf("[WARN] Got an error while trying to read Storage Bucket (%s) Policy: %s", bucket, err)
			return "", nil
		}
		return "", fmt.Errorf("error getting current policy: %w", err)
	}
	if v := pol.Policy; v != nil {
		policy, err := normalizeJsonString(aws.StringValue(v))
		if err != nil {
			return "", fmt.Errorf("policy contains an invalid JSON: %w", err)
		}
		return policy, nil
	}
	return "", nil
}

func (c *Client) getCORSRules(ctx context.Context, bucket string) ([]map[string]interface{}, error) {
	resp, err := RetryLongTermOperations[*s3.GetBucketCorsOutput](ctx, func() (*s3.GetBucketCorsOutput, error) {
		return c.s3.GetBucketCorsWithContext(ctx, &s3.GetBucketCorsInput{
			Bucket: aws.String(bucket),
		})
	})
	if err != nil {
		if IsErr(err, NoSuchCORSConfiguration) {
			return nil, nil
		}
		return nil, fmt.Errorf("error reading Storage Bucket (%s) CORS configuration: %w", bucket, err)
	}
	log.Printf("[DEBUG] Storage get bucket CORS output: %#v", resp)

	corsRules := make([]map[string]interface{}, 0)
	if len(resp.CORSRules) > 0 {
		corsRules = make([]map[string]interface{}, 0, len(resp.CORSRules))
		for _, ruleObject := range resp.CORSRules {
			rule := make(map[string]interface{})
			rule["allowed_headers"] = flattenStringList(ruleObject.AllowedHeaders)
			rule["allowed_methods"] = flattenStringList(ruleObject.AllowedMethods)
			rule["allowed_origins"] = flattenStringList(ruleObject.AllowedOrigins)
			if ruleObject.ExposeHeaders != nil {
				rule["expose_headers"] = flattenStringList(ruleObject.ExposeHeaders)
			}
			if ruleObject.MaxAgeSeconds != nil {
				rule["max_age_seconds"] = int(*ruleObject.MaxAgeSeconds)
			}
			corsRules = append(corsRules, rule)
		}
	}
	return corsRules, nil
}

const websiteDomainURL = "website.yandexcloud.net"

func (c *Client) getBucketWebsite(ctx context.Context, bucket string) (*WebsiteInfo, error) {
	rawData, err := c.getBucketWebsiteRawData(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if len(rawData) == 0 {
		return nil, nil
	}

	return &WebsiteInfo{
		RawData:  rawData,
		Endpoint: fmt.Sprintf("%s.%s", bucket, websiteDomainURL),
		Domain:   websiteDomainURL,
	}, nil
}

func (c *Client) getBucketWebsiteRawData(ctx context.Context, bucket string) ([]map[string]interface{}, error) {
	ws, err := RetryLongTermOperations[*s3.GetBucketWebsiteOutput](
		ctx,
		func() (*s3.GetBucketWebsiteOutput, error) {
			return c.s3.GetBucketWebsiteWithContext(ctx, &s3.GetBucketWebsiteInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		if IsErr(err, NoSuchWebsiteConfiguration) || IsErr(err, NotImplemented) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting Storage Bucket website configuration: %w", err)
	}
	log.Printf("[DEBUG] Storage get bucket website output: %#v", ws)

	websites := make([]map[string]interface{}, 0, 1)
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
			var host, path, query string
			if parsedHostName, err := url.Parse(aws.StringValue(v.HostName)); err != nil {
				host = aws.StringValue(v.HostName)
			} else {
				host = parsedHostName.Host
				path = parsedHostName.Path
				query = parsedHostName.RawQuery
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
			return nil, fmt.Errorf("error while marshaling routing rules: %w", err)
		}
		w["routing_rules"] = rr
	}

	// We have special handling for the website configuration,
	// so only add the configuration if there is any
	if len(w) > 0 {
		websites = append(websites, w)
	}

	return websites, nil
}

func (c *Client) getBucketGrants(ctx context.Context, bucket, acl string) ([]interface{}, error) {
	if acl != "" {
		return nil, nil
	}

	apResponse, err := RetryLongTermOperations[*s3.GetBucketAclOutput](
		ctx,
		func() (*s3.GetBucketAclOutput, error) {
			return c.s3.GetBucketAclWithContext(ctx, &s3.GetBucketAclInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		// Ignore access denied error, when reading ACL for bucket.
		if IsErr(err, AccessDenied) || IsErr(err, Forbidden) {
			log.Printf("[WARN] Got an error while trying to read Storage Bucket (%s) ACL: %s", bucket, err)
			return nil, nil
		}
		return nil, fmt.Errorf("error getting Storage Bucket (%s) ACL: %w", bucket, err)
	}

	log.Printf("[DEBUG] getting storage: %s, read ACL grants policy: %+v", bucket, apResponse)
	grants := flattenGrants(apResponse)
	return grants, nil
}

func (c *Client) getBucketVersioning(ctx context.Context, bucket string) ([]map[string]interface{}, error) {
	versioning, err := RetryLongTermOperations[*s3.GetBucketVersioningOutput](
		ctx,
		func() (*s3.GetBucketVersioningOutput, error) {
			return c.s3.GetBucketVersioningWithContext(ctx, &s3.GetBucketVersioningInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting Storage Bucket (%s) versioning: %w", bucket, err)
	}

	vcl := make([]map[string]interface{}, 0, 1)
	vc := make(map[string]interface{})
	if versioning.Status != nil && aws.StringValue(versioning.Status) == s3.BucketVersioningStatusEnabled {
		vc["enabled"] = true
	} else {
		vc["enabled"] = false
	}

	return append(vcl, vc), nil
}

func (c *Client) getBucketObjectLock(ctx context.Context, bucket string) ([]map[string]interface{}, error) {
	objectLockConfig, err := RetryLongTermOperations[*s3.GetObjectLockConfigurationOutput](
		ctx,
		func() (*s3.GetObjectLockConfigurationOutput, error) {
			return c.s3.GetObjectLockConfigurationWithContext(ctx, &s3.GetObjectLockConfigurationInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		if IsErr(err, ObjectLockConfigurationNotFoundError) || IsErr(err, AccessDenied) {
			log.Printf(
				"[WARN] Got an error while trying to read Storage Bucket (%s) ObjectLockConfigurationt: %s",
				bucket,
				err,
			)
			return nil, nil
		}
		return nil, err
	}
	if objectLockConfig.ObjectLockConfiguration == nil {
		return nil, nil
	}

	log.Printf("[DEBUG] Storage get bucket object lock config output: %#v", objectLockConfig)
	olcl := make([]map[string]interface{}, 0, 1)
	olc := make(map[string]interface{})

	enabled := objectLockConfig.ObjectLockConfiguration.ObjectLockEnabled
	rule := objectLockConfig.ObjectLockConfiguration.Rule

	if aws.StringValue(enabled) != "" {
		olc["object_lock_enabled"] = aws.StringValue(enabled)
	}

	if rule != nil {
		rt := make(map[string]interface{}, 2)
		defaultRetention := rule.DefaultRetention

		rt["mode"] = aws.StringValue(defaultRetention.Mode)
		if defaultRetention.Days != nil {
			rt["days"] = aws.Int64Value(defaultRetention.Days)
		}
		if defaultRetention.Years != nil {
			rt["years"] = aws.Int64Value(defaultRetention.Years)
		}

		dr := make(map[string]interface{})
		dr["default_retention"] = []interface{}{rt}
		olc["rule"] = []interface{}{dr}
	}

	return append(olcl, olc), nil
}

func (c *Client) getBucketLogging(ctx context.Context, bucket string) ([]map[string]interface{}, error) {
	logging, err := RetryLongTermOperations[*s3.GetBucketLoggingOutput](
		ctx,
		func() (*s3.GetBucketLoggingOutput, error) {
			return c.s3.GetBucketLoggingWithContext(ctx, &s3.GetBucketLoggingInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting S3 Bucket logging: %w", err)
	}
	if logging.LoggingEnabled == nil {
		return nil, nil
	}

	lcl := make([]map[string]interface{}, 0, 1)
	v := logging.LoggingEnabled
	lc := make(map[string]interface{})
	if aws.StringValue(v.TargetBucket) != "" {
		lc["target_bucket"] = aws.StringValue(v.TargetBucket)
	}
	if aws.StringValue(v.TargetPrefix) != "" {
		lc["target_prefix"] = aws.StringValue(v.TargetPrefix)
	}
	return append(lcl, lc), nil
}

func (c *Client) getBucketLifecycle(ctx context.Context, bucket string) ([]map[string]interface{}, error) {
	lifecycle, err := RetryLongTermOperations[*s3.GetBucketLifecycleConfigurationOutput](
		ctx,
		func() (*s3.GetBucketLifecycleConfigurationOutput, error) {
			return c.s3.GetBucketLifecycleConfigurationWithContext(ctx, &s3.GetBucketLifecycleConfigurationInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		if IsErr(err, NoSuchLifecycleConfiguration) {
			return nil, nil
		}
		return nil, err
	}
	if len(lifecycle.Rules) == 0 {
		return nil, nil
	}

	lifecycleRules := make([]map[string]interface{}, 0, len(lifecycle.Rules))
	for _, lifecycleRule := range lifecycle.Rules {
		log.Printf("[DEBUG] S3 bucket: %s, read lifecycle rule: %v", bucket, lifecycleRule)
		rule := make(map[string]interface{})

		// ID
		if lifecycleRule.ID != nil && aws.StringValue(lifecycleRule.ID) != "" {
			rule["id"] = aws.StringValue(lifecycleRule.ID)
		}
		filter := lifecycleRule.Filter
		if filter != nil {
			ruleFilter := make([]map[string]interface{}, 0, 1)
			if filter.And != nil {
				and := make(map[string]interface{})
				andList := make([]map[string]interface{}, 0, 1)
				// ObjectSizeGreaterThan
				if filter.And.ObjectSizeGreaterThan != nil {
					and["object_size_greater_than"] = int(aws.Int64Value(filter.And.ObjectSizeGreaterThan))
				}
				// ObjectSizeLessThan
				if filter.And.ObjectSizeLessThan != nil {
					and["object_size_less_than"] = int(aws.Int64Value(filter.And.ObjectSizeLessThan))
				}
				// Prefix
				if filter.And.Prefix != nil && aws.StringValue(filter.And.Prefix) != "" {
					and["prefix"] = aws.StringValue(filter.And.Prefix)
				}
				// Tags
				if len(filter.And.Tags) > 0 {
					if tags := S3TagsToRaw(filter.And.Tags); tags != nil {
						and["tags"] = tags
					}
				}
				ruleFilter = append(ruleFilter, map[string]interface{}{"and": append(andList, and)})
			} else {
				if filter.ObjectSizeGreaterThan != nil {
					// ObjectSizeGreaterThan
					ruleFilter = append(ruleFilter, map[string]interface{}{"object_size_greater_than": int(aws.Int64Value(filter.ObjectSizeGreaterThan))})
				} else if filter.ObjectSizeLessThan != nil {
					// ObjectSizeLessThan
					ruleFilter = append(ruleFilter, map[string]interface{}{"object_size_less_than": int(aws.Int64Value(filter.ObjectSizeLessThan))})
				} else if filter.Prefix != nil && aws.StringValue(filter.Prefix) != "" {
					// Prefix
					ruleFilter = append(ruleFilter, map[string]interface{}{"prefix": aws.StringValue(filter.Prefix)})
				} else if filter.Tag != nil {
					tag := make(map[string]interface{})
					tagList := make([]map[string]interface{}, 0, 1)
					tag["key"] = aws.StringValue(filter.Tag.Key)
					tag["value"] = aws.StringValue(filter.Tag.Value)
					ruleFilter = append(ruleFilter, map[string]interface{}{"tag": append(tagList, tag)})
				}
			}
			rule["filter"] = ruleFilter
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
				rule["abort_incomplete_multipart_upload_days"] = int(
					aws.Int64Value(lifecycleRule.AbortIncompleteMultipartUpload.DaysAfterInitiation),
				)
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
				e["expired_object_delete_marker"] = aws.BoolValue(
					lifecycleRule.Expiration.ExpiredObjectDeleteMarker,
				)
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
			rule["transition"] = schema.NewSet(TransitionHash, transitions)
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
			rule["noncurrent_version_transition"] = schema.NewSet(TransitionHash, transitions)
		}

		lifecycleRules = append(lifecycleRules, rule)
	}

	return lifecycleRules, nil
}

func (c *Client) getBucketServerSideEncryption(ctx context.Context, bucket string) ([]map[string]interface{}, error) {
	encryption, err := RetryLongTermOperations[*s3.GetBucketEncryptionOutput](
		ctx,
		func() (*s3.GetBucketEncryptionOutput, error) {
			return c.s3.GetBucketEncryptionWithContext(ctx, &s3.GetBucketEncryptionInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		if IsErr(err, ServerSideEncryptionConfigurationNotFoundError) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting S3 Bucket encryption: %w", err)
	}
	if encryption.ServerSideEncryptionConfiguration == nil {
		return nil, nil
	}

	return flattenS3ServerSideEncryptionConfiguration(
		encryption.ServerSideEncryptionConfiguration,
	), nil
}

func (c *Client) getBucketTags(ctx context.Context, bucket string) ([]Tag, error) {
	tags, err := RetryLongTermOperations[*s3.GetBucketTaggingOutput](
		ctx,
		func() (*s3.GetBucketTaggingOutput, error) {
			return c.s3.GetBucketTaggingWithContext(ctx, &s3.GetBucketTaggingInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting S3 Bucket tags: %w", err)
	}

	return newTagsFromS3(tags.TagSet), nil
}

func normalizeJsonString(jsonString interface{}) (string, error) {
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

func TransitionHash(v interface{}) int {
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
