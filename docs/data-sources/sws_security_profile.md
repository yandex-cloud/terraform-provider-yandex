---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about SmartWebSecurity Profile.
---

# {{.Name}} ({{.Type}})

Get information about SecurityProfile. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/profiles).

This data source is used to define SecurityProfile that can be used by other resources.

## Example usage

{{ tffile "examples/sws_security_profile/d_sws_security_profile_1.tf" }}

{{ tffile "examples/sws_security_profile/d_sws_security_profile_2.tf" }}


## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the security profile.
* `security_profile_id` - (Optional) ID of the security profile.

~> One of `security_profile_id` or `name` should be specified.

## Attributes Reference

The following attributes are exported:

* `id` - ID of the profile.

* `created_at` - The profile creation timestamp.

* `name` - Name of the security profile. The name is unique within the folder. 1-50 characters long.

* `labels` - Labels as key:value pairs. Maximum of 64 per resource.

* `description` - Optional description of the security profile.

* `default_action` - Action to perform if none of rules matched. Possible values: `ALLOW` or `DENY`.

* `captcha_id` - Captcha ID to use with this security profile. Set empty to use default.

* `advanced_rate_limiter_profile_id` - Advanced rate limiter profile ID to use with this security profile. Set empty to use default.

* `security_rule` - List of security rules. The structure is documented below.

---

The `security_rule` block supports:

* `name` - Name of the rule. The name is unique within the security profile. 1-50 characters long.

* `priority` - Determines the priority for checking the incoming traffic.

* `dry_run` - This mode allows you to test your security profile or a single rule.

* `description` - Optional description of the rule. 0-512 characters long.

* `smart_protection` - Smart Protection rule, see [Smart Protection rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#smart-protection-rules). The structure is documented below.

* `rule_condition` - Rule actions, see [Rule actions](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#rule-action). The structure is documented below.

* `waf` - Web Application Firewall (WAF) rule, see [WAF rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#waf-rules). The structure is documented below.

~> Exactly one rule specifier: `smart_protection` or `rule_condition` or `waf` should be specified.

---

The `rule_condition` block supports:

* `action` - Action to perform if this rule matched. Possible values: `ALLOW` or `DENY`.

* `condition` - The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

---

The `smart_protection` block supports:

* `mode` - Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).

* `condition` - The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

---

The `waf` block supports:

* `mode` - Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).

* `condition` - The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

* `waf_profile_id` - ID of WAF profile to use in this rule.
