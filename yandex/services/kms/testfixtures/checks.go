package testfixtures

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	kmssdk "github.com/yandex-cloud/go-sdk/services/kms/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

type SymmetricKeyGetter interface {
	Get(context.Context, *kms.GetSymmetricKeyRequest, ...grpc.CallOption) (*kms.SymmetricKey, error)
}

func CheckSymmetricKeysDestroy(clientForTest func() (SymmetricKeyGetter, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var client SymmetricKeyGetter
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "yandex_kms_symmetric_key" {
				continue
			}
			if client == nil {
				var err error
				client, err = clientForTest()
				if err != nil {
					return fmt.Errorf("get KMS client: %w", err)
				}
			}
			_, err := client.Get(context.Background(), &kms.GetSymmetricKeyRequest{KeyId: rs.Primary.ID})
			switch status.Code(err) {
			case codes.OK:
				return fmt.Errorf("KMS key %s still exists", rs.Primary.ID)
			case codes.NotFound:
				continue
			default:
				return fmt.Errorf("check deletion of KMS key %s: %w", rs.Primary.ID, err)
			}
		}
		return nil
	}
}

func CheckSymmetricKeyExists(config func() (*providerconfig.Config, error), name string, symmetricKey *kms.SymmetricKey) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		cfg, configErr := config()
		if configErr != nil {
			return configErr
		}

		client := kmssdk.NewSymmetricKeyClient(cfg.SDK)

		found, err := client.Get(context.Background(), &kms.GetSymmetricKeyRequest{
			KeyId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.Id != rs.Primary.ID {
			return fmt.Errorf("KMS Symmetric Key not found")
		}

		*symmetricKey = *found

		return nil
	}
}
