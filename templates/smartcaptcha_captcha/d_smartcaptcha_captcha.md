---
subcategory: "Smart Captcha"
page_title: "Yandex: {{.Name}}"
description: |-
  Get information about Yandex SmartCaptcha.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/smartcaptcha_captcha/d_smartcaptcha_captcha_1.tf" }}

{{ tffile "examples/smartcaptcha_captcha/d_smartcaptcha_captcha_2.tf" }}

{{ .SchemaMarkdown | trimspace }}
