//
// Create a new YDB Topic.
//
resource "yandex_ydb_topic" "my_topic" {
  database_endpoint = yandex_ydb_database_serverless.database_name.ydb_full_endpoint
  name              = "topic-test"

  supported_codecs    = ["raw", "gzip"]
  partitions_count    = 1
  retention_period_ms = 2000000
  consumer {
    name                          = "consumer-name"
    supported_codecs              = ["raw", "gzip"]
    starting_message_timestamp_ms = 0
  }
}

resource "yandex_ydb_database_serverless" "database_name" {
  name        = "database-name"
  location_id = "ru-central1"
}
