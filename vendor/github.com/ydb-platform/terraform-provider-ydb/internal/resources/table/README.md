# ydb_table resource

`ydb_table` resource is used to manage YDB table entity.

## Example

```tf
resource "ydb_table" "table" {
    table_path        = "path/to/table"
    connection_string = "grpc://localhost:2136/?database=/local"
    column {
        name = "a"
        type = "Utf8"
    }
    column {
        name = "b"
        type = "Uint32"
    }

    primary_key = ["b", "a"]
}
```