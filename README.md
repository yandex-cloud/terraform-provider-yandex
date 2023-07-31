Terraform Provider
==================

- Documentation: https://registry.terraform.io/providers/yandex-cloud/yandex/latest/docs

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.12+
- [Go](https://golang.org/doc/install) 1.20 (to build the provider plugin)

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
