// Package ycsdk is the official Yandex Cloud SDK v2 for the Go programming language.
//
// go-sdk-v2 is the the v2 of the Yandex Cloud SDK for the Go programming language.
//
// # Getting started
//
// The best way to get started working with the SDK is to use `go get` to add the
// SDK and desired service clients to your Go dependencies explicitly.
//
//	go get github.com/yandex-cloud/go-sdk-v2
//	go get github.com/yandex-cloud/go-sdk-v2/services/compute
//
// # Hello Yandex Cloud
//
// This example shows how you can use the v2 SDK to make an API request using the
// SDK's Yandex Compute client.
//
//      package main
//
//      import (
//          "context"
//          "fmt"
//          "log"
//          "os"
//
//          computeapi "github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
//          "github.com/yandex-cloud/go-sdk-v2"
//          "github.com/yandex-cloud/go-sdk-v2/credentials"
//          computesdk "github.com/yandex-cloud/go-sdk-v2/services/compute/v1"
//      )
//
//      func main() {
//          ctx := context.Background()
//          sdk, err := ycsdk.Build(ctx,
//              ycsdk.WithCredentials(credentials.IAMToken(os.Getenv("YC_IAM_TOKEN"))))
//          if err != nil {
//              log.Fatalf("failed to build SDK: %v", err)
//          }
//
//          imageClient := computesdk.NewImageClient(sdk)
//          resp, err := imageClient.GetLatestByFamily(ctx, &computeapi.GetImageLatestByFamilyRequest{
//              FolderId: os.Getenv("YC_FOLDER_ID"),
//              Family:   "debian-9",
//          })
//          if err != nil {
//              log.Fatalf("failed to get image: %v", err)
//          }
//
//          fmt.Println(resp)
//
//      }

package ycsdk
