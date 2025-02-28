---
subcategory: "Managed Service for Redis"
page_title: "Yandex: {{.Name}}"
description: |-
  Manages a Redis cluster within Yandex Cloud.
---

# {{.Name}} ({{.Type}})

{{ .Description | trimspace }}

## Example Usage

{{ tffile "examples/mdb_redis_cluster_v2/r_mdb_redis_cluster_v2_1.tf" }}

Example of creating a high available Redis Cluster.

{{ tffile "examples/mdb_redis_cluster_v2/r_mdb_redis_cluster_v2_2.tf" }}

{{ .SchemaMarkdown | trimspace }}

## Import

The resource can be imported by using their `resource ID`. For getting the cluster ID you can use Yandex Cloud [Web Console](https://console.yandex.cloud) or [YC CLI](https://yandex.cloud/docs/cli/quickstart).

After using import, you need to run terraform apply to pull up the host tags from the config to the state

{{ codefile "bash" "examples/mdb_redis_cluster_v2/import.sh" }}
