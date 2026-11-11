package acceptance

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	iotfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iot_core/testfixtures"

	statecheck "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance/statecheck"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
	devicessdk "github.com/yandex-cloud/go-sdk/services/iot/devices/v1"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

const iotRegistryResource = "yandex_iot_core_registry.test-registry"

func TestAccYandexIoTCoreRegistry_basic(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	folderID := acceptance.FolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexIoTCoreRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreRegistryBasic(registryName),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckRegistryExists(suite.Config, iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "folder_id"),
					resource.TestCheckResourceAttr(iotRegistryResource, "folder_id", folderID),
					testYandexIoTCoreRegistryContainsLabel(&registry, "tf-label", "tf-label-value"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "empty-label", ""),
					statecheck.CheckCreatedAtAttr(iotRegistryResource),
				),
			},
		},
	})
}

type yandexIotCoreAuth struct {
	certificates map[string]interface{}
	passwords    map[string]interface{}
}

func TestAccYandexIoTCoreRegistry_update(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")

	authInfo := &yandexIotCoreAuth{}

	certDefault, _ := ioutil.ReadFile("testdata/reg_default.pub")
	certUpdated, _ := ioutil.ReadFile("testdata/reg_updated.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		Providers:    suite.Providers(),
		CheckDestroy: suite.testYandexIoTCoreRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreRegistryFull(
					registryName,
					"description",
					"label_key",
					"label",
					"ERROR",
					"0123456789_abcd",
					string(certDefault)),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckRegistryExists(suite.Config, iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName),
					resource.TestCheckResourceAttr(iotRegistryResource, "description", "description"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.min_level", "ERROR"),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "log_options.0.log_group_id"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "label_key", "label"),
					suite.testYandexIoTCoreStoreCertificates(authInfo, &registry),
					suite.testYandexIoTCoreStorePasswords(authInfo, &registry),
				),
			},
			{
				Config: testYandexIoTCoreRegistryFull(
					registryName+"_updated",
					"description_updated",
					"label_key_updated",
					"label_updated",
					"DEBUG",
					"0123456789_updated",
					string(certDefault)),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckRegistryExists(suite.Config, iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName+"_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "description", "description_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.min_level", "DEBUG"),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "log_options.0.log_group_id"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "label_key_updated", "label_updated"),
					suite.testYandexIoTCoreNoChangeCertificates(authInfo, &registry),
					suite.testYandexIoTCoreChangePasswords(authInfo, &registry),
					suite.testYandexIoTCoreStorePasswords(authInfo, &registry),
				),
			},
			{
				Config: testYandexIoTCoreRegistryFull(
					registryName+"_updated",
					"description_updated",
					"label_key_updated",
					"label_updated",
					"DEBUG",
					"0123456789_updated",
					string(certUpdated)),
				Check: resource.ComposeTestCheckFunc(
					iotfixtures.CheckRegistryExists(suite.Config, iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName+"_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "description", "description_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.min_level", "DEBUG"),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "log_options.0.log_group_id"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "label_key_updated", "label_updated"),
					suite.testYandexIoTCoreChangeCertificates(authInfo, &registry),
					suite.testYandexIoTCoreStoreCertificates(authInfo, &registry),
					suite.testYandexIoTCoreNoChangePasswords(authInfo, &registry),
				),
			},
		},
	})
}

func (suite testSuite) testYandexIoTCoreRegistryDestroy(s *terraform.State) error {
	config, configErr := suite.Config()
	if configErr != nil {
		return configErr
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iot_core_registry" {
			continue
		}

		_, err := iotfixtures.GetRegistryByID(config, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Registry still exists")
		}
	}

	return nil
}

func testYandexIoTCoreRegistryContainsLabel(registry *iot.Registry, key string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		v, ok := registry.Labels[key]
		if !ok {
			return fmt.Errorf("Expected label with key '%s' not found", key)
		}
		if v != value {
			return fmt.Errorf("Incorrect label value for key '%s': expected '%s' but found '%s'", key, value, v)
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreStoreCertificates(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		authInfo.certificates, err = testGetRegistryCertificatesByID(providerConfig, registry.Id)
		return err
	}
}

func (suite testSuite) testYandexIoTCoreStorePasswords(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		authInfo.passwords, err = testGetRegistryPasswordsByID(providerConfig, registry.Id)
		return err
	}
}

func (suite testSuite) testYandexIoTCoreNoChangeCertificates(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		certificatesNew, err := testGetRegistryCertificatesByID(providerConfig, registry.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must not be changed, but it is")
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreNoChangePasswords(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		passwordsNew, err := testGetRegistryPasswordsByID(providerConfig, registry.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.passwords, passwordsNew) {
			return fmt.Errorf("Passwords must not be changed, but it is")
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreChangeCertificates(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		certificatesNew, err := testGetRegistryCertificatesByID(providerConfig, registry.Id)
		if err != nil {
			return err
		}
		if testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must be changed, but it isn't")
		}
		return nil
	}
}

func (suite testSuite) testYandexIoTCoreChangePasswords(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConfig, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}
		passwordsNew, err := testGetRegistryPasswordsByID(providerConfig, registry.Id)
		if err != nil {
			return err
		}
		if testIsEqualSets(authInfo.passwords, passwordsNew) {
			return fmt.Errorf("Passwords must be changed, but it isn't")
		}
		return nil
	}
}

func testYandexIoTCoreRegistryBasic(name string) string {
	return fmt.Sprintf(`
resource "yandex_iot_core_registry" "test-registry" {
  name = "%s"
  labels = {
    tf-label    = "tf-label-value"
    empty-label = ""
  }
}
	`, name)
}

func testYandexIoTCoreRegistryFull(name string, descr string, labelKey string, label string, minLogLevel string, password string, cert string) string {
	return sdkcommon.TemplateConfig(`
resource "yandex_iot_core_registry" "test-registry" {
  name        = "{{.Name}}"
  description = "{{.Descr}}"
  labels = {
    {{.LabelKey}} = "{{.Label}}"
    empty-label   = ""
  }
  log_options {
	log_group_id = yandex_logging_group.logging-group.id
	min_level = "{{.MinLogLevel}}"
  }
  passwords = [
    "{{.Password}}"
  ]
  certificates = [<<EOF
{{.Cert}}
EOF
  ]
}

resource "yandex_logging_group" "logging-group" {
}

	`, map[string]interface{}{
		"Name":        name,
		"Descr":       descr,
		"LabelKey":    labelKey,
		"Label":       label,
		"MinLogLevel": minLogLevel,
		"Password":    password,
		"Cert":        cert,
	})
}

func testGetRegistryCertificatesByID(config *providerconfig.Config, ID string) (map[string]interface{}, error) {
	client := devicessdk.NewRegistryClient(config.SDK)

	certs, err := client.ListCertificates(context.Background(), &iot.ListRegistryCertificatesRequest{RegistryId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, cert := range certs.Certificates {
		res[cert.CertificateData] = nil
	}
	return res, nil
}

func testGetRegistryPasswordsByID(config *providerconfig.Config, ID string) (map[string]interface{}, error) {
	client := devicessdk.NewRegistryClient(config.SDK)

	passwords, err := client.ListPasswords(context.Background(), &iot.ListRegistryPasswordsRequest{RegistryId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, pass := range passwords.Passwords {
		res[pass.Id] = nil
	}
	return res, nil
}

func testIsEqualSets(s1 map[string]interface{}, s2 map[string]interface{}) bool {
	if len(s1) != len(s2) {
		return false
	}

	for e := range s1 {
		_, ok := s2[e]
		if !ok {
			return false
		}
	}
	return true
}
