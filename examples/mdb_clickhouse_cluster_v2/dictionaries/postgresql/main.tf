// Native PostgreSQL external dictionary.
// Configure the provider with your folder and credentials before applying.
// The subnet must have network access to the existing PostgreSQL hosts.
// The PostgreSQL table public.cities must contain id (non-negative bigint)
// and name (text) columns. The source user needs SELECT access to this table.

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

variable "postgresql_hosts" {
  description = "PostgreSQL hosts reachable from the ClickHouse subnet."
  type        = list(string)
}

variable "postgresql_database" {
  type = string
}

variable "postgresql_user" {
  type = string
}

variable "postgresql_password" {
  type      = string
  sensitive = true
}

resource "yandex_mdb_clickhouse_cluster_v2" "postgresql_dictionary" {
  name        = "postgresql-dictionary-v2"
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
        postgresql_source_v2 = {
          db       = var.postgresql_database
          table    = "cities"
          port     = 5432
          user     = var.postgresql_user
          password = var.postgresql_password
          ssl_mode = "REQUIRE"

          // Replicas inherit port, user and password from the source.
          replicas = [for index, host in var.postgresql_hosts : {
            host     = host
            priority = index + 1
          }]

          // For custom SQL, replace table with:
          // query = "SELECT id, name FROM public.cities"
          // Specify exactly one of table or query; do not combine query with where.
          background_reconnect = true
        }
      }
    }
  }
}
