package sweep

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	cloudPrefix = "tfacc"

	// Delete can last up to 30 minutes, approved by IAM.
	yandexResourceManagerCloudDeleteTimeout = 30 * time.Minute
)

func registerBillingCloudBindingResoursemanagerCloud(add func(string, *resource.Sweeper)) {
	add("yandex_resourcemanager_cloud", &resource.Sweeper{
		Name:         "yandex_resourcemanager_cloud",
		F:            sweepAllClouds,
		Dependencies: []string{},
	})
}

func sweepAllClouds(string) error {
	if os.Getenv("YC_ENABLE_CLOUD_SWEEPING") != "1" {
		return nil
	}

	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	it := resourcemanagersdk.NewCloudClient(conf.SDKv2).Iterator(
		context.Background(),
		&resourcemanager.ListCloudsRequest{
			OrganizationId: conf.ProviderState.OrganizationID.ValueString(),
		},
	)

	result := &multierror.Error{}
	for it.Next() {
		if !strings.HasPrefix(it.Value().Name, cloudPrefix) {
			continue
		}

		id := it.Value().GetId()
		if !test.SweepWithRetry(sweepCloudOnce, conf, "Cloud", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Cloud %q", id))
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

func sweepCloudOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexResourceManagerCloudDeleteTimeout)
	defer cancel()

	op, err := resourcemanagersdk.NewCloudClient(conf.SDKv2).Delete(ctx, &resourcemanager.DeleteCloudRequest{
		CloudId:     id,
		DeleteAfter: timestamppb.Now(),
	})

	return test.HandleSweepOperation(ctx, op, err)
}
