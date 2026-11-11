package sweep

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	idp "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp"
	idpsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const (
	userpoolSweepPageSizeIdpUserpool  = 1000
	userpoolSweepDeleteTimeout        = 15 * time.Minute
	testResourceNamePrefixIdpUserpool = "tf-acc-test-userpool"
)

func registerYandexOrganizationmanagerIdpUserpoolResource(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_idp_userpool", &resource.Sweeper{
		Name:         "yandex_organizationmanager_idp_userpool",
		F:            sweepAllIdpUserpool,
		Dependencies: []string{},
	})
}

func sweepAllIdpUserpool(_ string) error {
	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	organizationID := test.GetExampleOrganizationID()
	if organizationID == "" {
		log.Printf("[WARN] organization ID is not set, skipping userpool sweep")
		return nil
	}

	req := &idp.ListUserpoolsRequest{
		OrganizationId: organizationID,
		PageSize:       userpoolSweepPageSizeIdpUserpool,
	}

	client := idpsdk.NewUserpoolClient(conf.SDKv2)
	resp, err := client.List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting list of userpools: %s", err)
	}

	result := &multierror.Error{}
	for _, userpool := range resp.Userpools {
		if strings.HasPrefix(userpool.Name, testResourceNamePrefixIdpUserpool) {
			if !sweepIdpUserpool(conf, userpool.Id) {
				result = multierror.Append(result, fmt.Errorf("failed to sweep Idp Userpool %q", userpool.Id))
			}
		}
	}

	// Handle pagination if needed
	for resp.NextPageToken != "" {
		req.PageToken = resp.NextPageToken
		resp, err = client.List(context.Background(), req)
		if err != nil {
			result = multierror.Append(result, fmt.Errorf("error getting next page of userpools: %s", err))
			break
		}

		for _, userpool := range resp.Userpools {
			if strings.HasPrefix(userpool.Name, testResourceNamePrefixIdpUserpool) {
				if !sweepIdpUserpool(conf, userpool.Id) {
					result = multierror.Append(result, fmt.Errorf("failed to sweep Idp Userpool %q", userpool.Id))
				}
			}
		}
	}

	return result.ErrorOrNil()
}

func sweepIdpUserpool(conf *provider_config.Config, id string) bool {
	return test.SweepWithRetry(sweepIdpUserpoolOnce, conf, "Idp Userpool", id)
}

func sweepIdpUserpoolOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), userpoolSweepDeleteTimeout)
	defer cancel()

	client := idpsdk.NewUserpoolClient(conf.SDKv2)
	op, err := client.Delete(ctx, &idp.DeleteUserpoolRequest{
		UserpoolId: id,
	})
	if err != nil {
		return err
	}

	_, err = op.Wait(ctx)
	return err
}
