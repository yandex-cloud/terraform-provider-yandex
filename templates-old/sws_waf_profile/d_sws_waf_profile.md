---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about SmartWebSecurity WAF Profile.
---

# {{.Name}} ({{.Type}})

Get information about WAF Profile. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/quickstart#waf).

## Example usage

{{ tffile "examples/sws_waf_profile/d_sws_waf_profile_1.tf" }}

This data source is used to define WAF Profile that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the WAF profile.
* `waf_profile_id` - (Optional) ID of the WAF profile.

~> One of `waf_profile_id` or `name` should be specified.

## Attributes Reference

The following attributes are exported:

* `id` - ID of the profile.

* `created_at` - The profile creation timestamp.

* `name` - Name of the WAF profile. The name is unique within the folder. 1-50 characters long.

* `folder_id` - ID of the folder to create a profile in. If omitted, the provider folder is used.

* `labels` - Labels as key:value pairs. Maximum of 64 per resource.

* `description` - Optional description of the WAF profile.

* `rule` - Settings for each rule in rule set. The structure is documented below.

* `exclusion_rule` - List of exclusion rules. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#exclusion-rules). The structure is documented below.

* `core_rule_set` - Core rule set settings. See [Basic rule set](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#rules-set) for details. The structure is documented below.

* `analyze_request_body` - Parameters for request body analyzer. The structure is documented below.

---

The `rule` block supports:

* `rule_id` - Rule ID.

* `is_enabled` - Determines is it rule enabled or not.

* `is_blocking` - Determines is it rule blocking or not.

---

The `exclusion_rule` block supports:

* `name` - Name of exclusion rule.

* `description` - Optional description of the rule. 0-512 characters long.

* `exclude_rules` - Exclude rules. The structure is documented below.

* `log_excluded` - Records the fact that an exception rule is triggered.

---

The `exclude_rules` block supports:

* `exclude_all` - Set this option true to exclude all rules.

* `rule_ids` - List of rules to exclude.

---

The `core_rule_set` block supports:

* `inbound_anomaly_score` - Anomaly score. Enter an integer within the range of 2 and 10000. The higher this value, the more likely it is that the request that satisfies the rule is an attack. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#anomaly) for more details.

* `paranoia_level` - Paranoia level. Enter an integer within the range of 1 and 4. Paranoia level classifies rules according to their aggression. The higher the paranoia level, the better your protection, but also the higher the probability of WAF false positives. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#paranoia) for more details. NOTE: this option has no effect on enabling or disabling rules, it is used only as recommendation for user to enable all rules with paranoia_level <= this value.

---

The `analyze_request_body` block supports:

* `is_enabled` - Possible to turn analyzer on and turn if off.

* `size_limit` - Maximum size of body to pass to analyzer. In kilobytes.

* `size_limit_action` - Action to perform if maximum size of body exceeded. Possible values: `IGNORE` and `DENY`.
