package pluginsdk

import (
	"context"
	"errors"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/storage/s3"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func GetS3ClientByKeys(ctx context.Context, accessKey, secretKey string, c *providerconfig.Config) (*s3.Client, error) {
	if accessKey == "" || secretKey == "" {
		if c.DefaultS3Client() == nil {
			return nil, fmt.Errorf("failed to get default storage client")
		}
		return c.DefaultS3Client(), nil
	}
	// iamToken is not needed here, since we cannot specify it in the resource.
	// Otherwise, defaultS3Client must be initialised.
	return s3.NewClient(ctx, accessKey, secretKey, "", c.StorageEndpoint)
}

func getS3Client(ctx context.Context, d *schema.ResourceData, c *providerconfig.Config) (*s3.Client, error) {
	accessKey, secretKey, err := getS3Keys(d)
	if err != nil {
		return nil, err
	}
	return GetS3ClientByKeys(ctx, accessKey, secretKey, c)
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
