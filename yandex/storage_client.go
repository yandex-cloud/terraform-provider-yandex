package yandex

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
