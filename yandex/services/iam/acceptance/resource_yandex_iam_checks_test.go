package acceptance

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/vault/helper/pgpkeys"
	svc_iam "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/pluginsdk"
)

func (suite testSuite) testAccCheckServiceAccountIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		bindings, err := svc_iam.GetServiceAccountAccessBindings(config.Context(), config, rs.Primary.ID)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("Binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func testDecryptKeyAndTest(name, key, pgpKey string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		ciphertext, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("can't find '%s' attr for %s resource", key, name)
		}

		// We can't verify that the decrypted ciphertext is correct, because we don't
		// have it. We can verify that decrypting it does not error
		_, err := pgpkeys.DecryptBytes(ciphertext, pgpKey)
		if err != nil {
			return fmt.Errorf("error decrypting ciphertext: %s", err)
		}

		return nil
	}
}
