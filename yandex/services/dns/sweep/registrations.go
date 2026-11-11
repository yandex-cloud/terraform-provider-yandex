package sweep

import "github.com/hashicorp/terraform-plugin-testing/helper/resource"

func Register(add func(string, *resource.Sweeper)) {
	registerYandexDnsFirewallResource(add)
	registerYandexDnsFirewallIamBindingResource(add)
}
