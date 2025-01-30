---
subcategory: "Smart Captcha"
page_title: "Yandex: yandex_smartcaptcha_captcha"
description: |-
  Get information about Yandex SmartCaptcha.
---

# yandex_smartcaptcha_captcha (Data Source)

Get information about Yandex SmartCaptcha. For more information, see [the official documentation](https://yandex.cloud/docs/smartcaptcha/).

This data source is used to define Captcha that can be used by other resources.

## Example Usage

```terraform
//
// Get SmartCaptcha details by Id
//
data "yandex_smartcaptcha_captcha" "by-id" {
  captcha_id = yandex_smartcaptcha_captcha.my-captcha.id
}
```

```terraform
//
// Get SmartCaptcha details by Name
//
data "yandex_smartcaptcha_captcha" "by-name" {
  name = yandex_smartcaptcha_captcha.my-captcha.name
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Optional) Name of the Captcha.
* `captcha_id` - (Optional) ID of the Captcha.

~> One of `captcha_id` or `name` should be specified.

## Attributes Reference

The following attributes are exported:

* `id` - ID of the captcha.

* `created_at` - The Captcha creation timestamp.

* `client_key` - Client key of the captcha, see [CAPTCHA keys](https://yandex.cloud/docs/smartcaptcha/concepts/keys).

* `name` - Name of the captcha. The name must be unique within the folder.

* `complexity` - Complexity of the captcha. Possible values are documented below.

* `pre_check_type` - Basic check type of the captcha. Possible values are documented below.

* `challenge_type` - Additional task type of the captcha. Possible values are documented below.

* `allowed_sites` - List of allowed host names, see [Domain validation](https://yandex.cloud/docs/smartcaptcha/concepts/domain-validation).

* `style_json` - JSON with variables to define the captcha appearance. For more details see generated JSON in cloud console.

* `turn_off_hostname_check` - Turn off host name check, see [Domain validation](https://yandex.cloud/docs/smartcaptcha/concepts/domain-validation).

* `deletion_protection` - Determines whether captcha is protected from being deleted.

* `security_rule` - List of security rules. The structure is documented below.

* `override_variant` - List of variants to use in security_rules. The structure is documented below.

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

* `uuid` - Unique identifier of the variant.

* `complexity` - Complexity of the captcha.

* `pre_check_type` - Basic check type of the captcha.

* `pre_check_type` - Additional task type of the captcha.

* `description` - Optional description of the rule. 0-512 characters long.

---

The `security_rule` block supports:

* `name` - Name of the rule. The name is unique within the captcha. 1-50 characters long.

* `priority` - Priority of the rule. Lower value means higher priority.

* `description` - Optional description of the rule. 0-512 characters long.

* `condition` - The condition for matching the rule. You can find all possibilities of condition in [gRPC specs](https://github.com/yandex-cloud/cloudapi/blob/master/yandex/cloud/smartcaptcha/v1/captcha.proto).  

* `override_variant_uuid` - Variant UUID to show in case of match the rule. Keep empty to use defaults.
