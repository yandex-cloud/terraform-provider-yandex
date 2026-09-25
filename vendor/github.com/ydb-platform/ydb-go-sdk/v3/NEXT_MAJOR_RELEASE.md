# Breaking changes for the next major release
- [ ] Make connection string as required param to ydb.New call (instead optional params ydb.WithConnectionString, ydb.WithEndpoint, ydb.WithDatabase, ydb.WithInsecure)
- [ ] Change signature of `closer.Closer.Close(ctx) error` to `closer.Closer.Close()`
- [ ] Remove `retry.FastBackoff` and `retry.SlowBackoff` public variables
