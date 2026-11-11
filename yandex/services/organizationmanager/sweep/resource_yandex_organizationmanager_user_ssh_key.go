package sweep

import (
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexOrganizationmanagerUserSshKey(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_user_ssh_key", &resource.Sweeper{
		Name:         "yandex_organizationmanager_user_ssh_key",
		F:            sweepAllUserSshKeys,
		Dependencies: []string{},
	})
}

func sweepAllUserSshKeys(_ string) error {
	return sweepAllUserSshKeysForOrganization(sweepOrganizationID())
}

func sweepAllUserSshKeysForOrganization(organizationID string) error {
	if organizationID == "" {
		return nil
	}

	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	userID, err := loginToUserID(conf, os.Getenv("YC_LOGIN"))
	if err != nil {
		return err
	}
	req := &organizationmanager.ListUserSshKeysRequest{
		OrganizationId: organizationID,
		SubjectId:      userID,
	}
	client := organizationmanagersdk.NewUserSshKeyClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweep.SweepWithRetry(sweepUserSshKeyOnce, conf, "UserSshKey", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep UserSshKey %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepUserSshKeyOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()
	client := organizationmanagersdk.NewUserSshKeyClient(conf.SDK)

	op, err := client.Delete(ctx, &organizationmanager.DeleteUserSshKeyRequest{
		UserSshKeyId: id,
	})

	return sweep.HandleSweepOperationV2(ctx, op, err)
}
