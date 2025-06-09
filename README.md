Terraform Provider
==================

- Documentation: https://terraform-provider.yandexcloud.net

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.21 (to build the provider plugin)

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/yandex-cloud/terraform-provider-yandex`

```sh
$ mkdir -p $GOPATH/src/github.com/yandex-cloud; cd $GOPATH/src/github.com/yandex-cloud
$ git clone git@github.com:yandex-cloud/terraform-provider-yandex
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/yandex-cloud/terraform-provider-yandex
$ make build
```

Using the provider
----------------------
If you're building the provider, follow the instructions to [install it as a plugin.](https://www.terraform.io/docs/plugins/basics.html#installing-plugins) After placing it into your plugins directory,  run `terraform init` to initialize it. Documentation about the provider specific configuration options can be found on the [provider's website](https://registry.terraform.io/providers/yandex-cloud/yandex/latest/docs).
An example of using an installed provider from local directory: 

Write following config into  `~/.terraformrc`
```
provider_installation {
   dev_overrides {
    "yandex-cloud/yandex" = "/path/to/local/provider"
  }

   direct {}
 }
```

Developing the Provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-yandex
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of [Acceptance tests](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html), run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```

---

### Documentation Guide

Our documentation generator follows a specific flow. Based on the documentation template and the description fields for resources/datasources and their schema fields in the Terraform provider, we generate the corresponding documentation.

For every resource and data source defined in the Yandex provider, a documentation template is automatically generated.

If your resource/datasource does not have a template, please run the following command:

```shell
make generate-docs
```

If it is a new service, first update `templates/categories.yaml`.

#### Example of a Generated Template

For reference, you can view a generated template for [container_repository](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates/container_repository).

By default, these templates are generated automatically. However, if you want to enhance the documentation of your resource or data source, you can manually add or rewrite information in the template. To do this, you can use the available [template fields and functions](https://github.com/hashicorp/terraform-plugin-docs?tab=readme-ov-file#templates) that the Terraform documentation generator will process.

#### Adding Examples to Documentation

You can enhance your resource documentation by adding more examples. For instance, here’s an example from the [storage documentation template](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates):

```md
### Simple Private Bucket With Static Access Keys

{{tffile "examples/storage/resources/example_2.tf"}}

### Static Website Hosting

{{tffile "examples/storage/resources/example_3.tf"}}

### Using ACL Policy Grants

{{tffile "examples/storage/resources/example_4.tf"}}

### Using CORS

{{tffile "examples/storage/resources/example_5.tf"}}
```

#### Adding new service to terraform provider documentation

If you add a new service, please update `templates/categories.yaml`. The key in this file is the service path in the templates directory, and the value is the service name. This is used for grouping services on the documentation website. Then run the following command:

```shell
make generate-docs
```

On Mac M1 run the following command:

```shell
GOOS=linux GOARCH=amd64 make generate-docs
```

#### Resolving Template Variables

Here are some examples of how template variables resolve:

- [{{.Name}}](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates/dns/resources/dns_zone.md.tmpl#L8) resolves to [yandex_dns_zone](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/docs/resources/dns_zone.md#L8).
- [{{.Description}}](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates/dns/resources/dns_zone.md.tmpl#L10) resolves to [Manages a DNS Zone.](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/docs/resources/dns_zone.md#L10).
- [{{ .SchemaMarkdown }}](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates/dns/resources/dns_zone.md.tmpl#L11) resolves to the generated schema documentation, such as [Schema ...](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/docs/resources/dns_zone.md#L28).

#### General Recommendations

To ensure consistency and completeness in your documentation:

1. **Example Usage:**
   Uncomment the following block and add examples of resource/data source usage in the specified path:

   ```md
   {{- /* Uncomment this block as you add "examples/{serviceName}/resources/example_1.tf"

   ## Example Usage

   {{tffile "examples/{serviceName}/resources/example_1.tf" }}

   */ -}}
   ```

2. **Import Syntax:**
   Uncomment the following block and add examples of resource import in the specified path:

   ```md
   {{- /* Uncomment this block as you add "examples/{serviceName}/resources/import/import.sh"

   ## Import

   Import is supported using the following syntax:

   {{codefile "shell" "examples/{serviceName}/resources/import/import.sh" }}

   */ -}}
   ```

#### Documentation Build and Publishing

To build the documentation website locally, follow these steps:

1. First, install YFM:

    ```shell
    make install-yfm
    ```

2. Once installed, you can build your local version of the documentation by running the following command:

    ```shell
    make build-website
    ```

3. To publish updates to the documentation, set the following environment variables:

   ```shell
   export FM_STORAGE_KEY_ID=**censored**
   export YFM_STORAGE_SECRET_KEY=**censored**
   ```

   Then, run the following command:

   ```shell
   make publish-website
   ```

### Documentation Migration Guide

To migrate your document template to automatically generate resource/data source schema documentation, follow these steps:

1. **Add a Description:**
   Add a description to your resource in the Terraform provider. Example: [dns_zone](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/yandex/resource_yandex_dns_zone.go#L20).

2. **Describe Schema Fields:**
   Add a description to every schema field for the resource/data source. Example: [dns_zone.zone](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/yandex/resource_yandex_dns_zone.go#L45).

3. **Remove Manual Documentation:**
   - Remove the manual description of the resource/data source in the template. Example: [datasource](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates/datasphere/resources/datasphere_project.md.tmpl#L14).
   - Remove the `Argument Reference` block. Example: [Argument Reference](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates/datasphere/resources/datasphere_project.md.tmpl#L18-L27).
   - Remove the `Attributes Reference` block. Example: [Attributes Reference](https://github.com/yandex-cloud/terraform-provider-yandex/tree/master/templates/datasphere/resources/datasphere_project.md.tmpl#L29-L70).

4. **Automate Schema Documentation:**
   Place the `{{ .SchemaMarkdown }}` field in the template to automatically generate documentation based on field descriptions.

### Fixing Markdown Issues

Ensure that the markdown used in your documentation is properly formatted for consistent display and readability. This includes checking for proper syntax, resolving template variables, and ensuring that links and examples are correctly referenced.
