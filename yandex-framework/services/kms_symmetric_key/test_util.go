package kms_symmetric_key

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/kms/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func TestAccCheckYandexKmsSymmetricKeyAllDestroyed(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kms_symmetric_key" {
			continue
		}
		if err := testAccCheckYandexKmsSymmetricKeyDestroyed(rs.Primary.ID); err != nil {
			return err
		}
	}
	return nil
}

func testAccCheckYandexKmsSymmetricKeyDestroyed(id string) error {
	cfg := test.AccProvider.(*provider.Provider).GetConfig()
	_, err := cfg.SDK.KMS().SymmetricKey().Get(context.Background(), &kms.GetSymmetricKeyRequest{
		KeyId: id,
	})
	if err == nil {
		return fmt.Errorf("KMS %s still exists", id)
	}
	return nil
}
