---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: yandex_sws_waf_profile"
description: |-
  Manage a Web Application Firewall in Yandex Cloud.
---

# yandex_sws_waf_profile (Resource)

Creates a WAF Profile in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/quickstart#waf).

## Example usage

```terraform
//
// Create a new SWS WAF Profile (Empty).
//
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
//
// Create a new SWS WAF Profile (Default).
//
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

* `name` - (Required) Name of the WAF profile. The name is unique within the folder. 1-50 characters long.

* `folder_id` - (Optional) ID of the folder to create a profile in. If omitted, the provider folder is used.

* `labels` - (Optional) Labels as key:value pairs. Maximum of 64 per resource.

* `description` - (Optional) Optional description of the WAF profile.

* `rule` - (Optional) Settings for each rule in rule set. The structure is documented below.

* `exclusion_rule` - (Optional) List of exclusion rules. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#exclusion-rules). The structure is documented below.

* `core_rule_set` - (Required) Core rule set settings. See [Basic rule set](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#rules-set) for details. The structure is documented below.

* `analyze_request_body` - (Optional) Parameters for request body analyzer. The structure is documented below.

---

The `rule` block supports:

* `rule_id` - (Required) Rule ID.

* `is_enabled` - (Optional) Determines is it rule enabled or not.

* `is_blocking` - (Optional) Determines is it rule blocking or not.

---

The `exclusion_rule` block supports:

* `name` - (Required) Name of exclusion rule.

* `description` - (Optional) Optional description of the rule. 0-512 characters long.

* `exclude_rules` - (Optional) Exclude rules. The structure is documented below.

* `log_excluded` - (Optional) Records the fact that an exception rule is triggered.

---

The `exclude_rules` block supports:

* `exclude_all` - (Optional) Set this option true to exclude all rules.

* `rule_ids` - (Optional) List of rules to exclude.

---

The `core_rule_set` block supports:

* `inbound_anomaly_score` - (Required) Anomaly score. Enter an integer within the range of 2 and 10000. The higher this value, the more likely it is that the request that satisfies the rule is an attack. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#anomaly) for more details.

* `paranoia_level` - (Required) Paranoia level. Enter an integer within the range of 1 and 4. Paranoia level classifies rules according to their aggression. The higher the paranoia level, the better your protection, but also the higher the probability of WAF false positives. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/waf#paranoia) for more details. NOTE: this option has no effect on enabling or disabling rules, it is used only as recommendation for user to enable all rules with paranoia_level <= this value.

---

The `analyze_request_body` block supports:

* `is_enabled` - (Optional) Possible to turn analyzer on and turn if off.

* `size_limit` - (Required) Maximum size of body to pass to analyzer. In kilobytes.

* `size_limit_action` - (Required) Action to perform if maximum size of body exceeded. Possible values: `IGNORE` and `DENY`.

---

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - ID of the WAF profile.

* `created_at` - The WAF Profile creation timestamp.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_sws_waf_profile.<resource Name> <resource Id>
terraform import yandex_sws_waf_profile.default ...
```
