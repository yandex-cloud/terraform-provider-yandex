//
// Create a new MDB Clickhouse Cluster.
//
resource "yandex_mdb_clickhouse_cluster_v2" "my_cluster" {
    name        = "test"
    environment = "PRESTABLE"
    network_id  = yandex_vpc_network.foo.id

    clickhouse = {
        resources = {
            resource_preset_id = "s2.micro"
            disk_type_id       = "network-ssd"
            disk_size          = 32
        }

        config = {
            log_level                       = "TRACE"
            max_connections                 = 100
            max_concurrent_queries          = 50
            keep_alive_timeout              = 3000
            uncompressed_cache_size         = 8589934592
            max_table_size_to_drop          = 53687091200
            max_partition_size_to_drop      = 53687091200
            timezone                        = "UTC"
            geobase_uri                     = ""
            query_log_retention_size        = 1073741824
            query_log_retention_time        = 2592000
            query_thread_log_enabled        = true
            query_thread_log_retention_size = 536870912
            query_thread_log_retention_time = 2592000
            part_log_retention_size         = 536870912
            part_log_retention_time         = 2592000
            metric_log_enabled              = true
            metric_log_retention_size       = 536870912
            metric_log_retention_time       = 2592000
            trace_log_enabled               = true
            trace_log_retention_size        = 536870912
            trace_log_retention_time        = 2592000
            text_log_enabled                = true
            text_log_retention_size         = 536870912
            text_log_retention_time         = 2592000
            text_log_level                  = "TRACE"
            background_pool_size            = 16
            background_schedule_pool_size   = 16

            merge_tree = {
                replicated_deduplication_window                           = 100
                replicated_deduplication_window_seconds                   = 604800
                parts_to_delay_insert                                     = 150
                parts_to_throw_insert                                     = 300
                max_replicated_merges_in_queue                            = 16
                number_of_free_entries_in_pool_to_lower_max_size_of_merge = 8
                max_bytes_to_merge_at_min_space_in_pool                   = 1048576
                max_bytes_to_merge_at_max_space_in_pool                   = 161061273600
            }

            kafka = {
                security_protocol = "SECURITY_PROTOCOL_PLAINTEXT"
                sasl_mechanism    = "SASL_MECHANISM_GSSAPI"
                sasl_username     = "user1"
                sasl_password     = "pass1"
            }

            rabbitmq = {
                username = "rabbit_user"
                password = "rabbit_pass"
            }

            compression = [
                {
                    method              = "LZ4"
                    min_part_size       = 1024
                    min_part_size_ratio = 0.5
                },
                {
                    method              = "ZSTD"
                    min_part_size       = 2048
                    min_part_size_ratio = 0.7
                },
            ]

            graphite_rollup = [
                {
                    name = "rollup1"
                    patterns = [{
                        regexp   = "abc"
                        function = "func1"
                        retention = [{
                            age       = 1000
                            precision = 3
                        }]
                    }]
                },
                {
                    name = "rollup2"
                    patterns = [{
                        function = "func2"
                        retention = [{
                            age       = 2000
                            precision = 5
                        }]
                    }]
                }
            ]
        }
    }

    hosts = {
        "h1" = {
            type      = "CLICKHOUSE"
            zone      = "ru-central1-a"
            subnet_id = yandex_vpc_subnet.foo.id
        }
    }

    format_schema {
        name = "test_schema"
        type = "FORMAT_SCHEMA_TYPE_CAPNPROTO"
        uri  = "https://storage.yandexcloud.net/ch-data/schema.proto"
    }

    service_account_id = "your_service_account_id"

    cloud_storage = {
        enabled = false
    }

    maintenance_window {
        type = "ANYTIME"
    }
}

resource "yandex_vpc_network" "foo" {}

resource "yandex_vpc_subnet" "foo" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.foo.id
  v4_cidr_blocks = ["10.5.0.0/24"]
}
