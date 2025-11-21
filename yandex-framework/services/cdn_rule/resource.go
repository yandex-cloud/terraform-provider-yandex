package cdn_rule

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	cdn_resource "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cdn_resource"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	yandexCDNRuleDefaultTimeout = 5 * time.Minute
)

// cdnRuleIDRegex is compiled once at package initialization for performance
// Format: resource_id/rule_id (e.g., "bc851ft45fne********/123")
var cdnRuleIDRegex = regexp.MustCompile(`^([^/]+)/(\d+)$`)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &cdnRuleResource{}
	_ resource.ResourceWithConfigure   = &cdnRuleResource{}
	_ resource.ResourceWithImportState = &cdnRuleResource{}
)

type cdnRuleResource struct {
	providerConfig *provider_config.Config
}

// NewResource creates a new CDN rule resource
func NewResource() resource.Resource {
	return &cdnRuleResource{}
}

func (r *cdnRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn_rule"
}

func (r *cdnRuleResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CDNRuleSchema(ctx)
}

func (r *cdnRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.providerConfig = providerConfig
}

func (r *cdnRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CDNRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexCDNRuleDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	tflog.Debug(ctx, "Creating CDN rule", map[string]interface{}{
		"resource_id":  plan.ResourceID.ValueString(),
		"name":         plan.Name.ValueString(),
		"rule_pattern": plan.RulePattern.ValueString(),
	})

	// Expand options using cdn_resource shared function
	options := expandOptions(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create rule
	request := &cdn.CreateResourceRuleRequest{
		ResourceId:  plan.ResourceID.ValueString(),
		Name:        plan.Name.ValueString(),
		RulePattern: plan.RulePattern.ValueString(),
		Weight:      plan.Weight.ValueInt64(),
		Options:     options,
	}

	// Debug logging: log the complete create request
	tflog.Debug(ctx, "CDN Rule Create Request Details", map[string]interface{}{
		"resource_id":  request.ResourceId,
		"name":         request.Name,
		"rule_pattern": request.RulePattern,
		"weight":       request.Weight,
		"options":      fmt.Sprintf("%+v", request.Options),
	})

	if request.Options != nil {
		tflog.Debug(ctx, "CDN Rule Create Request - Options Details", map[string]interface{}{
			"disable_proxy_force_ranges": fmt.Sprintf("%+v", request.Options.DisableProxyForceRanges),
			"slice":                      fmt.Sprintf("%+v", request.Options.Slice),
			"ignore_cookie":              fmt.Sprintf("%+v", request.Options.IgnoreCookie),
			"proxy_cache_methods_set":    fmt.Sprintf("%+v", request.Options.ProxyCacheMethodsSet),
			"edge_cache_settings":        fmt.Sprintf("%+v", request.Options.EdgeCacheSettings),
			"browser_cache_settings":     fmt.Sprintf("%+v", request.Options.BrowserCacheSettings),
			"compression_options":        fmt.Sprintf("%+v", request.Options.CompressionOptions),
			"redirect_options":           fmt.Sprintf("%+v", request.Options.RedirectOptions),
			"host_options":               fmt.Sprintf("%+v", request.Options.HostOptions),
			"query_params_options":       fmt.Sprintf("%+v", request.Options.QueryParamsOptions),
		})
	}

	tflog.Debug(ctx, "Calling CDN ResourceRules.Create API")
	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.CDN().ResourceRules().Create(ctx, request),
	)
	if err != nil {
		tflog.Error(ctx, "CDN Rule Create API call failed", map[string]interface{}{
			"error":        err.Error(),
			"resource_id":  request.ResourceId,
			"name":         request.Name,
			"rule_pattern": request.RulePattern,
		})
		resp.Diagnostics.AddError(
			"Failed to create CDN rule",
			fmt.Sprintf("Error requesting API to create CDN rule: %s", err),
		)
		return
	}

	// Get rule ID from operation metadata
	protoMetadata, err := op.Metadata()
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to get operation metadata",
			fmt.Sprintf("Error getting CDN rule create operation metadata: %s", err),
		)
		return
	}

	md, ok := protoMetadata.(*cdn.CreateResourceRuleMetadata)
	if !ok {
		resp.Diagnostics.AddError(
			"Failed to parse operation metadata",
			"Could not get CDN rule ID from create operation metadata",
		)
		return
	}

	// Set composite ID immediately
	ruleIDStr := strconv.FormatInt(md.RuleId, 10)
	compositeID := fmt.Sprintf("%s/%s", md.ResourceId, ruleIDStr)
	plan.ID = types.StringValue(compositeID)
	plan.RuleID = types.StringValue(ruleIDStr)

	tflog.Debug(ctx, "Waiting for CDN rule create operation", map[string]interface{}{
		"id": compositeID,
	})

	// Wait for operation
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Operation failed",
			fmt.Sprintf("Error waiting for CDN rule create operation: %s", err),
		)
		return
	}

	if _, err := op.Response(); err != nil {
		resp.Diagnostics.AddError(
			"Operation response error",
			fmt.Sprintf("Error getting CDN rule create operation response: %s", err),
		)
		return
	}

	tflog.Info(ctx, "CDN rule created successfully", map[string]interface{}{
		"id": compositeID,
	})

	// Read back the rule to get full state
	r.readRule(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdnRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CDNRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	readTimeout, diags := state.Timeouts.Read(ctx, yandexCDNRuleDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	r.readRule(ctx, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *cdnRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CDNRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CDNRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexCDNRuleDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	resourceID, ruleID, err := parseCDNRuleID(plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Error parsing CDN rule ID: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Updating CDN rule", map[string]interface{}{
		"id":   plan.ID.ValueString(),
		"name": plan.Name.ValueString(),
	})

	// Handle options update
	// CRITICAL: ResourceOptions uses "replace semantics" (proto analysis confirmed)
	// When we send Options to API, it REPLACES ALL options completely.
	// Therefore we MUST merge plan + state to preserve Optional+Computed fields.
	//
	// This is the SAME pattern as cdn_resource uses (resource.go:347-497)

	optionsEqual := plan.Options.Equal(state.Options)
	tflog.Debug(ctx, "CDN rule options comparison", map[string]interface{}{
		"plan_is_null":  plan.Options.IsNull(),
		"state_is_null": state.Options.IsNull(),
		"options_equal": optionsEqual,
	})

	// Expand options - merge plan + state if changed
	var options *cdn.ResourceOptions
	mergedPlan := plan

	if !optionsEqual {
		tflog.Debug(ctx, "Options changed - merging plan with state for completeness")

		// Unpack plan and state options for merging
		var planOptionsModels []cdn_resource.CDNOptionsModel
		var stateOptionsModels []cdn_resource.CDNOptionsModel

		if !plan.Options.IsNull() && len(plan.Options.Elements()) > 0 {
			resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &planOptionsModels, false)...)
		}
		if !state.Options.IsNull() && len(state.Options.Elements()) > 0 {
			resp.Diagnostics.Append(state.Options.ElementsAs(ctx, &stateOptionsModels, false)...)
		}

		if resp.Diagnostics.HasError() {
			return
		}

		// Merge: use plan values where available, fallback to state for null/unknown fields
		if len(planOptionsModels) > 0 {
			mergedOptions := planOptionsModels[0]

			// If state has options, fill in null/unknown computed fields from state
			if len(stateOptionsModels) > 0 {
				stateOpt := stateOptionsModels[0]

				// For each Optional+Computed field: use plan if not null, otherwise use state
				// This preserves previously-set values that aren't being changed
				if (mergedOptions.EdgeCacheSettings.IsNull() || mergedOptions.EdgeCacheSettings.IsUnknown()) && !stateOpt.EdgeCacheSettings.IsNull() {
					mergedOptions.EdgeCacheSettings = stateOpt.EdgeCacheSettings
				}
				if (mergedOptions.BrowserCacheSettings.IsNull() || mergedOptions.BrowserCacheSettings.IsUnknown()) && !stateOpt.BrowserCacheSettings.IsNull() {
					mergedOptions.BrowserCacheSettings = stateOpt.BrowserCacheSettings
				}
				if (mergedOptions.GzipOn.IsNull() || mergedOptions.GzipOn.IsUnknown()) && !stateOpt.GzipOn.IsNull() {
					mergedOptions.GzipOn = stateOpt.GzipOn
				}
				if (mergedOptions.RedirectHttpToHttps.IsNull() || mergedOptions.RedirectHttpToHttps.IsUnknown()) && !stateOpt.RedirectHttpToHttps.IsNull() {
					mergedOptions.RedirectHttpToHttps = stateOpt.RedirectHttpToHttps
				}
				if (mergedOptions.DisableProxyForceRanges.IsNull() || mergedOptions.DisableProxyForceRanges.IsUnknown()) && !stateOpt.DisableProxyForceRanges.IsNull() {
					mergedOptions.DisableProxyForceRanges = stateOpt.DisableProxyForceRanges
				}
				if (mergedOptions.Rewrite.IsNull() || mergedOptions.Rewrite.IsUnknown()) && !stateOpt.Rewrite.IsNull() {
					mergedOptions.Rewrite = stateOpt.Rewrite
				}
				// Add other fields as needed...
			}

			// Pack merged options back into mergedPlan
			optionsList, d := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: cdn_resource.GetCDNOptionsAttrTypes(),
			}, []cdn_resource.CDNOptionsModel{mergedOptions})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			mergedPlan.Options = optionsList
		}

		// Expand merged options for API request
		options = expandOptions(ctx, &mergedPlan, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		tflog.Debug(ctx, "Options merged and expanded for API request")
	} else {
		tflog.Debug(ctx, "Options unchanged - skipping options update (sending nil)")
		options = nil
	}

	// CRITICAL: Name and RulePattern in proto3 UpdateResourceRuleRequest are optional,
	// BUT empty strings ("") are INVALID and cause Internal errors on server validation.
	// Proto3 doesn't distinguish "unset" vs "empty string" for regular string fields.
	// Solution: Fallback to state values if plan contains empty strings.
	name := plan.Name.ValueString()
	if name == "" {
		tflog.Debug(ctx, "Name is empty in plan, using state value")
		name = state.Name.ValueString()
	}

	rulePattern := plan.RulePattern.ValueString()
	if rulePattern == "" {
		tflog.Debug(ctx, "RulePattern is empty in plan, using state value")
		rulePattern = state.RulePattern.ValueString()
	}

	// Update rule
	updateReq := &cdn.UpdateResourceRuleRequest{
		ResourceId:  resourceID,
		RuleId:      ruleID,
		Name:        name,        // Always valid (not empty)
		RulePattern: rulePattern, // Always valid (not empty)
		Options:     options,
	}

	// Weight is always sent in update (not a pointer in the proto)
	weight := plan.Weight.ValueInt64()
	updateReq.Weight = &weight

	// Debug logging: log the complete update request
	tflog.Debug(ctx, "CDN Rule Update Request Details", map[string]interface{}{
		"resource_id":  updateReq.ResourceId,
		"rule_id":      updateReq.RuleId,
		"name":         updateReq.Name,
		"rule_pattern": updateReq.RulePattern,
		"weight":       *updateReq.Weight,
		"options":      fmt.Sprintf("%+v", updateReq.Options),
	})

	if updateReq.Options != nil {
		tflog.Debug(ctx, "CDN Rule Update Request - Options Details", map[string]interface{}{
			"disable_proxy_force_ranges": fmt.Sprintf("%+v", updateReq.Options.DisableProxyForceRanges),
			"slice":                      fmt.Sprintf("%+v", updateReq.Options.Slice),
			"ignore_cookie":              fmt.Sprintf("%+v", updateReq.Options.IgnoreCookie),
			"proxy_cache_methods_set":    fmt.Sprintf("%+v", updateReq.Options.ProxyCacheMethodsSet),
			"edge_cache_settings":        fmt.Sprintf("%+v", updateReq.Options.EdgeCacheSettings),
			"browser_cache_settings":     fmt.Sprintf("%+v", updateReq.Options.BrowserCacheSettings),
			"compression_options":        fmt.Sprintf("%+v", updateReq.Options.CompressionOptions),
			"redirect_options":           fmt.Sprintf("%+v", updateReq.Options.RedirectOptions),
			"host_options":               fmt.Sprintf("%+v", updateReq.Options.HostOptions),
			"query_params_options":       fmt.Sprintf("%+v", updateReq.Options.QueryParamsOptions),
		})
	}

	tflog.Debug(ctx, "Calling CDN ResourceRules.Update API")
	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.CDN().ResourceRules().Update(ctx, updateReq),
	)
	if err != nil {
		tflog.Error(ctx, "CDN Rule Update API call failed", map[string]interface{}{
			"error":        err.Error(),
			"resource_id":  updateReq.ResourceId,
			"rule_id":      updateReq.RuleId,
			"name":         updateReq.Name,
			"rule_pattern": updateReq.RulePattern,
		})
		resp.Diagnostics.AddError(
			"Failed to update CDN rule",
			fmt.Sprintf("Error requesting API to update CDN rule: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Waiting for CDN rule update operation")
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Operation failed",
			fmt.Sprintf("Error waiting for CDN rule update operation: %s", err),
		)
		return
	}

	if _, err := op.Response(); err != nil {
		resp.Diagnostics.AddError(
			"Operation response error",
			fmt.Sprintf("Error getting CDN rule update operation response: %s", err),
		)
		return
	}

	tflog.Info(ctx, "CDN rule updated successfully", map[string]interface{}{
		"id": plan.ID.ValueString(),
	})

	// Read back to get updated state
	r.readRule(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdnRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CDNRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexCDNRuleDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	resourceID, ruleID, err := parseCDNRuleID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid ID format",
			fmt.Sprintf("Error parsing CDN rule ID: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Deleting CDN rule", map[string]interface{}{
		"id": state.ID.ValueString(),
	})

	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.CDN().ResourceRules().Delete(ctx, &cdn.DeleteResourceRuleRequest{
			ResourceId: resourceID,
			RuleId:     ruleID,
		}),
	)

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Debug(ctx, "CDN rule already deleted")
			return
		}
		resp.Diagnostics.AddError(
			"Failed to delete CDN rule",
			fmt.Sprintf("Error requesting API to delete CDN rule: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Waiting for CDN rule delete operation")
	err = op.Wait(ctx)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Debug(ctx, "CDN rule already deleted during operation wait")
			return
		}
		resp.Diagnostics.AddError(
			"Operation failed",
			fmt.Sprintf("Error waiting for CDN rule delete operation: %s", err),
		)
		return
	}

	if _, err := op.Response(); err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			tflog.Debug(ctx, "CDN rule already deleted in operation response")
			return
		}
		resp.Diagnostics.AddError(
			"Operation response error",
			fmt.Sprintf("Error getting CDN rule delete operation response: %s", err),
		)
		return
	}

	tflog.Info(ctx, "CDN rule deleted successfully", map[string]interface{}{
		"id": state.ID.ValueString(),
	})
}

func (r *cdnRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: resource_id/rule_id
	resourceID, ruleID, err := parseCDNRuleID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID format",
			fmt.Sprintf("Expected format: resource_id/rule_id, got: %s", req.ID),
		)
		return
	}

	tflog.Debug(ctx, "Importing CDN rule", map[string]interface{}{
		"id":          req.ID,
		"resource_id": resourceID,
		"rule_id":     ruleID,
	})

	// Set the ID and resource_id
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_id"), resourceID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("rule_id"), strconv.FormatInt(ruleID, 10))...)
}

// readRule reads the CDN rule from the API and updates the model
func (r *cdnRuleResource) readRule(ctx context.Context, model *CDNRuleModel, diags *diag.Diagnostics) {
	resourceID, ruleID, err := parseCDNRuleID(model.ID.ValueString())
	if err != nil {
		diags.AddError(
			"Invalid ID format",
			fmt.Sprintf("Error parsing CDN rule ID: %s", err),
		)
		return
	}

	tflog.Debug(ctx, "Reading CDN rule from API", map[string]interface{}{
		"resource_id": resourceID,
		"rule_id":     ruleID,
	})

	rule, err := r.providerConfig.SDK.CDN().ResourceRules().Get(ctx, &cdn.GetResourceRuleRequest{
		ResourceId: resourceID,
		RuleId:     ruleID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			diags.AddError(
				"CDN rule not found",
				fmt.Sprintf("CDN rule %s was not found and may have been deleted", model.ID.ValueString()),
			)
			return
		}
		diags.AddError(
			"Failed to read CDN rule",
			fmt.Sprintf("Error reading CDN rule: %s", err),
		)
		return
	}

	// Flatten the rule into model
	model.ResourceID = types.StringValue(resourceID)
	model.RuleID = types.StringValue(strconv.FormatInt(rule.Id, 10))
	model.Name = types.StringValue(rule.Name)
	model.RulePattern = types.StringValue(rule.RulePattern)
	model.Weight = types.Int64Value(rule.Weight)

	// Flatten options using cdn_resource shared function
	model.Options = flattenOptions(ctx, rule.Options, diags)
	if diags.HasError() {
		return
	}

	tflog.Debug(ctx, "Successfully read CDN rule", map[string]interface{}{
		"id":   model.ID.ValueString(),
		"name": model.Name.ValueString(),
	})
}

// parseCDNRuleID parses composite ID format: resource_id/rule_id
// Uses package-level compiled regex for performance
func parseCDNRuleID(id string) (string, int64, error) {
	parts := cdnRuleIDRegex.FindStringSubmatch(id)
	if len(parts) != 3 {
		return "", 0, fmt.Errorf("invalid CDN rule ID format: %s (expected: resource_id/rule_id)", id)
	}

	ruleID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid rule ID in CDN rule ID %s: %w", id, err)
	}

	return parts[1], ruleID, nil
}
