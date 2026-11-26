// Native YTsaurus external dictionary, available in public and internal clouds.
// Configure the provider with your folder and credentials before applying.
// The ClickHouse subnet must have network access to the YTsaurus HTTP proxies.
// The existing YTsaurus table must contain id (uint64) and name (string) columns.
// The OAuth token must grant read access to that table.

variable "network_id" {
  type = string
}

variable "subnet_id" {
  type = string
}

variable "zone" {
  type    = string
  default = "ru-central1-a"
}

variable "yt_http_proxy_urls" {
  description = "Non-empty list of YTsaurus HTTP proxy URLs reachable from ClickHouse."
  type        = list(string)
}

variable "yt_cypress_path" {
  description = "Cypress path to the existing YTsaurus table, for example //home/user/cities."
  type        = string
}

variable "yt_oauth_token" {
  type      = string
  sensitive = true
}

resource "yandex_mdb_clickhouse_cluster_v2" "yt_dictionary" {
  name        = "yt-dictionary-v2"
  environment = "PRESTABLE"
  network_id  = var.network_id

  clickhouse = {
    resources = {
      resource_preset_id = "s2.micro"
      disk_type_id       = "network-ssd"
      disk_size          = 32
    }
  }

  hosts = {
    clickhouse = {
      type       = "CLICKHOUSE"
      zone       = var.zone
      subnet_id  = var.subnet_id
      shard_name = "shard1"
    }
  }

  shards = {
    shard1 = {}
  }

  maintenance_window {
    type = "ANYTIME"
  }

  external_dictionary = {
    cities = {
      structure = {
        id = { name = "id" }
        attributes = [
          { name = "name", type = "String", null_value = "" }
        ]
      }
      layout   = { type = "FLAT" }
      lifetime = { fixed_lifetime = 300 }
      source = {
        yt_source_v2 = {
          http_proxy_urls = var.yt_http_proxy_urls
          cypress_path    = var.yt_cypress_path
          oauth_token     = var.yt_oauth_token
        }
      }
    }
  }
}
