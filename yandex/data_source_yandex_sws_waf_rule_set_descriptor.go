package yandex

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	waf "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1/waf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func dataSourceYandexSmartwebsecurityWafRuleSetDescriptor() *schema.Resource {
	return &schema.Resource{
		Description: "Get information about WAF rule sets. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#rules-set).\n\nThis data source is used to get list of rules that can be used by `yandex_sws_waf_profile`.\n\n",
		ReadContext: dataSourceYandexSmartwebsecurityWafRuleSetDescriptorRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the rule set.",
				Optional:    true,
				Computed:    true,
			},

			"rule_set_descriptor_id": {
				Type:        schema.TypeString,
				Description: "ID of the rule set.",
				Optional:    true,
			},

			"rules": {
				Type:        schema.TypeList,
				Description: "List of rules.\n  * `anomaly_score` (Number) Numeric anomaly value, i.e., a potential attack indicator. The higher this value, the more likely it is that the request that satisfies the rule is an attack. See [documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#anomaly).\n  * `paranoia_level` (Number) Paranoia level classifies rules according to their aggression. The higher the paranoia level, the better your protection, but also the higher the probability of WAF false positives. See [documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#paranoia).\n  * `id` (String) The rule ID.\n",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"anomaly_score": {
							Type:        schema.TypeInt,
							Description: "Numeric anomaly value, i.e., a potential attack indicator. The higher this value, the more likely it is that the request that satisfies the rule is an attack. See [documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#anomaly).",
							Computed:    true,
						},

						"id": {
							Type:        schema.TypeString,
							Description: "The rule ID.",
							Computed:    true,
						},

						"paranoia_level": {
							Type:        schema.TypeInt,
							Description: "Paranoia level classifies rules according to their aggression. The higher the paranoia level, the better your protection, but also the higher the probability of WAF false positives. See [documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#paranoia).",
							Computed:    true,
						},
					},
				},
				Computed: true,
			},

			"version": {
				Type:        schema.TypeString,
				Description: "Version of the rule set.",
				Required:    true,
			},
		},
	}
}

func dataSourceYandexSmartwebsecurityWafRuleSetDescriptorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Config)

	req := &waf.GetRuleSetDescriptorRequest{
		Name:    d.Get("name").(string),
		Version: d.Get("version").(string),
	}

	log.Printf("[DEBUG] Read RuleSetDescriptor request: %s", protoDump(req))

	md := new(metadata.MD)
	resp, err := config.sdk.SmartWebSecurityWaf().RuleSetDescriptor().Get(ctx, req, grpc.Header(md))
	if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read RuleSetDescriptor x-server-trace-id: %s", traceHeader[0])
	}
	if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
		log.Printf("[DEBUG] Read RuleSetDescriptor x-server-request-id: %s", traceHeader[0])
	}
	if err != nil {
		return diag.FromErr(handleNotFoundError(err, d, fmt.Sprintf("rule_set_descriptor %q", d.Get("rule_set_descriptor_id").(string))))
	}

	d.SetId(resp.Id)

	log.Printf("[DEBUG] Read RuleSetDescriptor response: %s", protoDump(resp))

	rules, err := flattenWafRulesSlice(resp.GetRules())
	if err != nil { // isElem: false, ret: 1
		return diag.FromErr(err)
	}

	if err := d.Set("name", resp.GetName()); err != nil {
		log.Printf("[ERROR] failed set field name: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("rule_set_descriptor_id", resp.GetId()); err != nil {
		log.Printf("[ERROR] failed set field rule_set_descriptor_id: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("rules", rules); err != nil {
		log.Printf("[ERROR] failed set field rules: %s", err)
		return diag.FromErr(err)
	}
	if err := d.Set("version", resp.GetVersion()); err != nil {
		log.Printf("[ERROR] failed set field version: %s", err)
		return diag.FromErr(err)
	}

	return nil
}
