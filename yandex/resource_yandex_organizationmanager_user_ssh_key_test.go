package yandex

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
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
	client := organizationmanagersdk.NewUserSshKeyClient(conf.SDK)

	op, err := client.Delete(ctx, &organizationmanager.DeleteUserSshKeyRequest{
		UserSshKeyId: id,
	})

	return handleSweepOperationV2(ctx, op, err)
}

func testSweepUserSshKeys(_ string) error {
	return testSweepUserSshKeysForOrganization(getExampleOrganizationID())
}

func testSweepUserSshKeysForOrganization(organizationID string) error {
	if organizationID == "" {
		return nil
	}

	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &organizationmanager.ListUserSshKeysRequest{
		OrganizationId: organizationID,
		SubjectId:      getExampleUserID1(),
	}
	client := organizationmanagersdk.NewUserSshKeyClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepWithRetry(testSweepUserSshKeyOnce, conf, "UserSshKey", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep UserSshKey %q", id))
		}
	}

	return result.ErrorOrNil()
}

func TestSweepUserSshKeysWithoutOrganizationIDIsNoop(t *testing.T) {
	t.Setenv("YC_TOKEN", "")
	t.Setenv("YC_SERVICE_ACCOUNT_KEY_FILE", "")

	if err := testSweepUserSshKeysForOrganization(""); err != nil {
		t.Fatalf("sweeper without organization ID must be a no-op: %v", err)
	}
}
