package s3

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mitchellh/go-homedir"
)

type SourceType string

const (
	SourceTypeFile          SourceType = "file"
	SourceTypeContent       SourceType = "content"
	SourceTypeContentBase64 SourceType = "content_base64"
)

type Source struct {
	Type  SourceType
	Value string
}

func (s *Source) Parse() (io.ReadSeeker, error) {
	var (
		data []byte
		err  error
	)

	switch s.Type {
	case SourceTypeFile:
		path, err := homedir.Expand(s.Value)
		if err != nil {
			return nil, fmt.Errorf("error expanding homedir in source (%s): %w", s.Value, err)
		}
		file, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("error opening storage bucket object source (%s): %w", path, err)
		}
		defer func() {
			err := file.Close()
			if err != nil {
				log.Printf("[WARN] Error closing storage bucket object source (%s): %s", path, err)
			}
		}()
		data, err = io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("error reading from file (%s): %w", path, err)
		}

	case SourceTypeContent:
		data = []byte(s.Value)

	case SourceTypeContentBase64:
		data, err = base64.StdEncoding.DecodeString(s.Value)
		if err != nil {
			return nil, fmt.Errorf("error decoding content_base64: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported source type: %s", s.Type)
	}

	return bytes.NewReader(data), nil
}

type ObjectRetention struct {
	Mode            string
	RetainUntilDate time.Time
}

type CreationData struct {
	Source                    *Source
	Bucket                    string
	Key                       string
	ACL                       string
	ContentType               string
	ObjectLockLegalHoldStatus string
	ObjectRetention           *ObjectRetention
	Tags                      []Tag
}

// CreateObject creates a new object in the bucket with the given key and source.
// It returns true if the object was created, false if it was not created (but no error occurred),
func (c *Client) CreateObject(ctx context.Context, data CreationData) (bool, error) {
	body, err := data.Source.Parse()
	if err != nil {
		return false, fmt.Errorf("error parsing source: %w", err)
	}

	putObjectInput := &s3.PutObjectInput{
		Bucket: aws.String(data.Bucket),
		Key:    aws.String(data.Key),
		ACL:    aws.String(data.ACL),
		Body:   body,
	}

	if data.ContentType != "" {
		putObjectInput.ContentType = aws.String(data.ContentType)
	}
	if data.ObjectLockLegalHoldStatus != "" {
		putObjectInput.SetObjectLockLegalHoldStatus(data.ObjectLockLegalHoldStatus)
	}
	if data.ObjectRetention != nil {
		putObjectInput.SetObjectLockMode(data.ObjectRetention.Mode)
		putObjectInput.SetObjectLockRetainUntilDate(data.ObjectRetention.RetainUntilDate)
	}

	log.Printf("[DEBUG] Sending putObjectInput %s", putObjectInput.String())
	if _, err := c.s3.PutObjectWithContext(ctx, putObjectInput); err != nil {
		return false, fmt.Errorf("error putting object in bucket %q: %w", data.Bucket, err)
	}

	// Use separate request to set tags since it allows to caught
	// NotImplemented error.
	if len(data.Tags) > 0 {
		log.Println("[DEBUG] Trying to set tags for object")
		input := &s3.PutObjectTaggingInput{
			Bucket: aws.String(data.Bucket),
			Key:    aws.String(data.Key),
			Tagging: &s3.Tagging{
				TagSet: TagsToS3(data.Tags),
			},
		}
		if _, err = c.s3.PutObjectTaggingWithContext(ctx, input); err != nil {
			return true, fmt.Errorf("error putting object tags in bucket %q: %w", data.Bucket, err)
		}
	}

	return true, nil
}

var ErrObjectNotFound = errors.New("object not found")

type Object struct {
	Bucket                    string
	Key                       string
	ContentType               *string
	ObjectLockLegalHoldStatus *string
	ObjectRetention           *ObjectRetention
	Tags                      []Tag
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) (*Object, error) {
	resp, err := c.s3.HeadObjectWithContext(
		ctx,
		&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		var awsError awserr.RequestFailure
		if errors.As(err, &awsError) && awsError.StatusCode() == 404 {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("error reading object (%s): %w", key, err)
	}
	log.Printf("[DEBUG] Reading storage object meta: %s", resp)

	object := &Object{
		Bucket:                    bucket,
		Key:                       key,
		ContentType:               resp.ContentType,
		ObjectLockLegalHoldStatus: resp.ObjectLockLegalHoldStatus,
	}
	if resp.ObjectLockMode != nil {
		object.ObjectRetention = &ObjectRetention{
			Mode:            aws.StringValue(resp.ObjectLockMode),
			RetainUntilDate: aws.TimeValue(resp.ObjectLockRetainUntilDate),
		}
	}

	tagsResponse, err := RetryLongTermOperations[*s3.GetObjectTaggingOutput](
		ctx,
		func() (*s3.GetObjectTaggingOutput, error) {
			return c.s3.GetObjectTaggingWithContext(ctx, &s3.GetObjectTaggingInput{
				Bucket:    aws.String(bucket),
				Key:       aws.String(key),
				VersionId: resp.VersionId,
			})
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error getting object tags in bucket %q: %w", bucket, err)
	}

	object.Tags = newTagsFromS3(tagsResponse.TagSet)
	return object, nil
}

func (c *Client) UpdateObjectACL(ctx context.Context, bucket, key, acl string) error {
	_, err := c.s3.PutObjectAclWithContext(ctx, &s3.PutObjectAclInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String(acl),
	})
	if err != nil {
		return fmt.Errorf("error updating object ACL (%s): %w", key, err)
	}
	return nil
}

func (c *Client) UpdateObjectLegalHold(ctx context.Context, bucket, key, status string) error {
	_, err := c.s3.PutObjectLegalHoldWithContext(ctx, &s3.PutObjectLegalHoldInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		LegalHold: &s3.ObjectLockLegalHold{
			Status: aws.String(status),
		},
	})
	if err != nil {
		return fmt.Errorf("error updating object legal hold (%s): %w", key, err)
	}
	return nil
}

func (c *Client) UpdateObjectRetention(
	ctx context.Context,
	bucket, key string,
	retention *ObjectRetention,
) error {
	awsRetention := s3.ObjectLockRetention{}
	if retention != nil {
		awsRetention.Mode = aws.String(retention.Mode)
		awsRetention.RetainUntilDate = aws.Time(retention.RetainUntilDate)
	}

	_, err := c.s3.PutObjectRetentionWithContext(ctx, &s3.PutObjectRetentionInput{
		Bucket:                    aws.String(bucket),
		Key:                       aws.String(key),
		Retention:                 &awsRetention,
		BypassGovernanceRetention: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("error updating object retention (%s): %w", key, err)
	}
	return nil
}

func (c *Client) UpdateObjectTags(ctx context.Context, bucket, key string, tags []Tag) error {
	if len(tags) == 0 {
		log.Printf("[DEBUG] Deleting Storage S3 object tags")
		request := &s3.DeleteObjectTaggingInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}
		_, err := RetryLongTermOperations(ctx, func() (any, error) {
			return c.s3.DeleteObjectTaggingWithContext(ctx, request)
		})
		if err != nil {
			return fmt.Errorf("error deleting object tags in bucket %q: %w", bucket, err)
		}
		return nil
	}

	log.Printf("[DEBUG] Updating Storage S3 object tags with %v", tags)
	request := &s3.PutObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Tagging: &s3.Tagging{
			TagSet: TagsToS3(tags),
		},
	}
	_, err := RetryLongTermOperations(ctx, func() (any, error) {
		return c.s3.PutObjectTaggingWithContext(ctx, request)
	})
	if err != nil {
		return fmt.Errorf("error updating object tags in bucket %q: %w", bucket, err)
	}
	return nil
}

func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	versionOutput, err := c.s3.ListObjectVersionsWithContext(ctx, &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("error getting version id for deliting storage object %q in bucket %s: %w", key, bucket, err)
	}

	_, err = c.s3.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(key),
		VersionId: versionOutput.Versions[0].VersionId,
	})
	if err != nil {
		return fmt.Errorf("error deleting storage object %q in bucket %q: %w ", key, bucket, err)
	}

	return nil
}
