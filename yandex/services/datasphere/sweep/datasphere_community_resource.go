package sweep

import (
	"context"
	"fmt"
	"time"

	dataspheresdk "github.com/yandex-cloud/go-sdk/services/datasphere/v2"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datasphere/v2"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerDatasphereCommunityResource(add func(string, *resource.Sweeper)) {
	add("yandex_datasphere_community", &resource.Sweeper{
		Name:         "yandex_datasphere_community",
		F:            sweepAllCommunity,
		Dependencies: []string{},
	})
}

func sweepAllCommunity(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	it := dataspheresdk.NewCommunityClient(conf.SDKv2).Iterator(
		context.Background(),
		&datasphere.ListCommunitiesRequest{OrganizationId: test.GetExampleOrganizationID()},
	)
	result := &multierror.Error{}

	for it.Next() {
		communityId := it.Value().GetId()
		if !test.IsTestResourceName(it.Value().GetName()) {
			continue
		}
		if !sweepCommunity(conf, communityId) {
			result = multierror.Append(
				result,
				fmt.Errorf("failed to sweep community id %q", communityId),
			)
		}
	}

	if err := it.Error(); err != nil {
		result = multierror.Append(
			result,
			fmt.Errorf("iterator error: %w", err),
		)
	}

	return result.ErrorOrNil()
}

func sweepCommunity(conf *provider_config.Config, cloudId string) bool {
	return test.SweepWithRetry(sweepCommunityOnce, conf, "yandex_datasphere_community", cloudId)
}

func sweepCommunityOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	op, err := dataspheresdk.NewCommunityClient(conf.SDKv2).Delete(ctx, &datasphere.DeleteCommunityRequest{
		CommunityId: id,
	})
	return test.HandleSweepOperation(ctx, op, err)
}
