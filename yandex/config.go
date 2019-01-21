package yandex

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/hashicorp/terraform/terraform"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/pkg/requestid"
)

type Config struct {
	Endpoint  string
	FolderID  string
	CloudID   string
	Zone      string
	Token     string
	Plaintext bool
	Insecure  bool

	userAgent string
	sdk       *ycsdk.SDK
}

// Client configures and returns a fully initialized Yandex.Cloud sdk
func (c *Config) initAndValidate() error {
	yandexSDKConfig := &ycsdk.Config{
		Credentials: ycsdk.OAuthToken(c.Token),
		Endpoint:    c.Endpoint,
		Plaintext:   c.Plaintext,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: c.Insecure,
		},
	}

	terraformURL := "https://www.terraform.io"
	c.userAgent = fmt.Sprintf("Terraform/%s (%s)", terraform.VersionString(), terraformURL)

	headerMD := metadata.Pairs("user-agent", c.userAgent)

	var err error
	c.sdk, err = ycsdk.Build(context.Background(), *yandexSDKConfig,
		grpc.WithUserAgent(c.userAgent),
		grpc.WithDefaultCallOptions(grpc.Header(&headerMD)),
		grpc.WithUnaryInterceptor(requestid.Interceptor()))

	return err
}
