package s3

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const retryTimeout = 5 * time.Second

type ErrCode string

const (
	NoSuchBucket                                   ErrCode = s3.ErrCodeNoSuchBucket
	AccessDenied                                   ErrCode = "AccessDenied"
	BadRequest                                     ErrCode = "BadRequest"
	Forbidden                                      ErrCode = "Forbidden"
	MalformedPolicy                                ErrCode = "MalformedPolicy"
	BucketNotEmpty                                 ErrCode = "BucketNotEmpty"
	NoSuchBucketPolicy                             ErrCode = "NoSuchBucketPolicy"
	NoSuchCORSConfiguration                        ErrCode = "NoSuchCORSConfiguration"
	NotImplemented                                 ErrCode = "NotImplemented"
	NoSuchWebsiteConfiguration                     ErrCode = "NoSuchWebsiteConfiguration"
	ObjectLockConfigurationNotFoundError           ErrCode = "ObjectLockConfigurationNotFoundError"
	NoSuchLifecycleConfiguration                   ErrCode = "NoSuchLifecycleConfiguration"
	ServerSideEncryptionConfigurationNotFoundError ErrCode = "ServerSideEncryptionConfigurationNotFoundError"
	NoSuchEncryptionConfiguration                  ErrCode = "NoSuchEncryptionConfiguration"
)

func RetryOnCodes[T any](ctx context.Context, codes []ErrCode, f func() (T, error)) (T, error) {
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

// RetryLongTermOperations retries on some AWS codes because some previous operations are asynchronous and need to wait for the result.
func RetryLongTermOperations[T any](ctx context.Context, f func() (T, error)) (T, error) {
	return RetryOnCodes[T](ctx, []ErrCode{NoSuchBucket, AccessDenied, Forbidden}, f)
}

// IsErr returns true if the error matches all these conditions:
//   - err is of type awserr.Error
//   - Error.Code() matches code
func IsErr(err error, code ErrCode) bool {
	var awsErr awserr.Error
	if errors.As(err, &awsErr) {
		return awsErr.Code() == string(code)
	}
	return false
}
