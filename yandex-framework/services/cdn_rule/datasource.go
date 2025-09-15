package cdn_rule

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	cdn_resource "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cdn_resource"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ datasource.DataSource              = &cdnRuleDataSource{}
	_ datasource.DataSourceWithConfigure = &cdnRuleDataSource{}
)

type cdnRuleDataSource struct {
	providerConfig *provider_config.Config
}

// NewDataSource creates a new CDN rule data source
func NewDataSource() datasource.DataSource {
	return &cdnRuleDataSource{}
}

func (d *cdnRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn_rule"
}

func (d *cdnRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = DataSourceCDNRuleSchema()
}

func (d *cdnRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.providerConfig = providerConfig
}

func (d *cdnRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CDNRuleDataSource
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading CDN rule data source")

	// resource_id is always required
	resourceID := state.ResourceID.ValueString()
	if resourceID == "" {
		resp.Diagnostics.AddError(
			"Missing required parameter",
			"resource_id must be specified",
		)
		return
	}

	// Determine rule ID: either from rule_id or resolve by name
	var ruleID int64
	if !state.RuleID.IsNull() && !state.RuleID.IsUnknown() && state.RuleID.ValueString() != "" {
		// Parse rule_id from string
		parsedID, err := strconv.ParseInt(state.RuleID.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid rule_id format",
				fmt.Sprintf("rule_id must be a valid integer, got: %s", state.RuleID.ValueString()),
			)
			return
		}
		ruleID = parsedID
		tflog.Debug(ctx, "Using provided rule_id", map[string]interface{}{
			"resource_id": resourceID,
			"rule_id":     ruleID,
		})
	} else if !state.Name.IsNull() && state.Name.ValueString() != "" {
		// Resolve by name using List
		resolvedID, err := d.resolveCDNRuleIDByName(ctx, resourceID, state.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to resolve CDN rule by name",
				fmt.Sprintf("Error resolving CDN rule by name %q: %s", state.Name.ValueString(), err),
			)
			return
		}
		ruleID = resolvedID
		tflog.Debug(ctx, "Resolved rule_id by name", map[string]interface{}{
			"resource_id": resourceID,
			"name":        state.Name.ValueString(),
			"rule_id":     ruleID,
		})
	} else {
		resp.Diagnostics.AddError(
			"Missing required parameter",
			"Either rule_id or name must be specified",
		)
		return
	}

	// Fetch CDN rule from API
	tflog.Debug(ctx, "Fetching CDN rule from API", map[string]interface{}{
		"resource_id": resourceID,
		"rule_id":     ruleID,
	})

	rule, err := d.providerConfig.SDK.CDN().ResourceRules().Get(ctx, &cdn.GetResourceRuleRequest{
		ResourceId: resourceID,
		RuleId:     ruleID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			resp.Diagnostics.AddError(
				"CDN rule not found",
				fmt.Sprintf("CDN rule with ID %d was not found in resource %s", ruleID, resourceID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to read CDN rule",
			fmt.Sprintf("Error reading CDN rule: %s", err),
		)
		return
	}

	// Convert API response to state
	state.ID = types.StringValue(fmt.Sprintf("%s/%d", resourceID, rule.Id))
	state.RuleID = types.StringValue(strconv.FormatInt(rule.Id, 10))
	state.ResourceID = types.StringValue(resourceID)
	state.Name = types.StringValue(rule.Name)
	state.RulePattern = types.StringValue(rule.RulePattern)
	state.Weight = types.Int64Value(rule.Weight)

	// Flatten options using existing function from cdn_resource
	// Pass empty list as plan options for data source (no disabled block preservation needed)
	// Using empty slice instead of null to avoid nil pointer issues
	var emptyPlanOptions types.List
	state.Options = cdn_resource.FlattenCDNResourceOptions(ctx, rule.Options, emptyPlanOptions, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Info(ctx, "Successfully read CDN rule data source", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// resolveCDNRuleIDByName resolves CDN rule ID by name using List API
func (d *cdnRuleDataSource) resolveCDNRuleIDByName(ctx context.Context, resourceID, name string) (int64, error) {
	if name == "" {
		return 0, fmt.Errorf("empty name for CDN rule")
	}

	tflog.Debug(ctx, "Resolving CDN rule ID by name", map[string]interface{}{
		"resource_id": resourceID,
		"name":        name,
	})

	// List all rules for the resource
	listResp, err := d.providerConfig.SDK.CDN().ResourceRules().List(ctx, &cdn.ListResourceRulesRequest{
		ResourceId: resourceID,
	})
	if err != nil {
		return 0, fmt.Errorf("error listing CDN rules: %w", err)
	}

	// Find rule by name
	for _, rule := range listResp.Rules {
		if name == rule.Name {
			tflog.Debug(ctx, "Found matching CDN rule", map[string]interface{}{
				"name":    name,
				"rule_id": rule.Id,
			})
			return rule.Id, nil
		}
	}

	return 0, fmt.Errorf("CDN rule with name %q not found in resource %s", name, resourceID)
}
