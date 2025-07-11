package s3

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

const (
	defaultS3Region = "ru-central1"
	iamTokenHeader  = "X-YaCloud-SubjectToken"
)

type Client struct {
	s3 *s3.S3
}

func newS3Client(ctx context.Context, accessKey, secretKey, iamToken, url string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("storage endpoint url is not specified")
	}

	config := &aws.Config{
		Endpoint: aws.String(url),
		Region:   aws.String(defaultS3Region),
		LogLevel: aws.LogLevel(aws.LogDebug),
		Logger: aws.LoggerFunc(func(args ...any) {
			tflog.Debug(ctx, fmt.Sprint(args...))
		}),
	}
	switch {
	case accessKey != "" && secretKey != "":
		config.Credentials = credentials.NewStaticCredentials(accessKey, secretKey, "")
	case iamToken != "":
		config.Credentials = credentials.AnonymousCredentials
		config.HTTPClient = &http.Client{
			Transport: newTransport(iamToken),
		}
	default:
		return nil, fmt.Errorf("nor token, nor access and secret keys are specified")
	}

	ssn, err := session.NewSession(config)
	if err != nil {
		return nil, fmt.Errorf("failed to init session: %w", err)
	}

	return &Client{
		s3: s3.New(ssn, config),
	}, nil
}

func GetS3Client(ctx context.Context, accessKey, secretKey string, c *provider_config.Config) (*Client, error) {
	if accessKey == "" && !c.ProviderState.StorageAccessKey.IsUnknown() && !c.ProviderState.StorageAccessKey.IsNull() {
		accessKey = c.ProviderState.StorageAccessKey.ValueString()
	}
	if secretKey == "" && !c.ProviderState.StorageSecretKey.IsUnknown() && !c.ProviderState.StorageSecretKey.IsNull() {
		secretKey = c.ProviderState.StorageSecretKey.ValueString()
	}

	token := ""
	if !c.ProviderState.Token.IsUnknown() && !c.ProviderState.Token.IsNull() {
		token = c.ProviderState.Token.ValueString()
	}

	storageEndpoint := ""
	if !c.ProviderState.StorageEndpoint.IsUnknown() && !c.ProviderState.StorageEndpoint.IsNull() {
		storageEndpoint = c.ProviderState.StorageEndpoint.ValueString()
	}

	return newS3Client(ctx, accessKey, secretKey, token, storageEndpoint)
}

type iamTransport struct {
	Transport http.RoundTripper
	IAMToken  string
}

func newTransport(iamToken string) http.RoundTripper {
	return &iamTransport{
		Transport: http.DefaultTransport,
		IAMToken:  iamToken,
	}
}

func (t *iamTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(iamTokenHeader, t.IAMToken)
	return t.Transport.RoundTrip(req)
}
