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
		ReadContext: dataSourceYandexSmartwebsecurityWafRuleSetDescriptorRead,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"rule_set_descriptor_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"rules": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"anomaly_score": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"paranoia_level": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Computed: true,
			},

			"version": {
				Type:     schema.TypeString,
				Required: true,
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
