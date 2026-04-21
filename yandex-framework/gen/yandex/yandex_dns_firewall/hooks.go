package yandex_dns_firewall

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/dns/v1"
	dnsv1sdk "github.com/yandex-cloud/go-sdk/services/dns/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func beforeUpdateHook(ctx context.Context, config *config.Config, _ *dns.UpdateDnsFirewallRequest, plan *yandexDnsFirewallModel, state *yandexDnsFirewallModel) diag.Diagnostics {
	dg := diag.Diagnostics{}
	return moveDnsFirewallIfNeeded(dg, ctx, config, plan, state)
}

func moveDnsFirewallIfNeeded(dg diag.Diagnostics, ctx context.Context, config *config.Config, plan *yandexDnsFirewallModel, state *yandexDnsFirewallModel) diag.Diagnostics {
	if plan.FolderId.ValueString() != state.FolderId.ValueString() {
		req := &dns.MoveDnsFirewallRequest{
			DnsFirewallId:       plan.DnsFirewallId.ValueString(),
			DestinationFolderId: plan.FolderId.ValueString(),
		}

		md := new(metadata.MD)
		op, err := dnsv1sdk.NewDnsFirewallClient(config.SDKv2).Move(ctx, req, grpc.Header(md))
		if traceHeader := md.Get("x-server-trace-id"); len(traceHeader) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Move dns_firewall x-server-trace-id: %s", traceHeader[0]))
		}
		if traceHeader := md.Get("x-server-request-id"); len(traceHeader) > 0 {
			tflog.Debug(ctx, fmt.Sprintf("Move dns_firewall x-server-request-id: %s", traceHeader[0]))
		}
		if err != nil {
			dg.AddError(
				"Failed to Read resource",
				"Error while requesting API to move dns_firewall:"+err.Error(),
			)
			return dg
		}
		moveRes, err := op.Wait(ctx)
		if err != nil {
			dg.AddError(
				"Unable to Move Resource",
				fmt.Sprintf("An unexpected error occurred while waiting longrunning response. "+
					"Please retry the operation or report this issue to the provider developers.\n\n"+
					"Error: %s", err),
			)
			return dg
		}
		tflog.Debug(ctx, fmt.Sprintf("Move dns_firewall response: %s", validate.ProtoDump(moveRes)))
	}
	return dg
}
