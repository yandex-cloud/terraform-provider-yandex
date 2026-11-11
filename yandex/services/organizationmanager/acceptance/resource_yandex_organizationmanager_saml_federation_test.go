package acceptance

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/acceptanceenv"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/saml"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func TestAccOrganizationManagerSamlFederation_createAndUpdate(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	// We do not expect to test every possible Create and Update scenarios in one run but we will eventually do this
	// as more and more runs are made over time.
	suite.testAccSamlFederationRunTest(t, testAccOrganizationManagerSamlFederation, true, 3)
}

func TestAccOrganizationManagerSamlFederation_import(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	info := newSamlFederationInfo()
	name := info.getResourceName(true)

	var fed saml.Federation
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckSamlFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganizationManagerSamlFederation(info),
				Check: resource.ComposeTestCheckFunc(
					suite.testAccCheckSamlFederationExists(name, &fed),
				),
			},
			organizationSamlFederationImportStep(name, fed.Id),
		},
	})
}

// The config here should match as closely as possible to the one presented to the user in the docs.
// Serves as a proof that the example config is viable.
func TestAccOrganizationManagerSamlFederation_example(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	config := fmt.Sprintf(`
resource "yandex_organizationmanager_saml_federation" federation {
  name            = "my-federation"
  description     = "My new SAML federation"
  organization_id = "%s"
  sso_url         = "https://my-sso.url"
  issuer          = "my-issuer"
  sso_binding     = "POST"
}
`, acceptanceenv.OrganizationID())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testAccCheckSamlFederationDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
			},
		},
	})
}

type SamlFederationConfigGenerateFunc func(info *resourceSamlFederationInfo) string

func (suite testSuite) testAccSamlFederationRunTest(t *testing.T, fun SamlFederationConfigGenerateFunc, rs bool, n int) {
	// Generate n federations, apply them to Terraform using fun and test according to resource type.
	for i := 0; i < n; i++ {
		info := newSamlFederationInfo()
		var fed saml.Federation
		resource.Test(t, resource.TestCase{
			PreCheck:     func() { acceptance.PreCheck(t) },
			Providers:    suite.Providers(),
			CheckDestroy: suite.testAccCheckSamlFederationDestroy,
			Steps: []resource.TestStep{
				{
					Config: fun(info),
					Check: resource.ComposeTestCheckFunc(
						suite.testAccCheckSamlFederationExists(info.getResourceName(rs), &fed),
						samlFederationResourceTestCheckFunc(&fed, info, rs),
					),
				},
			},
		})
	}
}

func (suite testSuite) testAccCheckSamlFederationExists(n string, federation *saml.Federation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := samlsdk.NewFederationClient(config.SDK)

		found, err := client.Get(context.Background(), &saml.GetFederationRequest{
			FederationId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("SAML Federation not found")
		}

		*federation = *found
		return nil
	}
}

func samlFederationResourceTestCheckFunc(fed *saml.Federation, info *resourceSamlFederationInfo, rs bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := info.getResourceName(rs)
		checkFuncsAr := []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(name, "name", info.Name),
			resource.TestCheckResourceAttr(name, "description", info.Description),
			resource.TestCheckResourceAttr(name, "sso_binding", info.SsoBinding),
			resource.TestCheckResourceAttr(name, "sso_url", info.SsoUrl),
			resource.TestCheckResourceAttr(name, "issuer", info.Issuer),
			statecheck.CheckAttributeWithSuppress(sdkcommon.ShouldSuppressDiffForTimeDuration, name, "cookie_max_age", info.CookieMaxAge),
			resource.TestCheckResourceAttr(name, "auto_create_account_on_login", strconv.FormatBool(info.AutoCreateAccountOnLogin)),
			resource.TestCheckResourceAttr(name, "case_insensitive_name_ids", strconv.FormatBool(info.CaseInsensitiveNameIds)),
			// Uncomment once labels are supported.
			// resource.TestCheckResourceAttr(name, fmt.Sprintf("labels.%s", info.LabelKey), info.LabelValue),
		}

		expectedEncryptedAssertions := "false"

		if info.SecuritySettings != nil && info.SecuritySettings.EncryptedAssertions {
			expectedEncryptedAssertions = strconv.FormatBool(info.SecuritySettings.EncryptedAssertions)
		}

		checkFuncsAr = append(checkFuncsAr, resource.TestCheckResourceAttr(name, "security_settings.0.encrypted_assertions", expectedEncryptedAssertions))

		if fed.SecuritySettings == nil {
			return fmt.Errorf("unexpected nil in federation's SecuritySettings")
		}
		value := info.SecuritySettings != nil && info.SecuritySettings.EncryptedAssertions
		if fed.SecuritySettings.EncryptedAssertions != value {
			return fmt.Errorf("expected %v for federation.SecuritySettings.EncryptedAssertions, got %v", value, fed.SecuritySettings.EncryptedAssertions)
		}
		return resource.ComposeTestCheckFunc(checkFuncsAr...)(s)
	}
}

func (suite testSuite) testAccCheckSamlFederationDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_saml_federation" {
			continue
		}

		client := samlsdk.NewFederationClient(config.SDK)

		_, err := client.Get(context.Background(), &saml.GetFederationRequest{
			FederationId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("SAML Federation still exists")
		}
	}

	return nil
}

type resourceSamlFederationInfo struct {
	OrganizationId           string
	Name                     string
	Description              string
	LabelKey                 string
	LabelValue               string
	Issuer                   string
	SsoBinding               string
	SsoUrl                   string
	CookieMaxAge             string
	AutoCreateAccountOnLogin bool
	CaseInsensitiveNameIds   bool
	SecuritySettings         *saml.FederationSecuritySettings
	ResourceName             string
}

func generateFederationSecuritySettings() *saml.FederationSecuritySettings {
	r := rand.Intn(7)

	switch r {
	case 0:
		return nil
	case 1:
		return &saml.FederationSecuritySettings{
			EncryptedAssertions: true,
			ForceAuthn:          true,
		}
	case 2:
		return &saml.FederationSecuritySettings{
			EncryptedAssertions: true,
			ForceAuthn:          false,
		}
	case 3:
		return &saml.FederationSecuritySettings{
			EncryptedAssertions: false,
			ForceAuthn:          true,
		}
	case 4:
		return &saml.FederationSecuritySettings{
			EncryptedAssertions: false,
			ForceAuthn:          false,
		}
	case 5:
		return &saml.FederationSecuritySettings{
			ForceAuthn: false,
		}
	case 6:
		return &saml.FederationSecuritySettings{
			EncryptedAssertions: true,
		}
	}

	panic("generated invalid saml.FederationSecuritySettings")
}

func generateFederationBinding() saml.BindingType {
	r := rand.Intn(3)

	switch r {
	case 0:
		return saml.BindingType_POST
	case 1:
		return saml.BindingType_REDIRECT
	case 2:
		return saml.BindingType_ARTIFACT
	}

	panic("generated invalid saml.BindingType")
}

func newSamlFederationInfo() *resourceSamlFederationInfo {
	duration := fmt.Sprintf("%dm", 10+rand.Intn(10))

	return &resourceSamlFederationInfo{
		OrganizationId:           acceptanceenv.OrganizationID(),
		Name:                     acctest.RandomWithPrefix("tf-acc"),
		Description:              acctest.RandString(20),
		Issuer:                   acctest.RandomWithPrefix("issuer"),
		SsoBinding:               generateFederationBinding().String(),
		SsoUrl:                   acctest.RandomWithPrefix("https://sso-url"),
		CookieMaxAge:             duration,
		ResourceName:             "foobar",
		AutoCreateAccountOnLogin: rand.Intn(1) == 1, // TODO: SA4030: math/rand.Intn(n) generates a random value 0 <= x < n; that is, the generated values don't include n; rand.Intn(1) therefore always returns 0 (staticcheck)
		CaseInsensitiveNameIds:   rand.Intn(1) == 1, // TODO: SA4030: math/rand.Intn(n) generates a random value 0 <= x < n; that is, the generated values don't include n; rand.Intn(1) therefore always returns 0 (staticcheck)
		SecuritySettings:         generateFederationSecuritySettings(),
		// Uncomment once labels are supported.
		// LabelKey:       "label_key",
		// LabelValue:     "label_value",
	}
}

func (i *resourceSamlFederationInfo) Map() map[string]interface{} {
	return structs.Map(i)
}

func (i *resourceSamlFederationInfo) getResourceName(rs bool) string {
	if rs {
		return "yandex_organizationmanager_saml_federation." + i.ResourceName
	}
	return "data.yandex_organizationmanager_saml_federation." + i.ResourceName
}

const samlFederationConfigTemplate = `
resource "yandex_organizationmanager_saml_federation" {{.ResourceName}} {
  name                         = "{{.Name}}"
  description                  = "{{.Description}}"
  organization_id              = "{{.OrganizationId}}"
  issuer                       = "{{.Issuer}}"
  sso_binding                  = "{{.SsoBinding}}"
  sso_url                      = "{{.SsoUrl}}"
  cookie_max_age               = "{{.CookieMaxAge}}"
  auto_create_account_on_login = {{.AutoCreateAccountOnLogin}}
  case_insensitive_name_ids    = {{.CaseInsensitiveNameIds}}
  {{if .SecuritySettings}}
  security_settings {
    encrypted_assertions = {{.SecuritySettings.EncryptedAssertions}} 
	force_authn          = {{.SecuritySettings.ForceAuthn}}
  }
  {{end}}

  {{if .LabelKey}}
  labels = {
    {{.LabelKey}} = "{{.LabelValue}}"
  }
  {{end}}
}
`

func testAccOrganizationManagerSamlFederation(info *resourceSamlFederationInfo) string {
	return sdkcommon.TemplateConfig(samlFederationConfigTemplate, info.Map())
}

func organizationSamlFederationImportStep(resourceName, federationID string) resource.TestStep {
	return resource.TestStep{
		ResourceName:      resourceName,
		ImportStateId:     federationID,
		ImportState:       true,
		ImportStateVerify: true,
	}
}
