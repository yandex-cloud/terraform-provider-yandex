package testfixtures

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	functionssdk "github.com/yandex-cloud/go-sdk/services/serverless/functions/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func CheckExists(configForTest func() (*providerconfig.Config, error), name string, function *functions.Function) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, configErr := configForTest()
		if configErr != nil {
			return configErr
		}

		found, err := GetByID(config, rs.Primary.ID)
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("Function not found")
		}

		*function = *found
		return nil
	}
}

func GetByID(config *providerconfig.Config, ID string) (*functions.Function, error) {
	req := functions.GetFunctionRequest{
		FunctionId: ID,
	}

	client := functionssdk.NewFunctionClient(config.SDK)

	return client.Get(context.Background(), &req)
}
