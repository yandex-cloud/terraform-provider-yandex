---
subcategory: "Smart Captcha"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud SmartCaptcha.
---

# {{.Name}} ({{.Type}})

Creates a Captcha in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/smartcaptcha/).

## Example Usage

{{ tffile "examples/smartcaptcha_captcha/r_smartcaptcha_captcha_1.tf" }}

{{ tffile "examples/smartcaptcha_captcha/r_smartcaptcha_captcha_2.tf" }}


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

{{ codefile "shell" "examples/smartcaptcha_captcha/import.sh" }}
