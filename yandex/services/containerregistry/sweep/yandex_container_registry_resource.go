package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	containerregistrysdk "github.com/yandex-cloud/go-sdk/services/containerregistry/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerYandexContainerRegistryResource(add func(string, *resource.Sweeper)) {
	add("yandex_container_registry", &resource.Sweeper{
		Name:         "yandex_container_registry",
		F:            sweepAllContainerRegistry,
		Dependencies: []string{},
	})
}

func sweepAllContainerRegistry(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &containerregistry.ListRegistriesRequest{FolderId: conf.ProviderState.FolderID.ValueString()}
	resp, err := containerregistrysdk.NewRegistryClient(conf.SDKv2).List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting clouds: %s", err)
	}
	result := &multierror.Error{}
	for _, registry := range resp.Registries {
		id := registry.GetId()
		if !test.SweepWithRetry(sweepContainerRegistryOnce, conf, "Container Registry", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Container Registry %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepContainerRegistryOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	op, err := containerregistrysdk.NewRegistryClient(conf.SDKv2).Delete(ctx, &containerregistry.DeleteRegistryRequest{RegistryId: id})
	_, err = op.Wait(ctx)
	return err
}
