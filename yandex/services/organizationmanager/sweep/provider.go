package sweep

import (
	"fmt"
	"os"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func sweepOrganizationID() string {
	if os.Getenv("TF_ACC") == "" {
		return ""
	}
	return os.Getenv("YC_ORGANIZATION_ID")
}

func loginToUserID(config *providerconfig.Config, login string) (string, error) {
	if login == "" {
		return "", fmt.Errorf("YC_LOGIN is required to sweep user SSH keys")
	}
	client := iamsdk.NewYandexPassportUserAccountClient(config.SDK)
	account, err := client.GetByLogin(config.Context(), &iam.GetUserAccountByLoginRequest{Login: login})
	if err != nil {
		return "", fmt.Errorf("resolve SSH key owner %q: %w", login, err)
	}
	if account.GetId() == "" {
		return "", fmt.Errorf("empty SSH key owner ID for %q", login)
	}
	return account.GetId(), nil
}
