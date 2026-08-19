---
subcategory: "Smart Web Security (SWS)"
---

# yandex_sws_security_profile_waf_profile_attachment (Resource)

Tracks a WAF rule attachment in a Smart Web Security profile and removes that
rule before the WAF profile is deleted.

The WAF rule remains configured in `yandex_sws_security_profile`. This resource
adds the lifecycle edge needed to detach the rule before Terraform destroys the
referenced WAF profile. It preserves all other Security Profile rules.

## Example usage

```terraform
resource "yandex_sws_security_profile_waf_profile_attachment" "attachment" {
  security_profile_id = yandex_sws_security_profile.security_profile.id
  waf_profile_id      = yandex_sws_waf_profile.waf.id
  security_rule_name  = "waf"
}
```

## Arguments & Attributes Reference

- `id` (String). Attachment ID in `<security_profile_id>/<waf_profile_id>/<security_rule_name>` format.
- `security_profile_id` (**Required**) (String). ID of the Smart Web Security profile.
- `security_rule_name` (**Required**) (String). Name of the WAF rule configured in the Smart Web Security profile.
- `waf_profile_id` (**Required**) (String). ID of the WAF profile.
- `timeouts` [Block].
  - `create` (String).
  - `delete` (String).
  - `read` (String).

## Import

```shell
terraform import yandex_sws_security_profile_waf_profile_attachment.attachment '<security_profile_id>/<waf_profile_id>/<security_rule_name>'
```
