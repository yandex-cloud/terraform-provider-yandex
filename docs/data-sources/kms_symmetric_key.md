---
subcategory: "Key Management Service (KMS)"
page_title: "Yandex: yandex_kms_symmetric_key"
description: |-
  Get data from Yandex KMS symmetric key.
---

# yandex_kms_symmetric_key (Data Source)



## Example Usage

```terraform
//
// TBD
//
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `default_algorithm` (String)
- `deletion_protection` (Boolean)
- `description` (String)
- `folder_id` (String)
- `labels` (Map of String)
- `name` (String)
- `rotation_period` (String)
- `symmetric_key_id` (String)

### Read-Only

- `created_at` (String)
- `id` (String) The ID of this resource.
- `rotated_at` (String)
- `status` (String)
