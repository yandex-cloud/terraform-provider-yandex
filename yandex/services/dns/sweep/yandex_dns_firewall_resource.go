package sweep

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	dns "github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	dnsv1sdk "github.com/yandex-cloud/go-sdk/services/dns/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const yandexDnsFirewallDefaultTimeout = 5 * time.Minute

func registerYandexDnsFirewallResource(add func(string, *resource.Sweeper)) {
	add("yandex_dns_firewall", &resource.Sweeper{
		Name:         "yandex_dns_firewall",
		F:            sweepAllDnsFirewall,
		Dependencies: []string{},
	})
}

func sweepAllDnsFirewall(_ string) error {
	conf, err := testhelpers.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := dnsv1sdk.NewDnsFirewallClient(conf.SDKv2).List(context.Background(), &dns.ListDnsFirewallsRequest{
		FolderId: conf.ProviderState.FolderID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error listing DNS firewalls: %s", err)
	}

	result := &multierror.Error{}
	for _, fw := range resp.DnsFirewalls {
		if !sweepDnsFirewall(conf, fw.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep DNS Firewall %q", fw.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepDnsFirewall(conf *provider_config.Config, id string) bool {
	return testhelpers.SweepWithRetry(sweepDnsFirewallOnce, conf, "yandex_dns_firewall", id)
}

func sweepDnsFirewallOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), yandexDnsFirewallDefaultTimeout)
	defer cancel()

	op, err := dnsv1sdk.NewDnsFirewallClient(conf.SDKv2).Delete(ctx, &dns.DeleteDnsFirewallRequest{
		DnsFirewallId: id,
	})
	if err != nil {
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
