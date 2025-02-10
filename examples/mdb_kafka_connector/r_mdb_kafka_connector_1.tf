//
// Create a new MDB Kafka Connector.
//
resource "yandex_mdb_kafka_connector" "my_conn" {
  cluster_id = yandex_mdb_kafka_cluster.my_cluster.id
  name       = "replication"
  tasks_max  = 3
  properties = {
    refresh.topics.enabled = "true"
  }
  connector_config_mirrormaker {
    topics             = "data.*"
    replication_factor = 1
    source_cluster {
      alias = "source"
      external_cluster {
        bootstrap_servers = "somebroker1:9091,somebroker2:9091"
        sasl_username     = "someuser"
        sasl_password     = "somepassword"
        sasl_mechanism    = "SCRAM-SHA-512"
        security_protocol = "SASL_SSL"
      }
    }
    target_cluster {
      alias = "target"
      this_cluster {}
    }
  }
}

resource "yandex_mdb_kafka_connector" "connector" {
  cluster_id = yandex_mdb_kafka_cluster.my_cluster.id
  name       = "s3-sink"
  tasks_max  = 3
  properties = {
    "key.converter"                  = "org.apache.kafka.connect.storage.StringConverter"
    "value.converter"                = "org.apache.kafka.connect.json.JsonConverter"
    "value.converter.schemas.enable" = "false"
    "format.output.type"             = "jsonl"
    "file.name.template"             = "dir1/dir2/{{topic}}-{{partition:padding=true}}-{{start_offset:padding=true}}.gz"
    "timestamp.timezone"             = "Europe/Moscow"
  }
  connector_config_s3_sink {
    topics                = "data.*"
    file_compression_type = "gzip"
    file_max_records      = 100
    s3_connection {
      bucket_name = "somebucket"
      external_s3 {
        endpoint          = "storage.yandexcloud.net"
        access_key_id     = "some_access_key_id"
        secret_access_key = "some_secret_access_key"
      }
    }
  }
}

resource "yandex_mdb_kafka_cluster" "my_cluster" {
  name       = "foo"
  network_id = "c64vs98keiqc7f24pvkd"

  config {
    version = "2.8"
    zones   = ["ru-central1-a"]
    kafka {
      resources {
        resource_preset_id = "s2.micro"
        disk_type_id       = "network-hdd"
        disk_size          = 16
      }
    }
  }
}
