# `topic` - pure Go native client for [YDB Topic](https://ydb.tech/en/docs/concepts/topic)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/ydb-platform/ydb/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/ydb-platform/ydb-go-sdk.svg?style=flat-square)](https://github.com/ydb-platform/ydb-go-sdk/releases)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ydb-platform/ydb-go-sdk/v3)](https://pkg.go.dev/github.com/ydb-platform/ydb-go-sdk/v3/topic)
![tests](https://github.com/ydb-platform/ydb-go-sdk/workflows/tests/badge.svg?branch=master)
![lint](https://github.com/ydb-platform/ydb-go-sdk/workflows/lint/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/ydb-platform/ydb-go-sdk/v3)](https://goreportcard.com/report/github.com/ydb-platform/ydb-go-sdk/v3)
[![codecov](https://codecov.io/gh/ydb-platform/ydb-go-sdk/branch/master/graph/badge.svg?precision=2)](https://app.codecov.io/gh/ydb-platform/ydb-go-sdk)
![Code lines](https://sloc.xyz/github/ydb-platform/ydb-go-sdk/?category=code)
[![View examples](https://img.shields.io/badge/learn-examples-brightgreen.svg)](https://github.com/ydb-platform/ydb-go-sdk/tree/master/examples/topic)
[![Telegram](https://img.shields.io/badge/chat-on%20Telegram-2ba2d9.svg)](https://t.me/YDBPlatform)
[![WebSite](https://img.shields.io/badge/website-ydb.tech-blue.svg)](https://ydb.tech)

See [ydb-go-sdk](https://github.com/ydb-platform/ydb-go-sdk) for describe all driver features.

`YDB` is an open-source Distributed SQL Database that combines high availability and scalability with strict consistency and [ACID](https://en.wikipedia.org/wiki/ACID) transactions with [CDC](https://en.wikipedia.org/wiki/Change_data_capture) and Data stream features.

## Installation

```sh
go get -u github.com/ydb-platform/ydb-go-sdk/v3
```

## Example Usage <a name="example"></a>
* connect to YDB
```golang
db, err := ydb.Open(ctx, "grpc://localhost:2136/local")
if err != nil {
    log.Fatal(err)
}
```

* send messages
```golang
producerAndGroupID := "group-id"
writer, err := db.Topic().StartWriter(producerAndGroupID, "topicName",
    topicoptions.WithMessageGroupID(producerAndGroupID),
    topicoptions.WithCodec(topictypes.CodecGzip),
)
if err != nil {
    log.Fatal(err)
}

data1 := []byte{1, 2, 3}
data2 := []byte{4, 5, 6}
mess1 := topicwriter.Message{Data: bytes.NewReader(data1)}
mess2 := topicwriter.Message{Data: bytes.NewReader(data2)}

err = writer.Write(ctx, mess1, mess2)
if err != nil {
	log.Fatal(err)
}
```

* read messages
```golang
for {
    msg, err := reader.ReadMessage(ctx)
    if err != nil {
        log.Fatal(err)
    }
    content, err := io.ReadAll(msg)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(content))
    err = reader.Commit(msg.Context(), msg)
    if err != nil {
        log.Fatal(err)
    }
}
 
```
