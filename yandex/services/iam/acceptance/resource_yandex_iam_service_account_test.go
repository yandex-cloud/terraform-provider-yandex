package acceptance

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
)

func (suite testSuite) testAccCheckYandexIAMServiceAccountExistsWithID(n string, sa *iam.ServiceAccount) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := suite.Config()
		if configErr != nil {
			return configErr
		}

		client := iamsdk.NewServiceAccountClient(config.SDK)

		found, err := client.Get(context.Background(), &iam.GetServiceAccountRequest{
			ServiceAccountId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Service account not found")
		}

		*sa = *found

		return nil
	}
}
