---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: {{.Name}}"
description: |-
  Manage a SWS Advanced Rate Limiter.
---

# {{.Name}} ({{.Type}})

Creates an ARL Profile in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/smartwebsecurity/quickstart#arl).

## Example usage

{{ tffile "examples/sws_advanced_rate_limiter_profile/r_sws_advanced_rate_limiter_profile_1.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the ARL profile. The name is unique within the folder. 1-50 characters long.

* `folder_id` - (Optional) ID of the folder to create a profile in. If omitted, the provider folder is used.

* `labels` - (Optional) Labels as key:value pairs. Maximum of 64 per resource.

* `description` - (Optional) Optional description of the ARL profile.

* `advanced_rate_limiter_rule` - (Required) List of rules. The structure is documented below.

---

The `advanced_rate_limiter_rule` block supports:

* `name` - (Required) Name of the rule. The name is unique within the ARL profile. 1-50 characters long.

* `priority` - (Required) Determines the priority in case there are several matched rules. Enter an integer within the range of 1 and 999999. The rule priority must be unique within the entire ARL profile. A lower numeric value means a higher priority.

* `description` - (Optional) Optional description of the rule. 0-512 characters long.

* `dry_run` - (Optional) This allows you to evaluate backend capabilities and find the optimum limit values. Requests will not be blocked in this mode.

* `static_quota` - (Optional) Static quota. Counting each request individually. The structure is documented below.

* `dynamic_quota` - (Optional) Dynamic quota. Grouping requests by a certain attribute and limiting the number of groups. The structure is documented below.

~> Exactly one rule specifier: `static_quota` or `dynamic_quota` should be specified.

---

The `static_quota` block supports:

* `action` - (Required) Action in case of exceeding this quota. Possible values: `DENY`.

* `condition` - (Optional) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

* `limit` - (Required) Desired maximum number of requests per period.

* `period` - (Required) Period of time in seconds.

---

The `dynamic_quota` block supports:

* `action` - (Required) Action in case of exceeding this quota. Possible values: `DENY`.

* `condition` - (Optional) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

* `limit` - (Required) Desired maximum number of requests per period.

* `period` - (Required) Period of time in seconds.

* `characteristics` - (Required) List of characteristics. The structure is documented below.

---

The `characteristics` block supports:

* `simple_characteristic` - (Optional) Characteristic automatically based on the Request path, HTTP method, IP address, Region, and Host attributes. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/arl#requests-counting) for more details. The structure is documented below.

* `key_characteristic` - (Optional) Characteristic based on key match in the Query params, HTTP header, and HTTP cookie attributes. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/arl#requests-counting) for more details. The structure is documented below.

* `case_insensitive` - (Optional) Determines case-sensitive or case-insensitive keys matching.

~> Exactly one characteristic specifier: `simple_characteristic` or `key_characteristic` should be specified.

---

The `simple_characteristic` block supports:

* `type` - (Required) Type of simple characteristic. Possible values: `REQUEST_PATH`, `HTTP_METHOD`, `IP`, `GEO`, `HOST`.

---

The `key_characteristic` block supports:

* `type` - (Required) Type of key characteristic. Possible values: `COOKIE_KEY`, `HEADER_KEY`, `QUERY_KEY`.

* `value` - (Required) String value of the key.

---

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - ID of the ARL profile.

* `created_at` - The ARL Profile creation timestamp.


## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{codefile "shell" "examples/sws_advanced_rate_limiter_profile/import.sh" }}
