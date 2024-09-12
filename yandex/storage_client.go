package yandex

import (
	"context"
	"errors"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/storage/s3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getS3ClientByKeys(ctx context.Context, accessKey, secretKey string, c *Config) (*s3.Client, error) {
	iamToken, err := c.getIAMToken(ctx)
	if err != nil {
		return nil, err
	}

	if (accessKey == "" || secretKey == "") && iamToken == "" {
		if c.defaultS3Client == nil {
			return nil, fmt.Errorf("failed to get default storage client")
		}

		return c.defaultS3Client, nil
	}

	return s3.NewClient(ctx, accessKey, secretKey, iamToken, c.StorageEndpoint)
}

func getS3Client(ctx context.Context, d *schema.ResourceData, c *Config) (*s3.Client, error) {
	ak, sk, err := getS3Keys(d)

	if err != nil {
		return nil, err
	}

	return getS3ClientByKeys(ctx, ak, sk, c)
}

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

	if hasAccessKey != hasSecretKey || (hasAccessKey && (accessKey == "" || secretKey == "")) {
		return "", "", errors.New("both access and secret keys should be specified")
	}

	return accessKey, secretKey, nil
}
