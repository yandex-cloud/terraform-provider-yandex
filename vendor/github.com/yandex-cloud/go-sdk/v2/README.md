# Yandex Cloud GO SDK

[![GoDoc](https://godoc.org/github.com/yandex-cloud/go-sdk?status.svg)](https://godoc.org/github.com/yandex-cloud/go-sdk)
[![CircleCI](https://circleci.com/gh/yandex-cloud/go-sdk.svg?style=shield)](https://circleci.com/gh/yandex-cloud/go-sdk)

Go SDK for Yandex Cloud services.

**NOTE:** SDK is under development, and may make backwards-incompatible changes.

## Table of Contents

- [Installation](#installation)
- [Example Usage](#example-usage)
  - [Initializing SDK](#initializing-sdk)
  - [SDK Authorization](#sdk-authorization)
  - [Service Usage](#service-usage)
  - [Retries](#retries)
- [More Examples](#more-examples)
## Installation

```bash
go get github.com/yandex-cloud/go-sdk/v2
```

## Example usages

### Initializing SDK

```go
sdk, err := ycsdk.Build(ctx,
    options..WithCredentials(credentials.IAMToken(os.Getenv("YC_IAM_TOKEN"))),
)
if err != nil {
    log.Fatal(err)
}
```

### SDK Authorization
https://yandex.cloud/ru/docs/iam/concepts/authorization/

####  IAM Token (Prefered)
https://yandex.cloud/ru/docs/iam/concepts/authorization/iam-token
```go
sdk, err := ycsdk.Build(ctx,
    options..WithCredentials(credentials.IAMToken(os.Getenv("YC_IAM_TOKEN"))),
)
if err != nil {
    log.Fatal(err)
}
```

####  Inside Yandex Cloud Virtual Machine (Prefered)
https://yandex.cloud/ru/docs/compute/operations/vm-connect/auth-inside-vm#auth-inside-vm
```go
sdk, err := ycsdk.Build(ctx,
    options.WithCredentials(credentials.InstanceServiceAccount()),
)
if err != nil {
	log.Fatal(err)
}
```

####  IAM Authorized Key (Service Account)
https://yandex.cloud/ru/docs/iam/concepts/authorization/key
```go
keyPath := "your-authorized-key-path"
creds, err := credentials.ServiceAccountKeyFile(keyPath)
if err != nil {
    log.Fatalf("failed to load credentials: %w", err)
}
sdk, err := ycsdk.Build(ctx,
    options.WithCredentials(creds),
)
if err != nil {
	log.Fatal(err)
}
```

### Service Usage

```go
import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yandex-cloud/go-sdk/v2"
    "github.com/yandex-cloud/go-sdk/v2/credentials"
    "github.com/yandex-cloud/go-sdk/v2/pkg/options"
    
    computeapi "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
    computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
)

func main() {
    // Create background context
    ctx := context.Background()

    // Initialize SDK with IAM token credentials from environment variable
    sdk, err := ycsdk.Build(ctx,
        options.WithCredentials(credentials.IAMToken(os.Getenv("YC_IAM_TOKEN")).
    ))
    
    if err != nil {
        log.Fatalf("failed to build SDK: %v", err)
    }

    // Create image client instance
    imageClient := computesdk.NewImageClient(sdk)

    // Request latest image from Debian 9
    resp, err := imageClient.GetLatestByFamily(ctx, &computeapi.GetImageLatestByFamilyRequest{
        FolderId: os.Getenv("YC_FOLDER_ID"), // Get folder ID from env variable
        Family:   "debian-9", // Specify Debian 9 image family
    })

    if err != nil {
        log.Fatalf("failed to get image: %v", err)
    }

    // Print response with image details
    fmt.Println(resp)
}

```


### Retries

SDK provide built-in retry policy, that supports [exponential backoff and jitter](https://aws.amazon.com/ru/blogs/architecture/exponential-backoff-and-jitter/), and also [retry budget](https://github.com/grpc/proposal/blob/master/A6-client-retries.md#throttling-retry-attempts-and-hedged-rpcs). 
It's necessary to avoid retry amplification.

```go
import (
    ...
    ycsdk "github.com/yandex-cloud/go-sdk/v2"
    "github.com/yandex-cloud/go-sdk/v2/credentials"
    "github.com/yandex-cloud/go-sdk/v2/pkg/options"
)

...

sdk, err := ycsdk.Build(
    ctx,
    options.WithCredentials(credentials.IAMToken(os.Getenv("YC_IAM_TOKEN"))),
    options.WithDefaultRetryOptions(),
)
```

SDK provide different modes for retry throttling policy:

* `persistent` is suitable when you use SDK in any long-lived application, when SDK instance will live long enough for manage budget;
* `temporary` is suitable when you use SDK in any short-lived application, e.g. scripts or CI/CD.

By default, SDK will use temporary mode, but you can change it through functional option.

### SDK Resolvers

SDK resolvers allows you to find resource id by its name.

```go

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    
    ycsdk "bb.yandex-team.ru/cloud/cloud-go/sdk-v2"
    "bb.yandex-team.ru/cloud/cloud-go/sdk-v2/credentials"
    "bb.yandex-team.ru/cloud/cloud-go/sdk-v2/pkg/options"
    "bb.yandex-team.ru/cloud/cloud-go/sdk-v2/pkg/sdkresolvers"
    computesdk "bb.yandex-team.ru/cloud/cloud-go/sdk-v2/sdkresolvers/compute/v1"
)

func main() {
	iamToken := flag.String("token", os.Getenv("YC_IAM_TOKEN"), "IAM token for Yandex.Cloud (env YC_IAM_TOKEN)")
	folderID := flag.String("folder-id", os.Getenv("YC_FOLDER_ID"), "Yandex.Cloud Folder ID (env YC_FOLDER_ID)")

	ctx := context.Background()
	sdk, err := ycsdk.Build(ctx,
		options.WithCredentials(credentials.IAMToken(*iamToken)),
	)
	if err != nil {
		log.Fatal(err)
	}

	name := "test_name"
	r := computesdk.DiskResolver(name, sdkresolvers.FolderID(*folderID))
	if err = r.Run(ctx, sdk); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("id of object %s is %s", name, r.ID())
}
```

## More examples

More examples can be found in [examples dir](examples).
