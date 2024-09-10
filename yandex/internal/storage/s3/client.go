package s3

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	defaultS3Region = "ru-central1"
)

type Client struct {
	s3 *s3.S3
}

func NewClient(ctx context.Context, accessKey, secretKey, url string) (*Client, error) {
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("accessKey or sercretKey is not specified")
	}
	if url == "" {
		return nil, fmt.Errorf("storage endpoint url is not specified")
	}

	config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:    aws.String(url),
		Region:      aws.String(defaultS3Region),
		LogLevel:    aws.LogLevel(aws.LogDebug),
		Logger: aws.LoggerFunc(func(args ...any) {
			tflog.Debug(ctx, fmt.Sprint(args...))
		}),
	}
	ssn, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to init session: %w", err)
	}

	return &Client{
		s3: s3.New(ssn, config),
	}, nil
}

// S3 use only for test for backward compatibility with old code
// do not use it in new code
func (c *Client) S3() *s3.S3 {
	return c.s3
}
