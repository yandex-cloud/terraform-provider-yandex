package cdn_resource

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/cdn/v1"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	yandexCDNResourceDefaultTimeout = 30 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &cdnResourceResource{}
	_ resource.ResourceWithConfigure   = &cdnResourceResource{}
	_ resource.ResourceWithImportState = &cdnResourceResource{}
)

type cdnResourceResource struct {
	providerConfig *provider_config.Config
}

// NewResource creates a new CDN resource
func NewResource() resource.Resource {
	return &cdnResourceResource{}
}

func (r *cdnResourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cdn_resource"
}

func (r *cdnResourceResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = CDNResourceSchema(ctx)
}

func (r *cdnResourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *cdnResourceResource) UpgradeState(context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			// PriorSchema is intentionally nil - we use RawState migration
			StateUpgrader: upgradeStateV0ToV1,
		},
		1: {
			// PriorSchema is intentionally nil - we use RawState migration
			// Migrates edge_cache_settings.cache_time: "*" → "any"
			StateUpgrader: upgradeStateV1ToV2,
		},
		2: {
			// PriorSchema is intentionally nil - we use RawState migration
			// Migrates ssl_certificate Set → List (MaxItems: 1)
			StateUpgrader: upgradeStateV2ToV3,
		},
	}
}

func (r *cdnResourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CDNResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexCDNResourceDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	// Resolve origin group ID
	originGroupID, err := r.resolveOriginGroupID(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError("Failed to resolve origin group", err.Error())
		return
	}

	// Determine origin protocol
	originProtocol := expandOriginProtocol(ctx, plan.OriginProtocol.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get folder ID (from resource or provider config)
	folderID := r.getFolderID(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := &cdn.CreateResourceRequest{
		FolderId: folderID,
		Cname:    plan.Cname.ValueString(),
		Origin: &cdn.CreateResourceRequest_Origin{
			OriginVariant: &cdn.CreateResourceRequest_Origin_OriginGroupId{
				OriginGroupId: originGroupID,
			},
		},
		Active: &wrapperspb.BoolValue{
			Value: plan.Active.ValueBool(),
		},
		OriginProtocol: originProtocol,
	}

	// Add provider type if specified
	if !plan.ProviderType.IsNull() && plan.ProviderType.ValueString() != "" {
		createReq.ProviderType = plan.ProviderType.ValueString()
	}

	// Add labels if specified
	if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() && len(plan.Labels.Elements()) > 0 {
		labels := make(map[string]string)
		resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Labels = labels
	}

	// Add secondary hostnames if specified
	if !plan.SecondaryHostnames.IsNull() && len(plan.SecondaryHostnames.Elements()) > 0 {
		var hostnames []string
		resp.Diagnostics.Append(plan.SecondaryHostnames.ElementsAs(ctx, &hostnames, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.SecondaryHostnames = &cdn.SecondaryHostnames{
			Values: hostnames,
		}
	}

	// Add SSL certificate if specified
	if !plan.SSLCertificate.IsNull() && len(plan.SSLCertificate.Elements()) > 0 {
		var certs []SSLCertificateModel
		resp.Diagnostics.Append(plan.SSLCertificate.ElementsAs(ctx, &certs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(certs) > 0 {
			cert := certs[0]
			sslCert := &cdn.SSLTargetCertificate{}

			switch cert.Type.ValueString() {
			case "not_used":
				sslCert.Type = cdn.SSLCertificateType_DONT_USE
			case "certificate_manager":
				sslCert.Type = cdn.SSLCertificateType_CM
				if !cert.CertificateManagerID.IsNull() {
					sslCert.Data = &cdn.SSLCertificateData{
						SslCertificateDataVariant: &cdn.SSLCertificateData_Cm{
							Cm: &cdn.SSLCertificateCMData{
								Id: cert.CertificateManagerID.ValueString(),
							},
						},
					}
				}
			case "lets_encrypt":
				sslCert.Type = cdn.SSLCertificateType_LETS_ENCRYPT_GCORE
			}

			createReq.SslCertificate = sslCert
		}
	}

	// Add options if specified
	if !plan.Options.IsNull() && len(plan.Options.Elements()) > 0 {
		var optionsModels []CDNOptionsModel
		resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &optionsModels, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		options := ExpandCDNResourceOptions(ctx, optionsModels, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Options = options
	}

	tflog.Debug(ctx, "Creating CDN resource", map[string]interface{}{
		"cname":           createReq.Cname,
		"origin_protocol": createReq.OriginProtocol,
	})

	// Call API
	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.CDN().Resource().Create(ctx, createReq),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create CDN resource", err.Error())
		return
	}

	// Wait for operation to complete
	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for CDN resource creation", err.Error())
		return
	}

	// Get resource metadata from operation
	metadata, err := op.Metadata()
	if err != nil {
		resp.Diagnostics.AddError("Failed to get operation metadata", err.Error())
		return
	}

	resourceMetadata, ok := metadata.(*cdn.CreateResourceMetadata)
	if !ok {
		resp.Diagnostics.AddError(
			"Invalid operation metadata",
			fmt.Sprintf("Expected *cdn.CreateResourceMetadata, got: %T", metadata),
		)
		return
	}

	plan.ID = types.StringValue(resourceMetadata.ResourceId)

	// Apply shielding configuration if specified
	if err := applyShieldingFromPlan(ctx, &plan, r.providerConfig.SDK); err != nil {
		resp.Diagnostics.AddError("Failed to configure shielding", err.Error())
		return
	}

	// Save plan options to preserve disabled cache blocks during readResourceToState
	// CRITICAL: Must save before calling readResourceToState which modifies plan
	originalPlanOptions := plan.Options

	// Read the created resource to populate computed fields
	if !r.readResourceToState(ctx, &plan, originalPlanOptions, &resp.Diagnostics) {
		resp.Diagnostics.AddError(
			"Failed to read created resource",
			"Resource was created but could not be read",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *cdnResourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CDNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state options to preserve disabled cache blocks during refresh
	// CRITICAL: Disabled blocks represent user's explicit configuration (enabled=false)
	originalStateOptions := state.Options

	if r.readResourceToState(ctx, &state, originalStateOptions, &resp.Diagnostics) {
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
	} else {
		// Resource not found, remove from state
		resp.State.RemoveResource(ctx)
	}
}

func (r *cdnResourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CDNResourceModel
	var state CDNResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexCDNResourceDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	// Build update request
	updateReq := &cdn.UpdateResourceRequest{
		ResourceId: state.ID.ValueString(),
		// UpdateMask will be populated based on changed fields
	}

	// Check for changes and build update request
	hasChanges := false

	if !plan.Active.Equal(state.Active) {
		updateReq.Active = &wrapperspb.BoolValue{
			Value: plan.Active.ValueBool(),
		}
		hasChanges = true
	}

	if !plan.SecondaryHostnames.Equal(state.SecondaryHostnames) {
		var hostnames []string
		if !plan.SecondaryHostnames.IsNull() {
			resp.Diagnostics.Append(plan.SecondaryHostnames.ElementsAs(ctx, &hostnames, false)...)
		}
		updateReq.SecondaryHostnames = &cdn.SecondaryHostnames{
			Values: hostnames,
		}
		hasChanges = true
	}

	if !plan.Labels.Equal(state.Labels) {
		labels := make(map[string]string)
		if !plan.Labels.IsNull() && !plan.Labels.IsUnknown() {
			resp.Diagnostics.Append(plan.Labels.ElementsAs(ctx, &labels, false)...)
			updateReq.Labels = labels
			hasChanges = true
		}
	}

	if !plan.OriginProtocol.Equal(state.OriginProtocol) {
		originProtocol := expandOriginProtocol(ctx, plan.OriginProtocol.ValueString(), &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.OriginProtocol = originProtocol
		hasChanges = true
	}

	// Handle SSL certificate update
	if !plan.SSLCertificate.Equal(state.SSLCertificate) {
		var certs []SSLCertificateModel
		if !plan.SSLCertificate.IsNull() && len(plan.SSLCertificate.Elements()) > 0 {
			resp.Diagnostics.Append(plan.SSLCertificate.ElementsAs(ctx, &certs, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
		}
		if len(certs) > 0 {
			cert := certs[0]
			sslCert := &cdn.SSLTargetCertificate{}
			switch cert.Type.ValueString() {
			case "not_used":
				sslCert.Type = cdn.SSLCertificateType_DONT_USE
			case "certificate_manager":
				sslCert.Type = cdn.SSLCertificateType_CM
				if !cert.CertificateManagerID.IsNull() && cert.CertificateManagerID.ValueString() != "" {
					sslCert.Data = &cdn.SSLCertificateData{
						SslCertificateDataVariant: &cdn.SSLCertificateData_Cm{
							Cm: &cdn.SSLCertificateCMData{
								Id: cert.CertificateManagerID.ValueString(),
							},
						},
					}
				}
			case "lets_encrypt":
				sslCert.Type = cdn.SSLCertificateType_LETS_ENCRYPT_GCORE
			default:
				// Unknown type, do not set change
				sslCert = nil
			}
			if sslCert != nil {
				updateReq.SslCertificate = sslCert
				hasChanges = true
			}
		}
	}

	// Handle options update
	// CRITICAL: API uses "replace" semantics for Options - when we send Options, API replaces ALL options
	// Therefore we MUST send COMPLETE options block with ALL fields, not just changed ones
	//
	// For migrated resources from SDKv2, Optional+Computed fields may be null in plan but not null in state
	// We merge plan + state: use plan values where available, fall back to state for computed fields
	if !plan.Options.Equal(state.Options) {
		// Unpack both plan and state options
		var planOptionsModels []CDNOptionsModel
		var stateOptionsModels []CDNOptionsModel

		if !plan.Options.IsNull() && len(plan.Options.Elements()) > 0 {
			resp.Diagnostics.Append(plan.Options.ElementsAs(ctx, &planOptionsModels, false)...)
		}
		if !state.Options.IsNull() && len(state.Options.Elements()) > 0 {
			resp.Diagnostics.Append(state.Options.ElementsAs(ctx, &stateOptionsModels, false)...)
		}

		if resp.Diagnostics.HasError() {
			return
		}

		// Merge plan with state - for each field: use plan if not null, otherwise use state
		// This ensures we send complete options to API (avoiding reset of unmentioned fields)
		var mergedOptions CDNOptionsModel
		if len(planOptionsModels) > 0 {
			mergedOptions = planOptionsModels[0]

			// If we have state options, fill in null/unknown computed fields from state
			// CRITICAL: Must check both IsNull() and IsUnknown() for Optional+Computed fields
			// IsUnknown() = true when field is computed and will be "known after apply"
			if len(stateOptionsModels) > 0 {
				stateOpt := stateOptionsModels[0]

				// Independent Boolean options
				if (mergedOptions.Slice.IsNull() || mergedOptions.Slice.IsUnknown()) && !stateOpt.Slice.IsNull() {
					mergedOptions.Slice = stateOpt.Slice
				}
				if (mergedOptions.IgnoreCookie.IsNull() || mergedOptions.IgnoreCookie.IsUnknown()) && !stateOpt.IgnoreCookie.IsNull() {
					mergedOptions.IgnoreCookie = stateOpt.IgnoreCookie
				}
				if (mergedOptions.ProxyCacheMethodsSet.IsNull() || mergedOptions.ProxyCacheMethodsSet.IsUnknown()) && !stateOpt.ProxyCacheMethodsSet.IsNull() {
					mergedOptions.ProxyCacheMethodsSet = stateOpt.ProxyCacheMethodsSet
				}
				if (mergedOptions.DisableProxyForceRanges.IsNull() || mergedOptions.DisableProxyForceRanges.IsUnknown()) && !stateOpt.DisableProxyForceRanges.IsNull() {
					mergedOptions.DisableProxyForceRanges = stateOpt.DisableProxyForceRanges
				}

				// Integer options
				if (mergedOptions.EdgeCacheSettings.IsNull() || mergedOptions.EdgeCacheSettings.IsUnknown()) && !stateOpt.EdgeCacheSettings.IsNull() {
					mergedOptions.EdgeCacheSettings = stateOpt.EdgeCacheSettings
				}
				if (mergedOptions.BrowserCacheSettings.IsNull() || mergedOptions.BrowserCacheSettings.IsUnknown()) && !stateOpt.BrowserCacheSettings.IsNull() {
					mergedOptions.BrowserCacheSettings = stateOpt.BrowserCacheSettings
				}

				// String options
				if (mergedOptions.CustomServerName.IsNull() || mergedOptions.CustomServerName.IsUnknown()) && !stateOpt.CustomServerName.IsNull() {
					mergedOptions.CustomServerName = stateOpt.CustomServerName
				}
				if (mergedOptions.SecureKey.IsNull() || mergedOptions.SecureKey.IsUnknown()) && !stateOpt.SecureKey.IsNull() {
					mergedOptions.SecureKey = stateOpt.SecureKey
					mergedOptions.EnableIPURLSigning = stateOpt.EnableIPURLSigning
				}

				// List options
				// DEPRECATED: cache_http_headers - always null, not merged
				mergedOptions.CacheHTTPHeaders = types.ListNull(types.StringType)

				if (mergedOptions.Cors.IsNull() || mergedOptions.Cors.IsUnknown()) && !stateOpt.Cors.IsNull() {
					mergedOptions.Cors = stateOpt.Cors
				}

				// Map options - DO NOT merge, user explicitly controls these
				// Null in plan means "delete/disable", not "preserve state"
				// If user wants to keep state value, they should specify it in config

				// Mutually exclusive groups - these should be in plan, but just in case
				// For mutually exclusive groups, ALL fields must be null/unknown to fallback to state
				allHostFieldsEmpty := (mergedOptions.CustomHostHeader.IsNull() || mergedOptions.CustomHostHeader.IsUnknown()) &&
					(mergedOptions.ForwardHostHeader.IsNull() || mergedOptions.ForwardHostHeader.IsUnknown())
				if allHostFieldsEmpty {
					if !stateOpt.CustomHostHeader.IsNull() {
						mergedOptions.CustomHostHeader = stateOpt.CustomHostHeader
					}
					if !stateOpt.ForwardHostHeader.IsNull() {
						mergedOptions.ForwardHostHeader = stateOpt.ForwardHostHeader
					}
				}

				allQueryParamsEmpty := (mergedOptions.IgnoreQueryParams.IsNull() || mergedOptions.IgnoreQueryParams.IsUnknown()) &&
					(mergedOptions.QueryParamsWhitelist.IsNull() || mergedOptions.QueryParamsWhitelist.IsUnknown()) &&
					(mergedOptions.QueryParamsBlacklist.IsNull() || mergedOptions.QueryParamsBlacklist.IsUnknown())
				if allQueryParamsEmpty {
					if !stateOpt.IgnoreQueryParams.IsNull() {
						mergedOptions.IgnoreQueryParams = stateOpt.IgnoreQueryParams
					}
					if !stateOpt.QueryParamsWhitelist.IsNull() {
						mergedOptions.QueryParamsWhitelist = stateOpt.QueryParamsWhitelist
					}
					if !stateOpt.QueryParamsBlacklist.IsNull() {
						mergedOptions.QueryParamsBlacklist = stateOpt.QueryParamsBlacklist
					}
				}

				allCompressionEmpty := (mergedOptions.GzipOn.IsNull() || mergedOptions.GzipOn.IsUnknown()) &&
					(mergedOptions.FetchedCompressed.IsNull() || mergedOptions.FetchedCompressed.IsUnknown())
				if allCompressionEmpty {
					if !stateOpt.GzipOn.IsNull() {
						mergedOptions.GzipOn = stateOpt.GzipOn
					}
					if !stateOpt.FetchedCompressed.IsNull() {
						mergedOptions.FetchedCompressed = stateOpt.FetchedCompressed
					}
				}

				allRedirectEmpty := (mergedOptions.RedirectHttpToHttps.IsNull() || mergedOptions.RedirectHttpToHttps.IsUnknown()) &&
					(mergedOptions.RedirectHttpsToHttp.IsNull() || mergedOptions.RedirectHttpsToHttp.IsUnknown())
				if allRedirectEmpty {
					if !stateOpt.RedirectHttpToHttps.IsNull() {
						mergedOptions.RedirectHttpToHttps = stateOpt.RedirectHttpToHttps
					}
					if !stateOpt.RedirectHttpsToHttp.IsNull() {
						mergedOptions.RedirectHttpsToHttp = stateOpt.RedirectHttpsToHttp
					}
				}

				// Nested blocks
				if (mergedOptions.IPAddressACL.IsNull() || mergedOptions.IPAddressACL.IsUnknown()) && !stateOpt.IPAddressACL.IsNull() {
					mergedOptions.IPAddressACL = stateOpt.IPAddressACL
				}
				if (mergedOptions.Rewrite.IsNull() || mergedOptions.Rewrite.IsUnknown()) && !stateOpt.Rewrite.IsNull() {
					mergedOptions.Rewrite = stateOpt.Rewrite
				}
			}
		} else if len(stateOptionsModels) > 0 {
			// No plan options, use state
			mergedOptions = stateOptionsModels[0]
		}

		options := ExpandCDNResourceOptions(ctx, []CDNOptionsModel{mergedOptions}, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq.Options = options
		hasChanges = true
	}

	if resp.Diagnostics.HasError() {
		return
	}

	if hasChanges {
		tflog.Debug(ctx, "Updating CDN resource", map[string]interface{}{
			"resource_id": updateReq.ResourceId,
		})

		op, err := r.providerConfig.SDK.WrapOperation(
			r.providerConfig.SDK.CDN().Resource().Update(ctx, updateReq),
		)
		if err != nil {
			resp.Diagnostics.AddError("Failed to update CDN resource", err.Error())
			return
		}

		err = op.Wait(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to wait for CDN resource update", err.Error())
			return
		}
	}

	// Update shielding configuration if changed
	if err := updateShieldingIfChanged(ctx, &plan, &state, r.providerConfig.SDK); err != nil {
		resp.Diagnostics.AddError("Failed to update shielding", err.Error())
		return
	}

	// Create new state model from plan to preserve timeouts and ID
	newState := CDNResourceModel{
		ID:       plan.ID,
		Timeouts: plan.Timeouts,
	}

	// Save plan options before reading
	originalPlanOptions := plan.Options

	// Read updated resource into new state (not plan!)
	// This prevents "Provider produced inconsistent result" error for computed fields like updated_at
	if !r.readResourceToState(ctx, &newState, originalPlanOptions, &resp.Diagnostics) {
		resp.Diagnostics.AddError(
			"Failed to read updated resource",
			"Resource was updated but could not be read",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *cdnResourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CDNResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexCDNResourceDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	tflog.Debug(ctx, "Deleting CDN resource", map[string]interface{}{
		"resource_id": state.ID.ValueString(),
	})

	deleteReq := &cdn.DeleteResourceRequest{
		ResourceId: state.ID.ValueString(),
	}

	op, err := r.providerConfig.SDK.WrapOperation(
		r.providerConfig.SDK.CDN().Resource().Delete(ctx, deleteReq),
	)
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete CDN resource", err.Error())
		return
	}

	err = op.Wait(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to wait for CDN resource deletion", err.Error())
		return
	}
}

func (r *cdnResourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// Helper functions

// readResourceToState reads CDN resource from API and populates state
// planOptionsForDisabledBlocks: optional plan options to preserve disabled cache blocks (types.ListNull during Read)
// Returns false if resource not found
func (r *cdnResourceResource) readResourceToState(ctx context.Context, state *CDNResourceModel, planOptionsForDisabledBlocks types.List, diags *diag.Diagnostics) bool {
	getReq := &cdn.GetResourceRequest{
		ResourceId: state.ID.ValueString(),
	}

	resource, err := r.providerConfig.SDK.CDN().Resource().Get(ctx, getReq)
	if err != nil {
		tflog.Debug(ctx, "CDN resource not found", map[string]interface{}{
			"resource_id": state.ID.ValueString(),
			"error":       err.Error(),
		})
		return false
	}

	// Populate state from API response
	state.Cname = types.StringValue(resource.Cname)
	state.FolderID = types.StringValue(resource.FolderId)
	state.Active = types.BoolValue(resource.Active)
	state.OriginGroupID = types.StringValue(fmt.Sprintf("%d", resource.OriginGroupId))
	state.OriginGroupName = types.StringValue(resource.OriginGroupName)
	state.ProviderCname = types.StringValue(resource.ProviderCname)
	state.ProviderType = types.StringValue(resource.ProviderType)
	state.CreatedAt = types.StringValue(resource.CreatedAt.AsTime().Format(time.RFC3339))
	state.UpdatedAt = types.StringValue(resource.UpdatedAt.AsTime().Format(time.RFC3339))

	// Origin Protocol - convert enum to string
	state.OriginProtocol = flattenOriginProtocol(ctx, resource.OriginProtocol, diags)

	// Labels - TOP-LEVEL FIELD: SDKv2 always set this field (even empty)
	// So we must return empty map instead of null for backward compatibility
	labels := resource.Labels
	if labels == nil {
		labels = map[string]string{}
	}
	labelsMap, d := types.MapValueFrom(ctx, types.StringType, labels)
	diags.Append(d...)
	state.Labels = labelsMap

	// Secondary hostnames - TOP-LEVEL FIELD: SDKv2 always set this field (even empty)
	// So we must return empty Set instead of null for backward compatibility
	hostnames := resource.SecondaryHostnames
	if hostnames == nil {
		hostnames = []string{}
	}
	hostnamesSet, d := types.SetValueFrom(ctx, types.StringType, hostnames)
	diags.Append(d...)
	state.SecondaryHostnames = hostnamesSet

	// SSL Certificate
	state.SSLCertificate = flattenSSLCertificate(ctx, resource.SslCertificate, diags)

	// Shielding - fetch from separate API
	shieldingLocation, err := getShieldingLocation(ctx, state.ID.ValueString(), r.providerConfig.SDK)
	if err != nil {
		diags.AddError("Failed to read shielding configuration", err.Error())
		return false
	}
	state.Shielding = flattenShielding(shieldingLocation)

	// Options - CRITICAL: Pass plan options to preserve disabled cache blocks
	state.Options = FlattenCDNResourceOptions(ctx, resource.Options, planOptionsForDisabledBlocks, diags)

	return !diags.HasError()
}

// resolveOriginGroupID resolves origin group ID from either ID or name
func (r *cdnResourceResource) resolveOriginGroupID(ctx context.Context, plan *CDNResourceModel) (int64, error) {
	if !plan.OriginGroupID.IsNull() && plan.OriginGroupID.ValueString() != "" {
		var id int64
		_, err := fmt.Sscanf(plan.OriginGroupID.ValueString(), "%d", &id)
		if err != nil {
			return 0, fmt.Errorf("invalid origin_group_id format: %w", err)
		}
		return id, nil
	}

	if !plan.OriginGroupName.IsNull() && plan.OriginGroupName.ValueString() != "" {
		// Get folder ID (from resource or provider config)
		var diags diag.Diagnostics
		folderID := r.getFolderID(plan, &diags)
		if diags.HasError() {
			return 0, fmt.Errorf("folder_id is required but not set")
		}

		// List origin groups and find by name
		listReq := &cdn.ListOriginGroupsRequest{
			FolderId: folderID,
		}

		it := r.providerConfig.SDK.CDN().OriginGroup().OriginGroupIterator(ctx, listReq)
		for it.Next() {
			group := it.Value()
			if group.Name == plan.OriginGroupName.ValueString() {
				return group.Id, nil
			}
		}

		if err := it.Error(); err != nil {
			return 0, fmt.Errorf("failed to list origin groups: %w", err)
		}

		return 0, fmt.Errorf("origin group with name %q not found", plan.OriginGroupName.ValueString())
	}

	return 0, fmt.Errorf("either origin_group_id or origin_group_name must be specified")
}

// getFolderID returns folder ID from model or provider config
func (r *cdnResourceResource) getFolderID(model *CDNResourceModel, diags *diag.Diagnostics) string {
	if !model.FolderID.IsNull() && model.FolderID.ValueString() != "" {
		return model.FolderID.ValueString()
	}
	if r.providerConfig.ProviderState.FolderID.ValueString() != "" {
		return r.providerConfig.ProviderState.FolderID.ValueString()
	}
	diags.AddError("folder_id is required", "Please set folder_id in this resource or at provider level")
	return ""
}
