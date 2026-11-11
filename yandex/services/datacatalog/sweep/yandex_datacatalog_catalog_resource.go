package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/datacatalog/v1"
	datacatalogv1sdk "github.com/yandex-cloud/go-sdk/services/datacatalog/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func registerYandexDatacatalogCatalogResource(add func(string, *resource.Sweeper)) {
	add("yandex_datacatalog_catalog", &resource.Sweeper{
		Name:         "yandex_datacatalog_catalog",
		F:            sweepAllCatalog,
		Dependencies: []string{},
	})
}

func sweepAllCatalog(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	resp, err := datacatalogv1sdk.NewCatalogClient(conf.SDKv2).ListCatalogs(context.Background(), &datacatalog.ListCatalogsRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error getting Datacatalog catalog: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Catalogs {
		if !sweepCatalog(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Datacatalog catalog %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepCatalog(conf *provider_config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepCatalogOnce, conf, "yandex_datacatalog_catalog", id)
}

func sweepCatalogOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	op, err := datacatalogv1sdk.NewCatalogClient(conf.SDKv2).DeleteCatalog(ctx, &datacatalog.DeleteCatalogRequest{
		CatalogId: id,
	})
	_, err = op.Wait(context.Background())
	return err
}
