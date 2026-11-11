package yandex

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	terraform2 "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/stretchr/testify/assert"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/providermux"
)

const providerDefaultValueInsecure = false
const providerDefaultValuePlaintext = false
const providerDefaultValueEndpoint = "api.cloud.yandex.net:443"

var testAccProviders map[string]*schema.Provider
var testAccProviderFactories map[string]func() (*schema.Provider, error)

var testAccProvider *schema.Provider
var testAccProviderServer tfprotov6.ProviderServer
var accProvider provider.Provider
var testAccProviderFactoriesV6 map[string]func() (tfprotov6.ProviderServer, error)

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

func NewFrameworkProviderServer(ctx context.Context) (func() tfprotov6.ProviderServer, error) {
	return providermux.New(ctx, testAccProvider, accProvider)
}

func init() {
	var err error
	testAccProvider, err = NewPluginSDKProvider()
	if err != nil {
		fmt.Fprintln(os.Stderr, "initialize acceptance provider:", err)
		os.Exit(1)
	}
	testAccProvider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		return providerConfigure(context.Background(), d, testAccProvider, false, true)
	}

	testAccProviders = map[string]*schema.Provider{
		"yandex": testAccProvider,
	}

	testAccProviderFactories = map[string]func() (*schema.Provider, error){
		"yandex": func() (*schema.Provider, error) {
			return testAccProvider, nil
		},
	}

	accProvider = yandex_framework.NewFrameworkProvider()
	testAccProviderFunc, err := NewFrameworkProviderServer(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "initialize acceptance provider mux:", err)
		os.Exit(1)
	}
	testAccProviderServer = testAccProviderFunc()
	testAccProviderFactoriesV6 = map[string]func() (tfprotov6.ProviderServer, error){
		"yandex": func() (tfprotov6.ProviderServer, error) {
			return testAccProviderServer, nil
		},
	}

	if os.Getenv("TF_ACC") != "" {
		acceptanceenv.Load()
	}
}

func TestProvider(t *testing.T) {
	provider, err := NewPluginSDKProvider()
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if err := provider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderWithRawConfig(t *testing.T) {
	testProvider, err := NewPluginSDKProvider()
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	raw := map[string]interface{}{
		"insecure": true,
		"token":    "any_string_like_a_oauth",
		"endpoint": "localhost:4433",
		"zone":     "ru-central1-a",
	}

	diags := testProvider.Configure(context.Background(), terraform2.NewResourceConfigRaw(raw))
	if diags != nil && diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("error configuring  provider: %s", d.Summary)
			}
		}
	}

	if err := testProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderDefaultValues(t *testing.T) {
	// save OS env vars
	envVars := []string{"YC_INSECURE", "YC_PLAINTEXT", "YC_ENDPOINT"}
	saveEnvVariable := acceptanceenv.SaveAndUnset(envVars)

	testProvider, err := NewPluginSDKProvider()
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	raw := map[string]interface{}{
		"token": "any_string_like_a_oauth",
		"zone":  "ru-central1-a",
	}

	diags := testProvider.Configure(context.Background(), terraform2.NewResourceConfigRaw(raw))
	if diags != nil && diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("error configuring provider: %s", d.Summary)
			}
		}
	}

	if err := testProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}

	conf := testProvider.Meta().(*Config)
	if conf.Endpoint != providerDefaultValueEndpoint {
		t.Errorf("there is not default API endpoint (%s), should be %s", conf.Endpoint, providerDefaultValueEndpoint)
	}

	if conf.Plaintext {
		t.Errorf(
			"there is not default option 'Plaintext' (%v), should be %v",
			conf.Plaintext,
			providerDefaultValuePlaintext,
		)
	}

	if conf.Insecure {
		t.Errorf(
			"there is not default option 'Insecure' (%v), should be %v",
			conf.Plaintext,
			providerDefaultValueInsecure,
		)
	}

	// restore OS env vars
	if err := acceptanceenv.Restore(saveEnvVariable); err != nil {
		t.Fatal("failed to restore OS env vars:", envVars, "after test", t.Name(), " - error:", err)
	}
}

func TestProviderOrganizationId(t *testing.T) {
	// save OS env vars
	envVars := []string{"YC_ORGANIZATION_ID"}
	saveEnvVariable := acceptanceenv.SaveAndUnset(envVars)
	defer func() {
		// restore OS env vars
		if err := acceptanceenv.Restore(saveEnvVariable); err != nil {
			t.Fatal("failed to restore OS env vars:", envVars, "after test", t.Name(), " - error:", err)
		}
	}()

	testProvider, err := NewPluginSDKProvider()
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	org := acctest.RandomWithPrefix("org")
	raw := map[string]interface{}{
		"token":           "any_string_like_a_oauth",
		"organization_id": org,
	}

	diags := testProvider.Configure(context.Background(), terraform2.NewResourceConfigRaw(raw))
	if diags != nil && diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("error configuring provider: %s", d.Summary)
			}
		}
	}

	if err := testProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}

	conf := testProvider.Meta().(*Config)
	assert.Equal(t, org, conf.OrganizationID)
}

func TestProviderSharedCredentialsFileAndProfile(t *testing.T) {
	testProvider, err := NewPluginSDKProvider()
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	raw := map[string]interface{}{
		"token":                   "any_string_like_a_oauth",
		"shared_credentials_file": "test-fixtures/shared-credentials-file",
		"profile":                 "prod-profile",
	}

	diags := testProvider.Configure(context.Background(), terraform2.NewResourceConfigRaw(raw))
	if diags != nil && diags.HasError() {
		for _, d := range diags {
			if d.Severity == diag.Error {
				t.Fatalf("error configuring provider: %s", d.Summary)
			}
		}
	}

	if err := testProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}

	conf := testProvider.Meta().(*Config)
	assert.Equal(t, "test-fixtures/shared-credentials-file", conf.SharedCredentialsFile)
	assert.Equal(t, "prod-profile", conf.Profile)
}

func testAccPreCheck(t *testing.T) {
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
