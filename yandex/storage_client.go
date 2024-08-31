package yandex

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultS3Region = "ru-central1"

func getS3ClientByKeys(ctx context.Context, accessKey, secretKey string, c *Config) (*s3.S3, error) {
	if accessKey == "" || secretKey == "" {
		if c.defaultS3Session == nil {
			return nil, fmt.Errorf("failed to get default storage client")
		}

		return newS3Client(ctx, c.defaultS3Session), nil
	}

	newSession, err := newS3Session(c.StorageEndpoint, accessKey, secretKey)
	if err != nil {
		return nil, err
	}

	return newS3Client(ctx, newSession), nil
}

func getS3Client(ctx context.Context, d *schema.ResourceData, c *Config) (*s3.S3, error) {
	ak, sk, err := getS3Keys(d)

	if err != nil {
		return nil, err
	}

	return getS3ClientByKeys(ctx, ak, sk, c)
}

type s3basicError string

func (err s3basicError) Error() string {
	return string(err)
}

const errNoAccessOrSecretKey s3basicError = "both access and secret keys should be specified"

func getS3Keys(b *schema.ResourceData) (accessKey, secretKey string, err error) {
	if b == nil {
		return "", "", nil
	}

	var hasAccessKey, hasSecretKey bool
	var v interface{}

	if v, hasAccessKey = b.GetOk("access_key"); hasAccessKey {
		accessKey = v.(string)
	}

	if v, hasSecretKey = b.GetOk("secret_key"); hasSecretKey {
		secretKey = v.(string)
	}

	switch {
	case hasAccessKey != hasSecretKey:
		err = errNoAccessOrSecretKey
	case hasAccessKey && (accessKey == "" || secretKey == ""):
		err = errNoAccessOrSecretKey
	}
	if err != nil {
		return "", "", err
	}

	return accessKey, secretKey, nil
}

func newS3Session(url, accessKey, secretKey string) (*session.Session, error) {
	if url == "" {
		return nil, fmt.Errorf("failed to create storage client, endpoint url is not specified")
	}

	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(url),
		Region:      aws.String(defaultS3Region),
	}

	newSession, err := session.NewSession(s3Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return newSession, nil
}

func newS3Client(ctx context.Context, session *session.Session) *s3.S3 {
	additionalS3Config := &aws.Config{
		LogLevel: aws.LogLevel(aws.LogDebug),
		Logger: aws.LoggerFunc(func(args ...any) {
			tflog.Debug(ctx, fmt.Sprint(args...))
		}),
	}

	return s3.New(session, additionalS3Config)
}

const retryTimeout = 5 * time.Second

type AwsCode string

const (
	AwsNoSuchBucket                                   AwsCode = s3.ErrCodeNoSuchBucket
	AwsAccessDenied                                   AwsCode = "AccessDenied"
	AwsForbidden                                      AwsCode = "Forbidden"
	AwsMalformedPolicy                                AwsCode = "MalformedPolicy"
	AwsOperationAborted                               AwsCode = "OperationAborted"
	AwsBucketNotEmpty                                 AwsCode = "BucketNotEmpty"
	AwsNoSuchBucketPolicy                             AwsCode = "NoSuchBucketPolicy"
	AwsNoSuchCORSConfiguration                        AwsCode = "NoSuchCORSConfiguration"
	AwsNotImplemented                                 AwsCode = "NotImplemented"
	AwsNoSuchWebsiteConfiguration                     AwsCode = "NoSuchWebsiteConfiguration"
	AwsObjectLockConfigurationNotFoundError           AwsCode = "ObjectLockConfigurationNotFoundError"
	AwsNoSuchLifecycleConfiguration                   AwsCode = "NoSuchLifecycleConfiguration"
	AwsServerSideEncryptionConfigurationNotFoundError AwsCode = "ServerSideEncryptionConfigurationNotFoundError"
	AwsNoSuchEncryptionConfiguration                  AwsCode = "NoSuchEncryptionConfiguration"
)

func retryOnAwsCodes[T any](ctx context.Context, codes []AwsCode, f func() (T, error)) (T, error) {
	var resp T
	err := retry.RetryContext(ctx, retryTimeout, func() *retry.RetryError {
		var err error
		resp, err = f()
		if err != nil {
			var awsErr awserr.Error
			if errors.As(err, &awsErr) {
				for _, code := range codes {
					if awsErr.Code() == string(code) {
						return retry.RetryableError(err)
					}
				}
			}
			return retry.NonRetryableError(err)
		}
		return nil
	})
	return resp, err
}

// retryLongTermOperations retries on some AWS codes because some previous operations are asynchronous and need to wait for the result.
func retryLongTermOperations[T any](ctx context.Context, f func() (T, error)) (T, error) {
	return retryOnAwsCodes[T](ctx, []AwsCode{AwsNoSuchBucket, AwsAccessDenied, AwsForbidden}, f)
}

// Returns true if the error matches all these conditions:
//   - err is of type awserr.Error
//   - Error.Code() matches code
//   - Error.Message() contains message
func isAWSErr(err error, code AwsCode, message string) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == string(code) && strings.Contains(awsErr.Message(), message)
	}
	return false
}
