---
page_title: "Provider: Yandex Cloud"
description: |-
  The Yandex Cloud provider is used to interact with Yandex Cloud services.
  The provider needs to be configured with the proper credentials before it can be used.
---

# Yandex Cloud Provider

The Yandex Cloud provider is used to interact with [Yandex Cloud services](https://yandex.cloud). The provider needs to be configured with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

{{ tffile "examples/provider/provider_1.tf" }}

{{ .SchemaMarkdown }}

## Shared credentials file

Shared credentials file must contain key/value credential pairs for different profiles in a specific format.

* Profile is specified in square brackets on a separate line (`[{profile_name}]`).
* Secret variables are specified in the `{key}={value}` format, one secret per line.

Every secret belongs to the closest profile above in the file.

You can find a configuration example below.

### Example of shared credentials usage:

{{ codefile "text" "examples/provider/config.txt" }}

{{ tffile "examples/provider/provider_2.tf" }}
