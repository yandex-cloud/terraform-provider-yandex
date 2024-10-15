---
subcategory: "Smart Web Security (SWS)"
page_title: "Yandex: yandex_sws_advanced_rate_limiter_profile"
description: |-
  Get information about SmartWebSecurity ARL Profile.
---


# yandex_sws_advanced_rate_limiter_profile




Get information about ARL Profile. For more information, see [the official documentation](https://yandex.cloud/en/docs/smartwebsecurity/quickstart/quickstart-arl).

## Example usage

```terraform
data "yandex_sws_advanced_rate_limiter_profile" "by-id" {
  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.my-profile.id
}
```

```terraform
data "yandex_sws_advanced_rate_limiter_profile" "by-name" {
  name = yandex_sws_advanced_rate_limiter_profile.my-profile.name
}
```

This data source is used to define ARL Profile that can be used by other resources.

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the ARL profile.
* `advanced_rate_limiter_profile_id` - (Optional) ID of the ARL profile.

~> **NOTE:** One of `advanced_rate_limiter_profile_id` or `name` should be specified.

## Attributes Reference

The following attributes are exported:

* `id` - ID of the profile.

* `created_at` - The profile creation timestamp.

* `name` - Name of the ARL profile. The name is unique within the folder. 1-50 characters long.

* `folder_id` - ID of the folder to create a profile in. If omitted, the provider folder is used.

* `labels` - Labels as key:value pairs. Maximum of 64 per resource.

* `description` - Optional description of the ARL profile.

* `advanced_rate_limiter_rule` - List of rules. The structure is documented below.

---

The `advanced_rate_limiter_rule` block supports:

* `name` - Name of the rule. The name is unique within the ARL profile. 1-50 characters long.

* `priority` - Determines the priority in case there are several matched rules. Enter an integer within the range of 1 and 999999. The rule priority must be unique within the entire ARL profile. A lower numeric value means a higher priority.

* `description` - Optional description of the rule. 0-512 characters long.

* `dry_run` - This allows you to evaluate backend capabilities and find the optimum limit values. Requests will not be blocked in this mode.

* `static_quota` - Static quota. Counting each request individually. The structure is documented below.

* `dynamic_quota` - Dynamic quota. Grouping requests by a certain attribute and limiting the number of groups. The structure is documented below.

~> **NOTE:** Exactly one rule specifier: `static_quota` or `dynamic_quota` should be specified.

---

The `static_quota` block supports:

* `action` - Action in case of exceeding this quota. Possible values: `DENY`.

* `condition` - The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

* `limit` - Desired maximum number of requests per period.

* `period` - Period of time in seconds.

---

The `dynamic_quota` block supports:

* `action` - Action in case of exceeding this quota. Possible values: `DENY`.

* `condition` - The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartwebsecurity/v1/security_profile.proto).

* `limit` - Desired maximum number of requests per period.

* `period` - Period of time in seconds.

* `characteristics` - List of characteristics. The structure is documented below.

---

The `characteristics` block supports:

* `simple_characteristic` - Characteristic automatically based on the Request path, HTTP method, IP address, Region, and Host attributes. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/arl#requests-counting) for more details. The structure is documented below.

* `key_characteristic` - Characteristic based on key match in the Query params, HTTP header, and HTTP cookie attributes. See [Rules](https://yandex.cloud/en/docs/smartwebsecurity/concepts/arl#requests-counting) for more details. The structure is documented below.

* `case_insensitive` - Determines case-sensitive or case-insensitive keys matching.

~> **NOTE:** Exactly one characteristic specifier: `simple_characteristic` or `key_characteristic` should be specified.

---

The `simple_characteristic` block supports:

* `type` - Type of simple characteristic. Possible values: `REQUEST_PATH`, `HTTP_METHOD`, `IP`, `GEO`, `HOST`.

---

The `key_characteristic` block supports:

* `type` - Type of key characteristic. Possible values: `COOKIE_KEY`, `HEADER_KEY`, `QUERY_KEY`.

* `value` - String value of the key.
