# ydb_table_index resource

## Example

```tf
resource "ydb_table_index" "index" {
    table_path        = "path/to/table"
    connection_string = "grpc://localhost:2136/?database=/local"
    name              = "my_index"
    type              = "global_sync"
    columns           = ["a", "b"]
    cover             = ["c"]
}
```