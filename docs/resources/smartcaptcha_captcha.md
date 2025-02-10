---
subcategory: "Smart Captcha"
page_title: "Yandex: yandex_smartcaptcha_captcha"
description: |-
  Allows management of a Yandex Cloud SmartCaptcha.
---

# yandex_smartcaptcha_captcha (Resource)

Creates a Captcha in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/smartcaptcha/).

## Example Usage

```terraform
// 
// Simple SmartCaptcha example.
//
resource "yandex_smartcaptcha_captcha" "demo-captcha-simple" {
  deletion_protection = true
  name                = "demo-captcha-simple"
  complexity          = "HARD"
  pre_check_type      = "SLIDER"
  challenge_type      = "IMAGE_TEXT"

  allowed_sites = [
    "example.com",
    "example.ru"
  ]
}
```

```terraform
//
// Advanced SmartCaptcha example.
//
resource "yandex_smartcaptcha_captcha" "demo-captcha-advanced" {
  deletion_protection = true
  name                = "demo-captcha-advanced"
  complexity          = "HARD"
  pre_check_type      = "SLIDER"
  challenge_type      = "IMAGE_TEXT"

  allowed_sites = [
    "example.com",
    "example.ru"
  ]

  override_variant {
    uuid        = "xxx"
    description = "override variant 1"

    complexity     = "EASY"
    pre_check_type = "CHECKBOX"
    challenge_type = "SILHOUETTES"
  }

  override_variant {
    uuid        = "yyy"
    description = "override variant 2"

    complexity     = "HARD"
    pre_check_type = "CHECKBOX"
    challenge_type = "KALEIDOSCOPE"
  }

  security_rule {
    name                  = "rule1"
    priority              = 11
    description           = "My first security rule. This rule it's just example to show possibilities of configuration."
    override_variant_uuid = "xxx"

    condition {
      host {
        hosts {
          exact_match = "example.com"
        }
        hosts {
          exact_match = "example.net"
        }
      }

      uri {
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

  security_rule {
    name                  = "rule2"
    priority              = 555
    description           = "Second rule"
    override_variant_uuid = "yyy"

    condition {
      uri {
        path {
          prefix_match = "/form"
        }
      }
    }
  }

  security_rule {
    name                  = "rule3"
    priority              = 99999
    description           = "Empty condition rule"
    override_variant_uuid = "yyy"
  }
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the captcha. The name must be unique within the folder.

* `complexity` - (Required) Complexity of the captcha. Possible values are documented below.

* `pre_check_type` - (Required) Basic check type of the captcha. Possible values are documented below.

* `challenge_type` - (Required) Additional task type of the captcha. Possible values are documented below.

* `folder_id` - (Optional) ID of the folder to create a captcha in. If omitted, the provider folder is used.

* `allowed_sites` - (Optional) List of allowed host names, see [Domain validation](https://yandex.cloud/docs/smartcaptcha/concepts/domain-validation).

* `style_json` - (Optional) JSON with variables to define the captcha appearance. For more details see generated JSON in cloud console.

* `turn_off_hostname_check` - (Optional) Turn off host name check, see [Domain validation](https://yandex.cloud/docs/smartcaptcha/concepts/domain-validation).

* `deletion_protection` - (Optional) Determines whether captcha is protected from being deleted.

* `security_rule` - (Optional) List of security rules. The structure is documented below.

* `override_variant` - (Optional) List of variants to use in security_rules. The structure is documented below.

---

Possible values of `complexity`:

* `EASY` - High chance to pass pre-check and easy advanced challenge.

* `MEDIUM` - Medium chance to pass pre-check and normal advanced challenge.

* `HARD` - Little chance to pass pre-check and hard advanced challenge.

* `FORCE_HARD` - Impossible to pass pre-check and hard advanced challenge.

---

Possible values of `pre_check_type`:

* `CHECKBOX` - User must click the "I am not a robot" button.

* `SLIDER` - User must move the slider from left to right.

---

Possible values of `challenge_type`:

* `IMAGE_TEXT` - Text recognition: The user has to type a distorted text from the picture into a special field.

* `SILHOUETTES` - Silhouettes: The user has to mark several icons from the picture in a particular order.

* `KALEIDOSCOPE` - Kaleidoscope: The user has to build a picture from individual parts by shuffling them using a slider.

---

The `override_variant` block supports:

* `uuid` - (Required) Unique identifier of the variant.

* `complexity` - (Required) Complexity of the captcha.

* `pre_check_type` - (Required) Basic check type of the captcha.

* `pre_check_type` - (Required) Additional task type of the captcha.

* `description` - (Optional) Optional description of the rule. 0-512 characters long.

---

The `security_rule` block supports:

* `name` - (Required) Name of the rule. The name is unique within the captcha. 1-50 characters long.

* `priority` - (Required) Priority of the rule. Lower value means higher priority.

* `description` - (Optional) Optional description of the rule. 0-512 characters long.

* `condition` - (Optional) The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartcaptcha/v1/captcha.proto).  

* `override_variant_uuid` - (Required) Variant UUID to show in case of match the rule. Keep empty to use defaults.

---

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - ID of the captcha.
* `created_at` - The Captcha creation timestamp.
* `client_key` - Client key of the captcha, see [CAPTCHA keys](https://yandex.cloud/docs/smartcaptcha/concepts/keys).

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

```shell
# terraform import yandex_smartcaptcha_captcha.<resource Name> <resource Id>
terraform import yandex_smartcaptcha_captcha.demo-captcha-simple ...
```
