---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: yandex_sws_security_profile"
description: |-
  With security profiles you can protect your infrastructure from DDoS attacks at the application level (L7).
---

# yandex_sws_security_profile (Resource)

With security profiles you can protect your infrastructure from DDoS attacks at the application level (L7).

Creates a Security Profile in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/concepts/profiles).

## Example usage

```terraform
//
// Create a new SWS Security Profile (Simple).
//
resource "yandex_sws_security_profile" "demo-profile-simple" {
  name           = "demo-profile-simple"
  default_action = "ALLOW"

  security_rule {
    name     = "smart-protection"
    priority = 99999

    smart_protection {
      mode = "API"
    }
  }
}
```

```terraform
//
// Create a new SWS Security Profile (Advanced).
//
resource "yandex_sws_security_profile" "demo-profile-advanced" {
  name                             = "demo-profile-advanced"
  default_action                   = "ALLOW"
  captcha_id                       = "<captcha_id>"
  advanced_rate_limiter_profile_id = "<arl_id>"

  security_rule {
    name     = "smart-protection"
    priority = 99999

    smart_protection {
      mode = "API"
    }
  }

  security_rule {
    name     = "waf"
    priority = 88888

    waf {
      mode           = "API"
      waf_profile_id = "<waf_id>"
    }
  }

  security_rule {
    name     = "rule-condition-1"
    priority = 1

    rule_condition {
      action = "ALLOW"

      condition {
        authority {
          authorities {
            exact_match = "example.com"
          }
          authorities {
            exact_match = "example.net"
          }
        }
      }
    }
  }

  security_rule {
    name     = "rule-condition-2"
    priority = 2

    rule_condition {
      action = "DENY"

      condition {
        http_method {
          http_methods {
            exact_match = "DELETE"
          }
          http_methods {
            exact_match = "PUT"
          }
        }
      }
    }
  }

  security_rule {
    name     = "rule-condition-3"
    priority = 3

    rule_condition {
      action = "DENY"

      condition {
        request_uri {
          path {
            prefix_match = "/form"
          }
          queries {
            key = "firstname"
            value {
              pire_regex_match = ".*ivan.*"
            }
          }
          queries {
            key = "lastname"
            value {
              pire_regex_match = ".*petr.*"
            }
          }
        }

        headers {
          name = "User-Agent"
          value {
            pire_regex_match = ".*curl.*"
          }
        }
        headers {
          name = "Referer"
          value {
            pire_regex_not_match = ".*bot.*"
          }
        }

        source_ip {
          ip_ranges_match {
            ip_ranges = ["1.2.33.44", "2.3.4.56"]
          }
          ip_ranges_not_match {
            ip_ranges = ["8.8.0.0/16", "10::1234:1abc:1/64"]
          }
          geo_ip_match {
            locations = ["ru", "es"]
          }
          geo_ip_not_match {
            locations = ["us", "fm", "gb"]
          }
        }
      }
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `advanced_rate_limiter_profile_id` (String) Advanced rate limiter profile ID to use with this security profile. Set empty to use default.
- `captcha_id` (String) Captcha ID to use with this security profile. Set empty to use default.
- `cloud_id` (String) The `Cloud ID` which resource belongs to. If it is not provided, the default provider `cloud-id` is used.
- `default_action` (String) Action to perform if none of rules matched. Possible values: `ALLOW` or `DENY`.
- `description` (String) The resource description.
- `folder_id` (String) The folder identifier that resource belongs to. If it is not provided, the default provider `folder-id` is used.
- `labels` (Map of String) A set of key/value label pairs which assigned to resource.
- `name` (String) The resource name.
- `security_rule` (Block List) List of security rules.

~> Exactly one rule specifier: `smart_protection` or `rule_condition` or `waf` should be specified. (see [below for nested schema](#nestedblock--security_rule))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (String) The creation timestamp of the resource.
- `id` (String) The ID of this resource.

<a id="nestedblock--security_rule"></a>
### Nested Schema for `security_rule`

Optional:

- `description` (String) Optional description of the rule. 0-512 characters long.
- `dry_run` (Boolean) This mode allows you to test your security profile or a single rule.
- `name` (String) Name of the rule. The name is unique within the security profile. 1-50 characters long.
- `priority` (Number) Determines the priority for checking the incoming traffic.
- `rule_condition` (Block List, Max: 1) Rule actions, see [Rule actions](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#rule-action). (see [below for nested schema](#nestedblock--security_rule--rule_condition))
- `smart_protection` (Block List, Max: 1) Smart Protection rule, see [Smart Protection rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#smart-protection-rules). (see [below for nested schema](#nestedblock--security_rule--smart_protection))
- `waf` (Block List, Max: 1) Web Application Firewall (WAF) rule, see [WAF rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/rules#waf-rules). (see [below for nested schema](#nestedblock--security_rule--waf))

<a id="nestedblock--security_rule--rule_condition"></a>
### Nested Schema for `security_rule.rule_condition`

Optional:

- `action` (String) Action to perform if this rule matched. Possible values: `ALLOW` or `DENY`.
- `condition` (Block List, Max: 1) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto). (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition))

<a id="nestedblock--security_rule--rule_condition--condition"></a>
### Nested Schema for `security_rule.rule_condition.condition`

Optional:

- `authority` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--authority))
- `headers` (Block List) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--headers))
- `http_method` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--http_method))
- `request_uri` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--request_uri))
- `source_ip` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--source_ip))

<a id="nestedblock--security_rule--rule_condition--condition--authority"></a>
### Nested Schema for `security_rule.rule_condition.condition.authority`

Optional:

- `authorities` (Block List) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--authority--authorities))

<a id="nestedblock--security_rule--rule_condition--condition--authority--authorities"></a>
### Nested Schema for `security_rule.rule_condition.condition.authority.authorities`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--rule_condition--condition--headers"></a>
### Nested Schema for `security_rule.rule_condition.condition.headers`

Required:

- `value` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--headers--value))

Optional:

- `name` (String)

<a id="nestedblock--security_rule--rule_condition--condition--headers--value"></a>
### Nested Schema for `security_rule.rule_condition.condition.headers.value`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--rule_condition--condition--http_method"></a>
### Nested Schema for `security_rule.rule_condition.condition.http_method`

Optional:

- `http_methods` (Block List) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--http_method--http_methods))

<a id="nestedblock--security_rule--rule_condition--condition--http_method--http_methods"></a>
### Nested Schema for `security_rule.rule_condition.condition.http_method.http_methods`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--rule_condition--condition--request_uri"></a>
### Nested Schema for `security_rule.rule_condition.condition.request_uri`

Optional:

- `path` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--request_uri--path))
- `queries` (Block List) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--request_uri--queries))

<a id="nestedblock--security_rule--rule_condition--condition--request_uri--path"></a>
### Nested Schema for `security_rule.rule_condition.condition.request_uri.path`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)


<a id="nestedblock--security_rule--rule_condition--condition--request_uri--queries"></a>
### Nested Schema for `security_rule.rule_condition.condition.request_uri.queries`

Required:

- `key` (String)
- `value` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--request_uri--queries--value))

<a id="nestedblock--security_rule--rule_condition--condition--request_uri--queries--value"></a>
### Nested Schema for `security_rule.rule_condition.condition.request_uri.queries.value`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)




<a id="nestedblock--security_rule--rule_condition--condition--source_ip"></a>
### Nested Schema for `security_rule.rule_condition.condition.source_ip`

Optional:

- `geo_ip_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--source_ip--geo_ip_match))
- `geo_ip_not_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--source_ip--geo_ip_not_match))
- `ip_ranges_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--source_ip--ip_ranges_match))
- `ip_ranges_not_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--rule_condition--condition--source_ip--ip_ranges_not_match))

<a id="nestedblock--security_rule--rule_condition--condition--source_ip--geo_ip_match"></a>
### Nested Schema for `security_rule.rule_condition.condition.source_ip.geo_ip_match`

Optional:

- `locations` (List of String)


<a id="nestedblock--security_rule--rule_condition--condition--source_ip--geo_ip_not_match"></a>
### Nested Schema for `security_rule.rule_condition.condition.source_ip.geo_ip_not_match`

Optional:

- `locations` (List of String)


<a id="nestedblock--security_rule--rule_condition--condition--source_ip--ip_ranges_match"></a>
### Nested Schema for `security_rule.rule_condition.condition.source_ip.ip_ranges_match`

Optional:

- `ip_ranges` (List of String)


<a id="nestedblock--security_rule--rule_condition--condition--source_ip--ip_ranges_not_match"></a>
### Nested Schema for `security_rule.rule_condition.condition.source_ip.ip_ranges_not_match`

Optional:

- `ip_ranges` (List of String)





<a id="nestedblock--security_rule--smart_protection"></a>
### Nested Schema for `security_rule.smart_protection`

Optional:

- `condition` (Block List, Max: 1) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto). (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition))
- `mode` (String) Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).

<a id="nestedblock--security_rule--smart_protection--condition"></a>
### Nested Schema for `security_rule.smart_protection.condition`

Optional:

- `authority` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--authority))
- `headers` (Block List) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--headers))
- `http_method` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--http_method))
- `request_uri` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--request_uri))
- `source_ip` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--source_ip))

<a id="nestedblock--security_rule--smart_protection--condition--authority"></a>
### Nested Schema for `security_rule.smart_protection.condition.authority`

Optional:

- `authorities` (Block List) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--authority--authorities))

<a id="nestedblock--security_rule--smart_protection--condition--authority--authorities"></a>
### Nested Schema for `security_rule.smart_protection.condition.authority.authorities`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--smart_protection--condition--headers"></a>
### Nested Schema for `security_rule.smart_protection.condition.headers`

Required:

- `value` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--headers--value))

Optional:

- `name` (String)

<a id="nestedblock--security_rule--smart_protection--condition--headers--value"></a>
### Nested Schema for `security_rule.smart_protection.condition.headers.value`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--smart_protection--condition--http_method"></a>
### Nested Schema for `security_rule.smart_protection.condition.http_method`

Optional:

- `http_methods` (Block List) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--http_method--http_methods))

<a id="nestedblock--security_rule--smart_protection--condition--http_method--http_methods"></a>
### Nested Schema for `security_rule.smart_protection.condition.http_method.http_methods`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--smart_protection--condition--request_uri"></a>
### Nested Schema for `security_rule.smart_protection.condition.request_uri`

Optional:

- `path` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--request_uri--path))
- `queries` (Block List) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--request_uri--queries))

<a id="nestedblock--security_rule--smart_protection--condition--request_uri--path"></a>
### Nested Schema for `security_rule.smart_protection.condition.request_uri.path`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)


<a id="nestedblock--security_rule--smart_protection--condition--request_uri--queries"></a>
### Nested Schema for `security_rule.smart_protection.condition.request_uri.queries`

Required:

- `key` (String)
- `value` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--request_uri--queries--value))

<a id="nestedblock--security_rule--smart_protection--condition--request_uri--queries--value"></a>
### Nested Schema for `security_rule.smart_protection.condition.request_uri.queries.value`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)




<a id="nestedblock--security_rule--smart_protection--condition--source_ip"></a>
### Nested Schema for `security_rule.smart_protection.condition.source_ip`

Optional:

- `geo_ip_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--source_ip--geo_ip_match))
- `geo_ip_not_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--source_ip--geo_ip_not_match))
- `ip_ranges_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--source_ip--ip_ranges_match))
- `ip_ranges_not_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--smart_protection--condition--source_ip--ip_ranges_not_match))

<a id="nestedblock--security_rule--smart_protection--condition--source_ip--geo_ip_match"></a>
### Nested Schema for `security_rule.smart_protection.condition.source_ip.geo_ip_match`

Optional:

- `locations` (List of String)


<a id="nestedblock--security_rule--smart_protection--condition--source_ip--geo_ip_not_match"></a>
### Nested Schema for `security_rule.smart_protection.condition.source_ip.geo_ip_not_match`

Optional:

- `locations` (List of String)


<a id="nestedblock--security_rule--smart_protection--condition--source_ip--ip_ranges_match"></a>
### Nested Schema for `security_rule.smart_protection.condition.source_ip.ip_ranges_match`

Optional:

- `ip_ranges` (List of String)


<a id="nestedblock--security_rule--smart_protection--condition--source_ip--ip_ranges_not_match"></a>
### Nested Schema for `security_rule.smart_protection.condition.source_ip.ip_ranges_not_match`

Optional:

- `ip_ranges` (List of String)





<a id="nestedblock--security_rule--waf"></a>
### Nested Schema for `security_rule.waf`

Required:

- `waf_profile_id` (String) ID of WAF profile to use in this rule.

Optional:

- `condition` (Block List, Max: 1) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto). (see [below for nested schema](#nestedblock--security_rule--waf--condition))
- `mode` (String) Mode of protection. Possible values: `FULL` (full protection means that the traffic will be checked based on ML models and behavioral analysis, with suspicious requests being sent to SmartCaptcha) or `API` (API protection means checking the traffic based on ML models and behavioral analysis without sending suspicious requests to SmartCaptcha. The suspicious requests will be blocked).

<a id="nestedblock--security_rule--waf--condition"></a>
### Nested Schema for `security_rule.waf.condition`

Optional:

- `authority` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--authority))
- `headers` (Block List) (see [below for nested schema](#nestedblock--security_rule--waf--condition--headers))
- `http_method` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--http_method))
- `request_uri` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--request_uri))
- `source_ip` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--source_ip))

<a id="nestedblock--security_rule--waf--condition--authority"></a>
### Nested Schema for `security_rule.waf.condition.authority`

Optional:

- `authorities` (Block List) (see [below for nested schema](#nestedblock--security_rule--waf--condition--authority--authorities))

<a id="nestedblock--security_rule--waf--condition--authority--authorities"></a>
### Nested Schema for `security_rule.waf.condition.authority.authorities`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--waf--condition--headers"></a>
### Nested Schema for `security_rule.waf.condition.headers`

Required:

- `value` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--headers--value))

Optional:

- `name` (String)

<a id="nestedblock--security_rule--waf--condition--headers--value"></a>
### Nested Schema for `security_rule.waf.condition.headers.value`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--waf--condition--http_method"></a>
### Nested Schema for `security_rule.waf.condition.http_method`

Optional:

- `http_methods` (Block List) (see [below for nested schema](#nestedblock--security_rule--waf--condition--http_method--http_methods))

<a id="nestedblock--security_rule--waf--condition--http_method--http_methods"></a>
### Nested Schema for `security_rule.waf.condition.http_method.http_methods`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)



<a id="nestedblock--security_rule--waf--condition--request_uri"></a>
### Nested Schema for `security_rule.waf.condition.request_uri`

Optional:

- `path` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--request_uri--path))
- `queries` (Block List) (see [below for nested schema](#nestedblock--security_rule--waf--condition--request_uri--queries))

<a id="nestedblock--security_rule--waf--condition--request_uri--path"></a>
### Nested Schema for `security_rule.waf.condition.request_uri.path`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)


<a id="nestedblock--security_rule--waf--condition--request_uri--queries"></a>
### Nested Schema for `security_rule.waf.condition.request_uri.queries`

Required:

- `key` (String)
- `value` (Block List, Min: 1, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--request_uri--queries--value))

<a id="nestedblock--security_rule--waf--condition--request_uri--queries--value"></a>
### Nested Schema for `security_rule.waf.condition.request_uri.queries.value`

Optional:

- `exact_match` (String)
- `exact_not_match` (String)
- `pire_regex_match` (String)
- `pire_regex_not_match` (String)
- `prefix_match` (String)
- `prefix_not_match` (String)




<a id="nestedblock--security_rule--waf--condition--source_ip"></a>
### Nested Schema for `security_rule.waf.condition.source_ip`

Optional:

- `geo_ip_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--source_ip--geo_ip_match))
- `geo_ip_not_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--source_ip--geo_ip_not_match))
- `ip_ranges_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--source_ip--ip_ranges_match))
- `ip_ranges_not_match` (Block List, Max: 1) (see [below for nested schema](#nestedblock--security_rule--waf--condition--source_ip--ip_ranges_not_match))

<a id="nestedblock--security_rule--waf--condition--source_ip--geo_ip_match"></a>
### Nested Schema for `security_rule.waf.condition.source_ip.geo_ip_match`

Optional:

- `locations` (List of String)


<a id="nestedblock--security_rule--waf--condition--source_ip--geo_ip_not_match"></a>
### Nested Schema for `security_rule.waf.condition.source_ip.geo_ip_not_match`

Optional:

- `locations` (List of String)


<a id="nestedblock--security_rule--waf--condition--source_ip--ip_ranges_match"></a>
### Nested Schema for `security_rule.waf.condition.source_ip.ip_ranges_match`

Optional:

- `ip_ranges` (List of String)


<a id="nestedblock--security_rule--waf--condition--source_ip--ip_ranges_not_match"></a>
### Nested Schema for `security_rule.waf.condition.source_ip.ip_ranges_not_match`

Optional:

- `ip_ranges` (List of String)






<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_sws_security_profile.<resource Name> <resource Id>
terraform import yandex_sws_security_profile.demo-profile-simple ...
```
