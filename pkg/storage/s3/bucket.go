package s3

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (c *Client) UpdateBucketACL(ctx context.Context, input *s3.PutBucketAclInput) error {
	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Storage put bucket ACL: %#v", input))

	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketAclWithContext(ctx, input)
	})
	if err != nil {
		return fmt.Errorf("error putting Storage Bucket ACL: %w", err)
	}

	return nil
}

func (c *Client) UpdateBucketGrants(ctx context.Context, bucket string, grants []*s3.Grant) error {
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
	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Storage Bucket: %s, read ACL grants policy: %+v", bucket, acl))

	grantsInput := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		AccessControlPolicy: &s3.AccessControlPolicy{
			Grants: grants,
			Owner:  acl.Owner,
		},
	}

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] Bucket: %s, put Grants: %#v", bucket, grantsInput))
	_, err = RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutBucketAclWithContext(ctx, grantsInput)
	})
	if err != nil {
		return fmt.Errorf("error putting Storage Bucket (%s) ACL: %w", bucket, err)
	}

	return nil
}

func (c *Client) GetBucketACL(ctx context.Context, bucket string) (*s3.GetBucketAclOutput, error) {
	acl, err := RetryLongTermOperations[*s3.GetBucketAclOutput](
		ctx,
		func() (*s3.GetBucketAclOutput, error) {
			return c.s3.GetBucketAclWithContext(ctx, &s3.GetBucketAclInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting Storage Bucket (%s) ACL: %w", bucket, err)
	}

	return acl, nil
}

func (c *Client) UpdateBucketPolicy(ctx context.Context, bucket, policy string) error {
	if policy == "" {
		tflog.Debug(ctx, fmt.Sprintf("[DEBUG] S3 bucket: %s, delete policy", bucket))

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

	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] S3 bucket: %s, update policy: %v", bucket, policy))

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

func (c *Client) GetBucketPolicy(ctx context.Context, bucket string) (string, error) {
	pol, err := RetryLongTermOperations[*s3.GetBucketPolicyOutput](
		ctx,
		func() (*s3.GetBucketPolicyOutput, error) {
			return c.s3.GetBucketPolicyWithContext(ctx, &s3.GetBucketPolicyInput{
				Bucket: aws.String(bucket),
			})
		},
	)
	tflog.Debug(ctx, fmt.Sprintf("[DEBUG] S3 bucket: %s, read policy: %v", bucket, pol))
	if err != nil {
		if IsErr(err, NoSuchBucketPolicy) {
			return "", nil
		}
		if IsErr(err, AccessDenied) {
			tflog.Debug(ctx, fmt.Sprintf("[WARN] Got an error while trying to read Storage Bucket (%s) Policy: %s", bucket, err))
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
