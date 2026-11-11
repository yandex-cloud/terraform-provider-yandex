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
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
)

func registerDatasphereProjectResource(add func(string, *resource.Sweeper)) {
	add("yandex_datasphere_project", &resource.Sweeper{
		Name:         "yandex_datasphere_project",
		F:            sweepAllProject,
		Dependencies: []string{},
	})
}

func sweepAllProject(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	// get all communities for user
	communities, err := getAllCommunityIDs(conf)
	if err != nil {
		return fmt.Errorf("get communities: %w", err)
	}

	result := &multierror.Error{}

	for _, communityID := range communities {
		it := dataspheresdk.NewProjectClient(conf.SDKv2).Iterator(
			context.Background(),
			&datasphere.ListProjectsRequest{
				CommunityId: communityID,
				OwnedById:   test.AccTestsUser,
			},
		)

		for it.Next() {
			projectId := it.Value().GetId()
			if !test.IsTestResourceName(it.Value().GetName()) {
				continue
			}
			if !sweepProject(conf, projectId) {
				result = multierror.Append(
					result,
					fmt.Errorf("failed to sweep project id %q", projectId),
				)
			}
		}

		if err := it.Error(); err != nil {
			result = multierror.Append(
				result,
				fmt.Errorf("iterator error: %w", err),
			)
		}
	}

	return result.ErrorOrNil()
}

func getAllCommunityIDs(conf *provider_config.Config) ([]string, error) {
	var (
		it = dataspheresdk.NewCommunityClient(conf.SDKv2).Iterator(
			context.Background(),
			&datasphere.ListCommunitiesRequest{OrganizationId: test.GetExampleOrganizationID()},
		)
		ids []string
	)

	for it.Next() {
		if !test.IsTestResourceName(it.Value().GetName()) {
			continue
		}
		ids = append(ids, it.Value().GetId())
	}

	if err := it.Error(); err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("iterator err: %w", err)
	}

	return ids, nil
}

func sweepProject(conf *provider_config.Config, cloudId string) bool {
	return test.SweepWithRetry(sweepProjectOnce, conf, "yandex_datasphere_project", cloudId)
}

func sweepProjectOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)

	defer cancel()

	op, err := dataspheresdk.NewProjectClient(conf.SDKv2).Delete(ctx, &datasphere.DeleteProjectRequest{
		ProjectId: id,
	})
	return test.HandleSweepOperation(ctx, op, err)
}
