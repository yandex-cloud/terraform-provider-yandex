package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	"sort"
)

func expandCustomDomains(v interface{}) ([]*apigateway.AttachedDomain, error) {
	if v == nil {
		return nil, nil
	}

	domainsList := v.(*schema.Set)
	result := make([]*apigateway.AttachedDomain, domainsList.Len())

	for i, item := range domainsList.List() {
		domain := item.(map[string]interface{})

		result[i] = &apigateway.AttachedDomain{
			DomainId:      domain["domain_id"].(string),
			Domain:        domain["fqdn"].(string),
			CertificateId: domain["certificate_id"].(string),
		}
	}

	return result, nil
}

func flattenCustomDomains(domains []*apigateway.AttachedDomain) []map[string]string {
	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Domain > domains[j].Domain
	})
	result := make([]map[string]string, len(domains))

	for i, domain := range domains {
		result[i] = map[string]string{
			"domain_id":      domain.DomainId,
			"fqdn":           domain.Domain,
			"certificate_id": domain.CertificateId,
		}
	}

	return result
}
