package acceptance

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	frameworkprovider "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/providermux"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
)

type Session struct {
	provider *schema.Provider
}

func New(t *testing.T) *Session {
	t.Helper()
	provider, err := yandex.NewPluginSDKProviderWithOptions(yandex.ProviderOptions{ConfigurationContext: context.Background()})
	if err != nil {
		t.Fatalf("register acceptance provider: %v", err)
	}
	return &Session{provider: provider}
}
func (s *Session) Provider() *schema.Provider {
	return s.provider
}

func (s *Session) Providers() map[string]*schema.Provider {
	return map[string]*schema.Provider{"yandex": s.provider}
}
func (s *Session) Protocol6Factories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){"yandex": func() (tfprotov6.ProviderServer, error) {
		factory, err := providermux.New(context.Background(), s.provider, frameworkprovider.NewFrameworkProvider())
		if err != nil {
			return nil, err
		}
		return factory(), nil
	}}
}
func (s *Session) Config() (*yandex.Config, error) {
	meta := s.provider.Meta()
	if meta == nil {
		return nil, fmt.Errorf("acceptance provider is not configured")
	}
	config, ok := meta.(*yandex.Config)
	if !ok {
		return nil, fmt.Errorf("acceptance provider configuration: expected non-nil *yandex.Config, got %T", meta)
	}
	if config == nil {
		return nil, fmt.Errorf("acceptance provider configuration is a nil *yandex.Config")
	}
	return config, nil
}

var testAccEnvVars = []string{
	"YC_FOLDER_ID",
	"YC_ZONE",
	"YC_LOGIN",
	"YC_LOGIN_2",
	"YC_STORAGE_ENDPOINT_URL",
	"YC_MESSAGE_QUEUE_ENDPOINT",
}

var testAccForAuthEnvVars = []string{
	"YC_TOKEN",
	"YC_SERVICE_ACCOUNT_KEY_FILE",
}

func PreCheck(t *testing.T) {
	for _, varName := range testAccEnvVars {
		if val := os.Getenv(varName); val == "" {
			t.Fatalf("%s must be set for acceptance tests", varName)
		}
	}

	for _, varName := range testAccForAuthEnvVars {
		if val := os.Getenv(varName); val != "" {
			return
		}
	}
	t.Fatalf("one of the variables: %s must be set for acceptance tests", strings.Join(testAccForAuthEnvVars, ", "))
}

func (s *Session) ProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){"yandex": func() (*schema.Provider, error) {
		return s.provider, nil
	}}
}

func FolderID() string {
	return os.Getenv("YC_FOLDER_ID")
}
