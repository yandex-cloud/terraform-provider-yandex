package pluginsdk

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
)

func expandContainerRepositoryLifecyclePolicyRules(d *schema.ResourceData) ([]*containerregistry.LifecycleRule, error) {
	var rules []*containerregistry.LifecycleRule

	for _, key := range sdkcommon.IterateKeys(d, "rule") {
		rule := &containerregistry.LifecycleRule{}

		if description, ok := d.GetOk(key + "description"); ok {
			rule.SetDescription(description.(string))
		}

		if expirePeriod, ok := d.GetOk(key + "expire_period"); ok {
			duration, err := sdkcommon.ParseDuration(expirePeriod.(string))
			if err != nil {
				return nil, err
			}

			rule.SetExpirePeriod(duration)
		}

		if retainedTop, ok := d.GetOk(key + "retained_top"); ok {
			rule.SetRetainedTop(int64(retainedTop.(int)))
		}

		if tagRegexp, ok := d.GetOk(key + "tag_regexp"); ok {
			rule.SetTagRegexp(tagRegexp.(string))
		}

		if untagged, ok := d.GetOk(key + "untagged"); ok {
			rule.SetUntagged(untagged.(bool))
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

func flattenContainerRepositoryLifecyclePolicyRules(lifecycleRules []*containerregistry.LifecycleRule) []interface{} {
	s := make([]interface{}, 0, len(lifecycleRules))

	for _, lifecycleRule := range lifecycleRules {
		rule := FlattenContainerRepositoryLifecyclePolicyRule(lifecycleRule)

		if len(rule) != 0 {
			s = append(s, rule)
		}
	}

	return s
}

func FlattenContainerRepositoryLifecyclePolicyRule(lifecycleRule *containerregistry.LifecycleRule) map[string]interface{} {
	m := make(map[string]interface{})

	m["description"] = lifecycleRule.GetDescription()
	m["expire_period"] = sdkcommon.FormatDuration(lifecycleRule.GetExpirePeriod())
	m["tag_regexp"] = lifecycleRule.GetTagRegexp()
	m["untagged"] = lifecycleRule.GetUntagged()
	m["retained_top"] = lifecycleRule.GetRetainedTop()

	return m
}
