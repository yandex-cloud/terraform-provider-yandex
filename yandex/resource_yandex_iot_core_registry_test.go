package yandex

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	iot "github.com/yandex-cloud/go-genproto/yandex/cloud/iot/devices/v1"
)

const iotRegistryResource = "yandex_iot_core_registry.test-registry"

func init() {
	resource.AddTestSweepers("yandex_iot_core_registry", &resource.Sweeper{
		Name: "yandex_iot_core_registry",
		F:    testSweepIoTCoreRegistry,
		Dependencies: []string{
			"yandex_iot_core_device",
		},
	})
}

func testSweepIoTCoreRegistry(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &iot.ListRegistriesRequest{FolderId: conf.FolderID}
	it := conf.sdk.IoT().Devices().Registry().RegistryIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepIoTCoreRegistry(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep IoT Core Registry %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepIoTCoreRegistry(conf *Config, id string) bool {
	return sweepWithRetry(sweepIoTCoreRegistryOnce, conf, "IoT Core Registry", id)
}

func sweepIoTCoreRegistryOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexIoTDefaultTimeout)
	defer cancel()

	op, err := conf.sdk.IoT().Devices().Registry().Delete(ctx, &iot.DeleteRegistryRequest{
		RegistryId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}

func TestAccYandexIoTCoreRegistry_basic(t *testing.T) {
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")
	folderID := getExampleFolderID()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testYandexIoTCoreRegistryBasic(registryName),
				Check: resource.ComposeTestCheckFunc(
					testYandexIoTCoreRegistryExists(iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "folder_id"),
					resource.TestCheckResourceAttr(iotRegistryResource, "folder_id", folderID),
					testYandexIoTCoreRegistryContainsLabel(&registry, "tf-label", "tf-label-value"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "empty-label", ""),
					testAccCheckCreatedAtAttr(iotRegistryResource),
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
	t.Parallel()

	var registry iot.Registry
	registryName := acctest.RandomWithPrefix("tf-iot-core-registry")

	authInfo := &yandexIotCoreAuth{}

	certDefault, _ := ioutil.ReadFile("test-fixtures/iot/reg_default.pub")
	certUpdated, _ := ioutil.ReadFile("test-fixtures/iot/reg_updated.pub")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testYandexIoTCoreRegistryDestroy,
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
					testYandexIoTCoreRegistryExists(iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName),
					resource.TestCheckResourceAttr(iotRegistryResource, "description", "description"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.min_level", "ERROR"),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "log_options.0.log_group_id"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "label_key", "label"),
					testYandexIoTCoreStoreCertificates(authInfo, &registry),
					testYandexIoTCoreStorePasswords(authInfo, &registry),
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
					testYandexIoTCoreRegistryExists(iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName+"_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "description", "description_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.min_level", "DEBUG"),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "log_options.0.log_group_id"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "label_key_updated", "label_updated"),
					testYandexIoTCoreNoChangeCertificates(authInfo, &registry),
					testYandexIoTCoreChangePasswords(authInfo, &registry),
					testYandexIoTCoreStorePasswords(authInfo, &registry),
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
					testYandexIoTCoreRegistryExists(iotRegistryResource, &registry),
					resource.TestCheckResourceAttr(iotRegistryResource, "name", registryName+"_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "description", "description_updated"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.disabled", "false"),
					resource.TestCheckResourceAttr(iotRegistryResource, "log_options.0.min_level", "DEBUG"),
					resource.TestCheckResourceAttrSet(iotRegistryResource, "log_options.0.log_group_id"),
					testYandexIoTCoreRegistryContainsLabel(&registry, "label_key_updated", "label_updated"),
					testYandexIoTCoreChangeCertificates(authInfo, &registry),
					testYandexIoTCoreStoreCertificates(authInfo, &registry),
					testYandexIoTCoreNoChangePasswords(authInfo, &registry),
				),
			},
		},
	})
}

func testYandexIoTCoreRegistryExists(name string, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := testGetRegistryByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("IoT Registry not found")
		}

		*registry = *found
		return nil
	}
}

func testYandexIoTCoreRegistryDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_iot_core_registry" {
			continue
		}

		_, err := testGetRegistryByID(config, rs.Primary.ID)
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

func testYandexIoTCoreStoreCertificates(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		authInfo.certificates, err = testGetRegistryCertificatesByID(testAccProvider.Meta().(*Config), registry.Id)
		return err
	}
}

func testYandexIoTCoreStorePasswords(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		authInfo.passwords, err = testGetRegistryPasswordsByID(testAccProvider.Meta().(*Config), registry.Id)
		return err
	}
}

func testYandexIoTCoreNoChangeCertificates(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		certificatesNew, err := testGetRegistryCertificatesByID(testAccProvider.Meta().(*Config), registry.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must not be changed, but it is")
		}
		return nil
	}
}

func testYandexIoTCoreNoChangePasswords(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		passwordsNew, err := testGetRegistryPasswordsByID(testAccProvider.Meta().(*Config), registry.Id)
		if err != nil {
			return err
		}
		if !testIsEqualSets(authInfo.passwords, passwordsNew) {
			return fmt.Errorf("Passwords must not be changed, but it is")
		}
		return nil
	}
}

func testYandexIoTCoreChangeCertificates(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		certificatesNew, err := testGetRegistryCertificatesByID(testAccProvider.Meta().(*Config), registry.Id)
		if err != nil {
			return err
		}
		if testIsEqualSets(authInfo.certificates, certificatesNew) {
			return fmt.Errorf("Certificates must be changed, but it isn't")
		}
		return nil
	}
}

func testYandexIoTCoreChangePasswords(authInfo *yandexIotCoreAuth, registry *iot.Registry) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		passwordsNew, err := testGetRegistryPasswordsByID(testAccProvider.Meta().(*Config), registry.Id)
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
	return templateConfig(`
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

func testGetRegistryByID(config *Config, ID string) (*iot.Registry, error) {
	req := iot.GetRegistryRequest{
		RegistryId: ID,
	}

	return config.sdk.IoT().Devices().Registry().Get(context.Background(), &req)
}

func testGetRegistryCertificatesByID(config *Config, ID string) (map[string]interface{}, error) {
	certs, err := config.sdk.IoT().Devices().Registry().ListCertificates(context.Background(), &iot.ListRegistryCertificatesRequest{RegistryId: ID})
	if err != nil {
		return nil, err
	}

	res := make(map[string]interface{})
	for _, cert := range certs.Certificates {
		res[cert.CertificateData] = nil
	}
	return res, nil
}

func testGetRegistryPasswordsByID(config *Config, ID string) (map[string]interface{}, error) {
	passwords, err := config.sdk.IoT().Devices().Registry().ListPasswords(context.Background(), &iot.ListRegistryPasswordsRequest{RegistryId: ID})
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
