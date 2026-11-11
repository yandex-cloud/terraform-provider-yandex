package providermux

import (
	"context"
	"fmt"

	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New(ctx context.Context, pluginSDKProvider *schema.Provider, frameworkProvider frameworkprovider.Provider) (func() tfprotov6.ProviderServer, error) {
	upgradedProvider, err := tf5to6server.UpgradeServer(ctx, pluginSDKProvider.GRPCProvider)
	if err != nil {
		return nil, fmt.Errorf("upgrade Terraform Plugin SDK protocol: %w", err)
	}

	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(frameworkProvider),
		func() tfprotov6.ProviderServer {
			return upgradedProvider
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, fmt.Errorf("combine provider servers: %w", err)
	}

	return muxServer.ProviderServer, nil
}
