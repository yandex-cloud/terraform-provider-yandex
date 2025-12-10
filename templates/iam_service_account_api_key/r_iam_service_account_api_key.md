---
subcategory: "Identity and Access Management (IAM)"
page_title: "Yandex: {{.Name}}"
description: |-
  Allows management of a Yandex Cloud IAM service account API key.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example usage

{{ tffile "examples/iam_service_account_api_key/r_iam_service_account_api_key_1.tf" }}

{% note info %}

Field `scope` has been deprecated for `scopes` to allow multiple scope values. It's left in code for backward compatibility, but will be removed in the next major release.

If you face false changes of this field during apply, use this directive (as in example above):

```
resource ... {
  lifecycle {
    ignore_changes = [scope]
  }
  ...
}
```

{% endnote %}

{{ .SchemaMarkdown | trimspace }}

## Import

~> Import for this resource is not implemented yet.