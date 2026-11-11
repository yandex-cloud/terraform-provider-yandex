package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func Register(add func(string, *resource.Sweeper)) {
	registerResourceYandexKmsSymmetricKey(add)
	registerYandexKmsAsymmetricEncryptionKeyResource(add)
	registerYandexKmsAsymmetricSignatureKeyResource(add)
}
