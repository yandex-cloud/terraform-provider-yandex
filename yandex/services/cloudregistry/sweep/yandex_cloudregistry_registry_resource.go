package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	cloudregistrysdk "github.com/yandex-cloud/go-sdk/services/cloudregistry/v1"
)

const yandexCloudRegistryDefaultTimeout = 15 * time.Minute

func registerYandexCloudregistryRegistryResource(add func(string, *resource.Sweeper)) {
	add("yandex_cloudregistry_registry", &resource.Sweeper{
		Name:         "yandex_cloudregistry_registry",
		F:            sweepAllCloudRegistry,
		Dependencies: []string{},
	})
}

func sweepAllCloudRegistry(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &cloudregistry.ListRegistriesRequest{FolderId: test.GetExampleFolderID()}
	it := cloudregistrysdk.NewRegistryClient(conf.SDKv2).Iterator(context.Background(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepCloudRegistry(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Cloud Registry %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCloudRegistry(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepcloudRegistryOnce, conf, "Cloud Registry", id)
}

func sweepcloudRegistryOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexCloudRegistryDefaultTimeout)
	defer cancel()

	op, err := cloudregistrysdk.NewRegistryClient(conf.SDKv2).Delete(ctx, &cloudregistry.DeleteRegistryRequest{
		RegistryId: id,
	})
	return test.HandleSweepOperation(ctx, op, err)
}
