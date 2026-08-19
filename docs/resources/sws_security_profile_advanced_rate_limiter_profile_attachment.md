---
subcategory: "Smart Web Security (SWS)"
---

# yandex_sws_security_profile_advanced_rate_limiter_profile_attachment (Resource)

Attaches an Advanced Rate Limiter profile to a Smart Web Security profile.

Use this resource when the ARL profile and its attachment are controlled by the
same conditional expression. Terraform then destroys the attachment before it
destroys either profile.

~> Do not set `advanced_rate_limiter_profile_id` in
`yandex_sws_security_profile` when this resource manages the same attachment.

## Migrating from an inline attachment

An existing inline attachment can be migrated without recreating either
profile. In the same configuration change, remove
`advanced_rate_limiter_profile_id` from `yandex_sws_security_profile` and add
this attachment resource with the same Security Profile and ARL Profile IDs.
After apply, the next plan is empty.

## Example usage

```terraform
resource "yandex_sws_advanced_rate_limiter_profile" "arl" {
  name = "example-arl-profile"

  advanced_rate_limiter_rule {
    name     = "example-rule"
    priority = 1
    dry_run  = true

    static_quota {
      action = "DENY"
      limit  = 1000
      period = 60
    }
  }
}

resource "yandex_sws_security_profile" "security_profile" {
  name                     = "example-security-profile"
  default_action           = "ALLOW"
  disallow_data_processing = false
}

resource "yandex_sws_security_profile_advanced_rate_limiter_profile_attachment" "attachment" {
  security_profile_id              = yandex_sws_security_profile.security_profile.id
  advanced_rate_limiter_profile_id = yandex_sws_advanced_rate_limiter_profile.arl.id
}
```

## Arguments & Attributes Reference

- `advanced_rate_limiter_profile_id` (**Required**) (String). ID of the Advanced Rate Limiter profile.
- `id` (String). Attachment ID in `<security_profile_id>/<advanced_rate_limiter_profile_id>` format.
- `security_profile_id` (**Required**) (String). ID of the Smart Web Security profile.
- `timeouts` [Block].
  - `create` (String).
  - `delete` (String).
  - `read` (String).

## Import

The attachment can be imported using both profile IDs separated by `/`:

```shell
terraform import yandex_sws_security_profile_advanced_rate_limiter_profile_attachment.attachment '<security_profile_id>/<advanced_rate_limiter_profile_id>'
```
