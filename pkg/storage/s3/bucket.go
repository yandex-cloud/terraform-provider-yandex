package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (c *Client) UpdateBucketACL(ctx context.Context, input *s3.PutBucketAclInput) error {
	log.Printf("[DEBUG] Storage put bucket ACL: %#v", input)

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
	log.Printf("[DEBUG] Storage Bucket: %s, read ACL grants policy: %+v", bucket, acl)

	grantsInput := &s3.PutBucketAclInput{
		Bucket: aws.String(bucket),
		AccessControlPolicy: &s3.AccessControlPolicy{
			Grants: grants,
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
