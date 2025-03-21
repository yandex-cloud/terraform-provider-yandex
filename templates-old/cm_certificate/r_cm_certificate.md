---
subcategory: "Certificate Manager"
page_title: "Yandex: {{.Name}}"
description: |-
  A TLS certificate signed by a certification authority confirming that it belongs to the owner of the domain name.
---

# {{.Name}} ({{.Type}})

Creates or requests a TLS certificate in the specified folder. For more information, see [the official documentation](https://yandex.cloud/docs/certificate-manager/concepts/).

~> At the moment, a resource may not work correctly if it declares the use of a DNS challenge, but the certificate is confirmed using an HTTP challenge. And vice versa.

In this case, the service does not provide the parameters of the required type of challenges.


## Example usage

{{ tffile "examples/cm_certificate/r_cm_certificate_1.tf" }}

{{ tffile "examples/cm_certificate/r_cm_certificate_2.tf" }}

{{ tffile "examples/cm_certificate/r_cm_certificate_3.tf" }}

{{ tffile "examples/cm_certificate/r_cm_certificate_4.tf" }}

## Argument Reference

The following arguments are supported:

* `name` - (Optional) Certificate name.
* `folder_id` - (Optional) Folder that the resource belongs to. If value is omitted, the default provider folder is used.
* `description` - (Optional) Certificate description.
* `labels` - (Optional) Labels to assign to this certificate.
* `domains` - (Optional) Domains for this certificate. Should be specified for managed certificates.
* `managed` - (Optional) Managed specification. Structure is documented below.
* `self_managed` - (Optional) Self-managed specification. Structure is documented below.

~> Only one type `managed` or `self_managed` should be specified.

The `managed` block supports:

* `challenge_type` - (Required) Domain owner-check method. Possible values:
  - "DNS_CNAME" - you will need to create a CNAME dns record with the specified value. Recommended for fully automated certificate renewal;
  - "DNS_TXT" - you will need to create a TXT dns record with specified value;
  - "HTTP" - you will need to place specified value into specified url.
* `challenge_count` - (Optional). Expected number of challenge count needed to validate certificate. Resource creation will fail if the specified value does not match the actual number of challenges received from issue provider. This argument is helpful for safe automatic resource creation for passing challenges for multi-domain certificates.

~> Resource creation awaits getting challenges from issue provider.

The `self_managed` block supports:

* `certificate` - (Required) Certificate with chain.
* `private_key` - (Optional) Private key of certificate.
* `private_key_lockbox_secret` - (Optional) Lockbox secret specification for getting private key. Structure is documented below.

~> Only one type `private_key` or `private_key_lockbox_secret` should be specified.

The `private_key_lockbox_secret` block supports:

* `id` - (Required) Lockbox secret Id.
* `key` - (Required) Key of the Lockbox secret, the value of which contains the private key of the certificate.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - Certificate Id.
* `created_at` - Certificate create timestamp.
* `updated_at` - Certificate update timestamp.
* `type` - Certificate type: `"MANAGED"` or `"IMPORTED"`.
* `status` - Certificate status: `"VALIDATING"`, `"INVALID"`, `"ISSUED"`, `"REVOKED"`, `"RENEWING"` or `"RENEWAL_FAILED"`.
* `issuer` - Certificate issuer.
* `subject` - Certificate subject.
* `serial` - Certificate serial number.
* `issued_at` - Certificate issue timestamp.
* `not_before` - Certificate start valid period.
* `not_after` - Certificate end valid period.
* `challenges` - Array of challenges. Structure is documented below.

The `challenges` block represents (for each array element):

* `domain` - Validated domain.
* `type` - Challenge type `"DNS"` or `"HTTP"`.
* `created_at` - Time the challenge was created.
* `updated_at` - Last time the challenge was updated.
* `message` - Current status message.
* `dns_name` - DNS record name (only for DNS challenge).
* `dns_type` - DNS record type: `"TXT"` or `"CNAME"` (only for DNS challenge).
* `dns_value` - DNS record value (only for DNS challenge).
* `http_url` - URL where the challenge content http_content should be placed (only for HTTP challenge).
* `http_content` - The content that should be made accessible with the given `http_url` (only for HTTP challenge).

## Timeouts

This resource provides the following configuration options for timeouts:

- `read` - Default is 1 minute.
- `create` - Default is 1 minute.
- `update` - Default is 1 minute.
- `delete` - Default is 1 minute.

## Import

The resource can be imported by using their `resource ID`. For getting the resource ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

{{ codefile "bash" "examples/cm_certificate/import.sh" }}
