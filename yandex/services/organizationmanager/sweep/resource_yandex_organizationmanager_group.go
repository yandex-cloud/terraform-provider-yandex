package sweep

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1"
	organizationmanagersdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexOrganizationmanagerGroup(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_group", &resource.Sweeper{
		Name:         "yandex_organizationmanager_group",
		F:            sweepAllGroups,
		Dependencies: []string{},
	})
}

func sweepAllGroups(_ string) error {
	return sweepAllGroupsForOrganization(sweepOrganizationID())
}

func sweepAllGroupsForOrganization(organizationID string) error {
	if organizationID == "" {
		return nil
	}

	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &organizationmanager.ListGroupsRequest{
		OrganizationId: organizationID,
	}
	client := organizationmanagersdk.NewGroupClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweep.SweepWithRetry(sweepGroupOnce, conf, "Group", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(1 * time.Minute)
	defer cancel()
	client := organizationmanagersdk.NewGroupClient(conf.SDK)

	op, err := client.Delete(ctx, &organizationmanager.DeleteGroupRequest{
		GroupId: id,
	})

	return sweep.HandleSweepOperationV2(ctx, op, err)
}
