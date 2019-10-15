package yandex

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const defaultS3Region = "ru-central1"

func getS3ClientByKeys(accessKey, secretKey string, c *Config) (*s3.S3, error) {
	if accessKey == "" || secretKey == "" {
		if c.defaultS3Client == nil {
			return nil, fmt.Errorf("failed to get default storage client")
		}

		return c.defaultS3Client, nil
	}

	return newS3Client(c.StorageEndpoint, accessKey, secretKey)
}

func getS3Client(d *schema.ResourceData, c *Config) (*s3.S3, error) {
	ak, sk, err := getS3Keys(d)

	if err != nil {
		return nil, err
	}

	return getS3ClientByKeys(ak, sk, c)
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

	if hasAccessKey != hasSecretKey {
		err = fmt.Errorf("both access and secret keys should be specified")
		return
	}

	if hasAccessKey && (accessKey == "" || secretKey == "") {
		err = fmt.Errorf("access and secret keys should not be empty")
		return
	}

	return
}

func newS3Client(url, accessKey, secretKey string) (*s3.S3, error) {
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
		return nil, err
	}

	return s3.New(newSession), nil
}
