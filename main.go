package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

func NewMuxProviderServer(ctx context.Context) (func() tfprotov6.ProviderServer, error) {

	upgradedSdkProvider, _ := tf5to6server.UpgradeServer(
		context.Background(),
		yandex.NewSDKProvider().GRPCProvider,
	)

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(yandex_framework.NewFrameworkProvider()),
		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}

func main() {
	ctx := context.Background()
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	muxServerFactory, err := NewMuxProviderServer(ctx)

	if err != nil {
		return
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
		return
	}
}
