package acceptance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
)

type testSuite struct{ *acceptance.Session }

func (suite testSuite) defaultAndEmptyFolderProviders(t *testing.T) []map[string]*schema.Provider {
	t.Helper()
	p, err := yandex.NewPluginSDKProviderWithOptions(yandex.ProviderOptions{IgnoreFolderID: true})
	if err != nil {
		t.Fatalf("register acceptance provider without a default folder: %v", err)
	}
	return []map[string]*schema.Provider{suite.Providers(), {"yandex": p}}
}
