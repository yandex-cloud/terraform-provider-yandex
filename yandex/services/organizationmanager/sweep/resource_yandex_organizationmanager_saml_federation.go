package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/saml"
	samlsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/saml"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/pluginsdk"
)

func registerResourceYandexOrganizationmanagerSamlFederation(add func(string, *resource.Sweeper)) {
	add("yandex_organizationmanager_saml_federation", &resource.Sweeper{
		Name:         "yandex_organizationmanager_saml_federation",
		F:            sweepAllSamlFederations,
		Dependencies: []string{},
	})
}

func sweepAllSamlFederations(_ string) error {
	organizationID := sweepOrganizationID()
	if organizationID == "" {
		return nil
	}

	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &saml.ListFederationsRequest{
		OrganizationId: organizationID,
	}
	client := samlsdk.NewFederationClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweep.SweepWithRetry(sweepSamlFederationOnce, conf, "SAML Federation", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep SAML Federation %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepSamlFederationOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_organization_manager.YandexOrganizationManagerSamlFederationDefaultTimeout)
	defer cancel()

	client := samlsdk.NewFederationClient(conf.SDK)

	op, err := client.Delete(ctx, &saml.DeleteFederationRequest{
		FederationId: id,
	})

	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
