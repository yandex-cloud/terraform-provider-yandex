---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about SmartWebSecurity WAF rule sets.
---

# {{.Name}} ({{.Type}})

Get information about WAF rule sets. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#rules-set).

## Example usage

{{ tffile "examples/sws_waf_rule_set_descriptor/d_sws_waf_rule_set_descriptor_1.tf" }}

This data source is used to get list of rules that can be used by `yandex_sws_waf_profile`.

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the rule set.
* `version` - (Required) Version of the rule set.

## Attributes Reference

The following attributes are exported:

* `id` - ID of the rule set.
* `rules` - List of rules. The structure is documented below.

---

The `rules` block supports:

* `id` - Rule ID.

* `anomaly_score` - Numeric anomaly value, i.e., a potential attack indicator. The higher this value, the more likely it is that the request that satisfies the rule is an attack. See [documentation](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#anomaly).

* `paranoia_level` - Paranoia level classifies rules according to their aggression. The higher the paranoia level, the better your protection, but also the higher the probability of WAF false positives. See [documentation](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#paranoia).
