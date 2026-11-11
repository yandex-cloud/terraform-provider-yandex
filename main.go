package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	frameworkprovider "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	"github.com/yandex-cloud/terraform-provider-yandex/internal/providermux"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
	grpcdns "google.golang.org/grpc/resolver/dns"
)

// Keep DNS lookups within the 20-second endpoint discovery deadline.
const grpcDNSResolvingTimeout = 10 * time.Second

func main() {
	grpcdns.SetResolvingTimeout(grpcDNSResolvingTimeout)

	ctx := context.Background()
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	provider, err := yandex.NewPluginSDKProvider()
	if err != nil {
		fmt.Fprintf(os.Stderr, "register Terraform Plugin SDK services: %v\n", err)
		os.Exit(1)
	}
	muxServerFactory, err := providermux.New(ctx, provider, frameworkprovider.NewFrameworkProvider())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var serveOpts []tf6server.ServeOpt

	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	err = tf6server.Serve(
		"yandex-cloud/yandex",
		muxServerFactory,
		serveOpts...,
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
