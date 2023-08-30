package main

import (
	"context"
	"flag"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework"
)

func NewMuxProividerServer(ctx context.Context) (func() tfprotov5.ProviderServer, error) {
	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(yandex_framework.NewFrameworkProvider()),
		yandex.NewSDKProvider().GRPCProvider,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
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

	muxServerFactory, err := NewMuxProividerServer(ctx)

	if err != nil {
		return
	}

	serveOpts := []tf5server.ServeOpt{}

	if debug {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	err = tf5server.Serve(
		"yandex-cloud/yandex",
		muxServerFactory,
		serveOpts...,
	)

	if err != nil {
		return
	}
}
