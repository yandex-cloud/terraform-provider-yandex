package yandex

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	smartwebsecurity "github.com/yandex-cloud/go-genproto/yandex/cloud/smartwebsecurity/v1"
	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
)

const yandexSWSSecurityProfileWAFAttachmentDefaultTimeout = 5 * time.Minute

func resourceYandexSWSSecurityProfileWAFProfileAttachment() *schema.Resource {
	return &schema.Resource{
		Description: "Tracks a WAF rule attachment in a Smart Web Security profile and removes the rule before the WAF profile is deleted. The WAF rule itself must be configured in `yandex_sws_security_profile`.",
		Create:      resourceYandexSWSSecurityProfileWAFAttachmentCreate,
		Read:        resourceYandexSWSSecurityProfileWAFAttachmentRead,
		Delete:      resourceYandexSWSSecurityProfileWAFAttachmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(yandexSWSSecurityProfileWAFAttachmentDefaultTimeout),
			Read:   schema.DefaultTimeout(yandexSWSSecurityProfileWAFAttachmentDefaultTimeout),
			Delete: schema.DefaultTimeout(yandexSWSSecurityProfileWAFAttachmentDefaultTimeout),
		},
		Schema: map[string]*schema.Schema{
			"security_profile_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Smart Web Security profile.",
			},
			"waf_profile_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the WAF profile.",
			},
			"security_rule_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the WAF security rule managed by `yandex_sws_security_profile`.",
			},
		},
	}
}

func resourceYandexSWSSecurityProfileWAFAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	securityProfileID := d.Get("security_profile_id").(string)
	wafProfileID := d.Get("waf_profile_id").(string)
	securityRuleName := d.Get("security_rule_name").(string)

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutCreate))
	defer cancel()

	profile, err := getSWSSecurityProfile(ctx, config, securityProfileID)
	if err != nil {
		return err
	}
	if !hasSWSWAFRule(profile, securityRuleName, wafProfileID) {
		return fmt.Errorf("Smart Web Security profile %q has no WAF rule %q attached to WAF profile %q", securityProfileID, securityRuleName, wafProfileID)
	}

	d.SetId(makeSWSSecurityProfileWAFAttachmentID(securityProfileID, wafProfileID, securityRuleName))
	return resourceYandexSWSSecurityProfileWAFAttachmentRead(d, meta)
}

func resourceYandexSWSSecurityProfileWAFAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	securityProfileID, wafProfileID, securityRuleName, err := swsSecurityProfileWAFAttachmentIDs(d)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutRead))
	defer cancel()

	profile, err := getSWSSecurityProfile(ctx, config, securityProfileID)
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			d.SetId("")
			return nil
		}
		return err
	}
	if !hasSWSWAFRule(profile, securityRuleName, wafProfileID) {
		d.SetId("")
		return nil
	}

	if err := d.Set("security_profile_id", securityProfileID); err != nil {
		return err
	}
	if err := d.Set("waf_profile_id", wafProfileID); err != nil {
		return err
	}
	return d.Set("security_rule_name", securityRuleName)
}

func resourceYandexSWSSecurityProfileWAFAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	securityProfileID, wafProfileID, securityRuleName, err := swsSecurityProfileWAFAttachmentIDs(d)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(config.Context(), d.Timeout(schema.TimeoutDelete))
	defer cancel()

	profile, err := getSWSSecurityProfile(ctx, config, securityProfileID)
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
			d.SetId("")
			return nil
		}
		return err
	}

	rules := make([]*smartwebsecurity.SecurityRule, 0, len(profile.SecurityRules))
	removed := false
	for _, rule := range profile.SecurityRules {
		if rule.Name == securityRuleName && rule.GetWaf().GetWafProfileId() == wafProfileID {
			removed = true
			continue
		}
		rules = append(rules, rule)
	}
	if removed {
		op, err := config.sdk.WrapOperation(config.sdk.SmartWebSecurity().SecurityProfile().Update(ctx, &smartwebsecurity.UpdateSecurityProfileRequest{
			SecurityProfileId: securityProfileID,
			SecurityRules:     rules,
			UpdateMask: &field_mask.FieldMask{
				Paths: []string{"security_rules"},
			},
		}))
		if err != nil {
			return fmt.Errorf("detaching WAF rule %q from Smart Web Security profile %q: %w", securityRuleName, securityProfileID, err)
		}
		if err := op.Wait(ctx); err != nil {
			return fmt.Errorf("waiting for WAF rule %q detachment from Smart Web Security profile %q: %w", securityRuleName, securityProfileID, err)
		}
	}

	d.SetId("")
	return nil
}

func getSWSSecurityProfile(ctx context.Context, config *Config, securityProfileID string) (*smartwebsecurity.SecurityProfile, error) {
	profile, err := config.sdk.SmartWebSecurity().SecurityProfile().Get(ctx, &smartwebsecurity.GetSecurityProfileRequest{
		SecurityProfileId: securityProfileID,
	})
	if err != nil {
		return nil, fmt.Errorf("reading Smart Web Security profile %q: %w", securityProfileID, err)
	}
	return profile, nil
}

func hasSWSWAFRule(profile *smartwebsecurity.SecurityProfile, securityRuleName, wafProfileID string) bool {
	for _, rule := range profile.SecurityRules {
		if rule.Name == securityRuleName && rule.GetWaf().GetWafProfileId() == wafProfileID {
			return true
		}
	}
	return false
}

func makeSWSSecurityProfileWAFAttachmentID(securityProfileID, wafProfileID, securityRuleName string) string {
	return strings.Join([]string{securityProfileID, wafProfileID, securityRuleName}, "/")
}

func swsSecurityProfileWAFAttachmentIDs(d *schema.ResourceData) (string, string, string, error) {
	securityProfileID, _ := d.Get("security_profile_id").(string)
	wafProfileID, _ := d.Get("waf_profile_id").(string)
	securityRuleName, _ := d.Get("security_rule_name").(string)
	if securityProfileID != "" && wafProfileID != "" && securityRuleName != "" {
		return securityProfileID, wafProfileID, securityRuleName, nil
	}

	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("invalid attachment ID %q: expected <security_profile_id>/<waf_profile_id>/<security_rule_name>", d.Id())
	}
	return parts[0], parts[1], parts[2], nil
}
