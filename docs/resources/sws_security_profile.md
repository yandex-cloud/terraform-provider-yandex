---
subcategory: "SWS (Smart Web Security)"
page_title: "Yandex: yandex_sws_security_profile"
description: |-
  With security profiles you can protect your infrastructure from DDoS attacks at the application level (L7).
---


Creates a Security Profile in the specified folder. For more information, see [the official documentation](https://yandex.cloud/en/docs/smartwebsecurity/concepts/profiles).

# yandex_sws_security_profile




```terraform
resource "yandex_sws_waf_profile" "empty" {
  // NOTE: this WAF profile do not contains any rules enabled.
  // See the next example to see how to enable default set of rules. 
  name = "waf-profile-dummy"

  core_rule_set {
    inbound_anomaly_score = 2
    paranoia_level        = 1
    rule_set {
      name    = "OWASP Core Ruleset"
      version = "4.0.0"
    }
  }
}
```

```terraform
locals {
  waf_paranoia_level = 1
}

data "yandex_sws_waf_rule_set_descriptor" "owasp4" {
  name    = "OWASP Core Ruleset"
  version = "4.0.0"
}

resource "yandex_sws_waf_profile" "default" {
  name = "waf-profile-default"

  core_rule_set {
    inbound_anomaly_score = 2
    paranoia_level        = local.waf_paranoia_level
    rule_set {
      name    = "OWASP Core Ruleset"
      version = "4.0.0"
    }
  }

  dynamic "rule" {
    for_each = [
      for rule in data.yandex_sws_waf_rule_set_descriptor.owasp4.rules : rule
      if rule.paranoia_level >= local.waf_paranoia_level
    ]
    content {
      rule_id     = rule.value.id
      is_enabled  = true
      is_blocking = false
    }
  }

  analyze_request_body {
    is_enabled        = true
    size_limit        = 8
    size_limit_action = "IGNORE"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the security profile. The name is unique within the folder. 1-50 characters long.

* `folder_id` - (Optional) ID of the folder to create a profile in. If omitted, the provider folder is used.

* `labels` - (Optional) Labels as key:value pairs. Maximum of 64 per resource.

* `description` - (Optional) Optional description of the security profile.

* `default_action` - (Required) Action to perform if none of rules matched. Possible values: `ALLOW` or `DENY`.

* `captcha_id` - (Optional) Captcha ID to use with this security profile. Set empty to use default.

* `advanced_rate_limiter_profile_id` - (Optional) Advanced rate limiter profile ID to use with this security profile. Set empty to use default.

* `security_rule` - (Optional) List of security rules. The structure is documented below.

---

The `security_rule` block supports:

* `name` - (Required) Name of the rule. The name is unique within the security profile. 1-50 characters long.

* `priority` - (Required) Determines the priority for checking the incoming traffic.

* `dry_run` - (Optional) This mode allows you to test your security profile or a single rule.

* `description` - (Optional) Optional description of the rule. 0-512 characters long.

* `smart_protection` - (Optional) Smart Protection rule, see [Smart Protection rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#smart-protection-rules). The structure is documented below.

* `rule_condition` - (Optional) Rule actions, see [Rule actions](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#rule-action). The structure is documented below.

* `waf` - (Optional) Web Application Firewall (WAF) rule, see [WAF rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#waf-rules). The structure is documented below.

~> **NOTE:** Exactly one rule specifier: `smart_protection` or `rule_condition` or `waf` should be specified.

---

The `rule_condition` block supports:

* `action` - (Required) Action to perform if this rule matched. Possible values: `ALLOW` or `DENY`.

* `condition` - (Optional) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

---

The `smart_protection` block supports:

* `mode` - (Required) Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).

* `condition` - (Optional) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

---

The `waf` block supports:

* `mode` - (Required) Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).

* `condition` - (Optional) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

* `waf_profile_id` - (Required) ID of WAF profile to use in this rule.

---

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - ID of the security profile.

* `created_at` - The Security Profile creation timestamp.

## Import

A Security Profile can be imported using the `id` of the resource, e.g.:

```
$ terraform import yandex_sws_security_profile.demo-profile-simple security_profile_id
```
