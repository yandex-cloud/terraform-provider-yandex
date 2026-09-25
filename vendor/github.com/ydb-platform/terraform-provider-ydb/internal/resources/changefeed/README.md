# ydb_table_changefeed resource

## Example

```tf
resource "ydb_table_changefeed" "changefeed" {
    table_path        = "path/to/table"
    connection_string = "grpc://localhost:2136/?database=/local"

    name     = "changefeed"
    mode     = "NEW_IMAGE"
    format   = "JSON"
}
```