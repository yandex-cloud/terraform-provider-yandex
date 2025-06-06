---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: yandex_sws_waf_rule_set_descriptor"
description: |-
  Get information about SmartWebSecurity WAF rule sets.
---

# yandex_sws_waf_rule_set_descriptor (Data Source)

Get information about WAF rule sets. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#rules-set).

This data source is used to get list of rules that can be used by `yandex_sws_waf_profile`.

## Example usage

```terraform
//
// Get information about existing SWS WAF Rule Descriptor
//
data "yandex_sws_waf_rule_set_descriptor" "owasp4" {
  name    = "OWASP Core Ruleset"
  version = "4.0.0"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `version` (String) Version of the rule set.

### Optional

- `name` (String) Name of the rule set.
- `rule_set_descriptor_id` (String) ID of the rule set.

### Read-Only

- `id` (String) The ID of this resource.
- `rules` (List of Object) List of rules.
  * `anomaly_score` (Number) Numeric anomaly value, i.e., a potential attack indicator. The higher this value, the more likely it is that the request that satisfies the rule is an attack. See [documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#anomaly).
  * `paranoia_level` (Number) Paranoia level classifies rules according to their aggression. The higher the paranoia level, the better your protection, but also the higher the probability of WAF false positives. See [documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/waf#paranoia).
  * `id` (String) The rule ID. (see [below for nested schema](#nestedatt--rules))

<a id="nestedatt--rules"></a>
### Nested Schema for `rules`

Read-Only:

- `anomaly_score` (Number)
- `id` (String)
- `paranoia_level` (Number)
