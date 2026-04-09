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

resource "yandex_mdb_kafka_connector" "iceberg_static" {
  cluster_id = yandex_mdb_kafka_cluster.my_cluster.id
  name       = "iceberg-sink-static"
  tasks_max  = 2
  properties = {
    "key.converter"   = "org.apache.kafka.connect.storage.StringConverter"
    "value.converter" = "org.apache.kafka.connect.json.JsonConverter"
  }
  connector_config_iceberg_sink {
    topics        = "topic1,topic2,topic3"
    control_topic = "iceberg-control"

    metastore_connection {
      catalog_uri = "thrift://metastore.example.com:9083"
      warehouse   = "s3a://my-bucket/warehouse"
    }

    s3_connection {
      external_s3 {
        endpoint          = "https://storage.yandexcloud.net"
        access_key_id     = "some_access_key_id"
        secret_access_key = "some_secret_access_key"
        region            = "ru-central1"
      }
    }

    static_tables {
      tables = "db.table1,db.table2,db.table3"
    }

    tables_config {
      default_commit_branch    = "main"
      default_id_columns       = "id"
      default_partition_by     = "year(timestamp),month(timestamp)"
      evolve_schema_enabled    = true
      schema_force_optional    = false
      schema_case_insensitive  = true
    }

    control_config {
      group_id_prefix      = "cg-iceberg"
      commit_interval_ms   = 300000
      commit_timeout_ms    = 30000
      commit_threads       = 4
      transactional_prefix = "txn-"
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
