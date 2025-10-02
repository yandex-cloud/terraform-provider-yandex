package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
)

func init() {
	resource.AddTestSweepers("yandex_organizationmanager_user_ssh_key", &resource.Sweeper{
		Name:         "yandex_organizationmanager_user_ssh_key",
		F:            testSweepUserSshKeys,
		Dependencies: []string{},
	})
}

func testSweepUserSshKeyOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()

	op, err := conf.sdk.OrganizationManager().UserSshKey().Delete(ctx, &organizationmanager.DeleteUserSshKeyRequest{
		UserSshKeyId: id,
	})

	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepUserSshKeys(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &organizationmanager.ListUserSshKeysRequest{
		OrganizationId: getExampleOrganizationID(),
		SubjectId:      getExampleUserID1(),
	}
	it := conf.sdk.OrganizationManager().UserSshKey().UserSshKeyIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(testSweepUserSshKeyOnce, conf, "UserSshKey", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep UserSshKey %q", id))
		}
	}

	return result.ErrorOrNil()
}
