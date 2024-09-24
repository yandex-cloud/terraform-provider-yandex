package opensearch

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	ycsdk "github.com/yandex-cloud/go-sdk"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/legacy"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/log"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/model"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/request"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/request/auth"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/request/cluster"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/request/nodegroups"
	common_schema "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb/opensearch/validate"
)

const (
	yandexMDBOpenSearchClusterCreateTimeout = 30 * time.Minute
	yandexMDBOpenSearchClusterDeleteTimeout = 15 * time.Minute
	yandexMDBOpenSearchClusterUpdateTimeout = 60 * time.Minute
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &openSearchClusterResource{}
var _ resource.ResourceWithImportState = &openSearchClusterResource{}
var _ resource.ResourceWithUpgradeState = &openSearchClusterResource{}

func NewResource() resource.Resource {
	return &openSearchClusterResource{}
}

type openSearchClusterResource struct {
	providerConfig *provider_config.Config
}

// Metadata implements resource.Resource.
func (o *openSearchClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_opensearch_cluster"
}

// Configure implements resource.Resource.
func (o *openSearchClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	o.providerConfig = providerConfig
}

// ImportState implements resource.ResourceWithImportState.
func (o *openSearchClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (o *openSearchClusterResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		// State upgrade implementation from 0 to 2 (Schema.Version)
		0: legacy.NewUpgraderFromV0(ctx),
		// State upgrade implementation from 1 (prior state version) to 2 (Schema.Version)
		1: legacy.NewUpgraderFromV1(ctx),
	}
}

// Create implements resource.Resource.
func (o *openSearchClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan model.OpenSearch
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating OpenSearch Cluster")

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBOpenSearchClusterCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	clusterCreateRequest, diags := cluster.PrepareCreateRequest(ctx, &plan, &o.providerConfig.ProviderState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Creating OpenSearch Cluster request: %+v", clusterCreateRequest))

	clusterID := request.CreateCluster(ctx, o.providerConfig.SDK, &resp.Diagnostics, clusterCreateRequest)
	if resp.Diagnostics.HasError() {
		return
	}

	planAuthSettings, d := model.AuthSettingsFromState(ctx, plan.AuthSettings)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	authSettingReq, d := auth.PrepareUpdateRequest(ctx, clusterID, planAuthSettings)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("UpdateAuthSettings request: %+v", authSettingReq))

	request.UpdateAuthSettings(ctx, o.providerConfig.SDK, &resp.Diagnostics, authSettingReq)
	if resp.Diagnostics.HasError() {
		return
	}

	//TODO: check maybe we need to getClusterById and store result to state?
	plan.ID = types.StringValue(clusterID)

	updateState(ctx, o.providerConfig.SDK, &plan, &resp.Diagnostics, false)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished creating OpenSearch Cluster", log.IdFromModel(&plan))
}

// Delete implements resource.Resource.
func (o *openSearchClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state model.OpenSearch
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "Deleting OpenSearch Cluster", log.IdFromModel(&state))

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBOpenSearchClusterDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	request.DeleteCluster(ctx, o.providerConfig.SDK, &resp.Diagnostics, state.ID.ValueString())

	state.ID = types.StringUnknown()
	tflog.Debug(ctx, "Finished deleting OpenSearch Cluster", log.IdFromModel(&state))
}

// Read implements resource.Resource.
func (o *openSearchClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state model.OpenSearch
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateState(ctx, o.providerConfig.SDK, &state, &resp.Diagnostics, true)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	tflog.Debug(ctx, "Finished reading OpenSearch Cluster", log.IdFromModel(&state))
}

// Update implements resource.Resource.
func (o *openSearchClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan model.OpenSearch
	var state model.OpenSearch
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating OpenSearch Cluster", log.IdFromModel(&plan))

	updateTimeout, diags := state.Timeouts.Update(ctx, yandexMDBOpenSearchClusterUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	tflog.Debug(ctx, fmt.Sprintf("UpdateOpenSearch Cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("UpdateOpenSearch Cluster plan: %+v", plan))

	updateReq, d := cluster.PrepareUpdateParamsRequest(ctx, &state, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("UpdateOpenSearch Cluster request: %+v", updateReq))
	request.UpdateClusterSpec(ctx, o.providerConfig.SDK, &resp.Diagnostics, updateReq)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.AuthSettings.Equal(state.AuthSettings) {
		planAuthSettings, d := model.AuthSettingsFromState(ctx, plan.AuthSettings)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		authSettingReq, d := auth.PrepareUpdateRequest(ctx, state.ID.ValueString(), planAuthSettings)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}

		tflog.Debug(ctx, fmt.Sprintf("UpdateAuthSettings request: %+v", authSettingReq))

		request.UpdateAuthSettings(ctx, o.providerConfig.SDK, &resp.Diagnostics, authSettingReq)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if plan.Config.Equal(state.Config) {
		tflog.Debug(ctx, "No changes in Config section. Finishing updating OpenSearch Cluster", log.IdFromModel(&plan))
		updateState(ctx, o.providerConfig.SDK, &plan, &resp.Diagnostics, false)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	planConfig, stateConfig, d := model.ParseGenerics(ctx, &plan, &state, model.ParseConfig)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(o.processOpenSearchNodeGroupsUpdate(ctx, plan.ID.ValueString(), planConfig, stateConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(o.processDashboardsNodeGroupsUpdate(ctx, plan.ID.ValueString(), planConfig, stateConfig)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateState(ctx, o.providerConfig.SDK, &plan, &resp.Diagnostics, false)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	tflog.Debug(ctx, "Finishing updating OpenSearch Cluster", log.IdFromModel(&plan))
}

func (o *openSearchClusterResource) processOpenSearchNodeGroupsUpdate(ctx context.Context, cid string, planConfig, stateConfig *model.Config) diag.Diagnostics {
	if planConfig.OpenSearch.Equal(stateConfig.OpenSearch) {
		tflog.Debug(ctx, "No changes in config.opensearch", log.IdFromStr(cid))
		return nil
	}

	planOpenSearchBlock, stateOpenSearchBlock, diags := model.ParseGenerics(ctx, planConfig, stateConfig, model.ParseOpenSearchSubConfig)
	if diags.HasError() {
		return diags
	}

	if planOpenSearchBlock.NodeGroups.Equal(stateOpenSearchBlock.NodeGroups) {
		tflog.Debug(ctx, "No changes in config.opensearch.node_groups", log.IdFromStr(cid))
		return nil
	}

	planNodeGroups, stateNodeGroups, diags := model.ParseGenerics(ctx, planOpenSearchBlock, stateOpenSearchBlock, nodegroups.PrepareOpenSearchCreate)
	if diags.HasError() {
		return diags
	}

	//Create new nodegroups
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareAddOpenSearchRequests, request.AddOpenSearchNodeGroup)
	if diags.HasError() {
		return diags
	}

	// -------

	//Update existing nodegroups
	//to proper update managers count we should use the following sequence:
	//1) Increase hostcount in dedicated manager group if exists
	//2) decrease hostcount in mixed data/manager groups
	//3) do all other operations, including deleting of a group(s)
	//4) decrease hostcount in dedicated manager group if exists

	//TODO: maybe we should separate changing hostcount from other operations?

	//1) increase managers count
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareManagersToIncreaseRequests, request.UpdateOpenSearchNodeGroup)
	if diags.HasError() {
		return diags
	}

	//2) decrease data/managers host count
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareDataManagersToDecreaseRequests, request.UpdateOpenSearchNodeGroup)
	if diags.HasError() {
		return diags
	}

	//3) all other activities
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareOtherGroupsToUpdateRequests, request.UpdateOpenSearchNodeGroup)
	if diags.HasError() {
		return diags
	}

	// Delete old nodegroups
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareDeleteOpenSearchRequests, request.DeleteOpenSearchNodeGroup)
	if diags.HasError() {
		return diags
	}

	//4) decrease host count in managers group
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareManagersToDecreaseRequests, request.UpdateOpenSearchNodeGroup)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (o *openSearchClusterResource) processDashboardsNodeGroupsUpdate(ctx context.Context, cid string, planConfig, stateConfig *model.Config) diag.Diagnostics {
	if planConfig.Dashboards.Equal(stateConfig.Dashboards) {
		tflog.Debug(ctx, "No changes in config.dashboards", log.IdFromStr(cid))
		return nil
	}

	planDashboardsBlock, stateDashboardsBlock, diags := model.ParseGenerics(ctx, planConfig, stateConfig, model.ParseDashboardSubConfig)
	if diags.HasError() {
		return diags
	}

	if stateDashboardsBlock != nil && planDashboardsBlock.NodeGroups.Equal(stateDashboardsBlock.NodeGroups) {
		tflog.Debug(ctx, "No changes in config.dashboards.node_groups", log.IdFromStr(cid))
		return nil
	}

	planNodeGroups, stateNodeGroups, diags := model.ParseGenerics(ctx, planDashboardsBlock, stateDashboardsBlock, nodegroups.PrepareDashboardsCreate)
	if diags.HasError() {
		return diags
	}

	//Create new nodegroups
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareAddDashboardsRequests, request.AddDashboardsNodeGroup)
	if diags.HasError() {
		return diags
	}

	//Update existing nodegroups
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareUpdateDashboardsRequests, request.UpdateDashboardsNodeGroup)
	if diags.HasError() {
		return diags
	}

	//Update existing nodegroups network settings
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareUpdateDashboardsZoneAndSubnetIdsRequests, request.UpdateDashboardsNodeGroup)
	if diags.HasError() {
		return diags
	}

	//Delete old nodegroups
	diags = request.PrepareAndExecute(ctx, o.providerConfig.SDK, cid, planNodeGroups, stateNodeGroups,
		nodegroups.PrepareDeleteDashboardsRequests, request.DeleteDashboardsNodeGroup)

	return diags
}

// Schema implements resource.Resource.
func (o *openSearchClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	tflog.Info(ctx, "Initializing opensearch data source schema")
	resp.Schema = schema.Schema{
		Version: 2,
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"config": schema.SingleNestedBlock{
				Description: "Configuration of the OpenSearch cluster.",
				Blocks: map[string]schema.Block{
					"opensearch": schema.SingleNestedBlock{
						Validators: []validator.Object{
							objectvalidator.IsRequired(),
						},
						Attributes: map[string]schema.Attribute{
							"plugins": schema.SetAttribute{
								Computed:    true,
								Optional:    true,
								ElementType: types.StringType,
							},
						},
						Blocks: map[string]schema.Block{
							//NOTE: changed "set" to "list+customValidator" because https://github.com/hashicorp/terraform-plugin-sdk/issues/1210
							"node_groups": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.IsRequired(),
									listvalidator.SizeAtLeast(1),
									validate.UniqueByField("name", func(x model.OpenSearchNode) string { return x.Name.ValueString() }),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"resources": common_schema.NodeResource(),
									},
									Attributes: map[string]schema.Attribute{
										"name":        schema.StringAttribute{Required: true},
										"hosts_count": schema.Int64Attribute{Required: true},
										"zone_ids": schema.SetAttribute{
											Required:    true,
											ElementType: types.StringType,
										},
										"subnet_ids": schema.ListAttribute{
											Optional:    true,
											Computed:    true,
											ElementType: types.StringType,
										},

										"assign_public_ip": schema.BoolAttribute{Computed: true, Optional: true},
										"roles": schema.SetAttribute{
											Required:    true,
											ElementType: types.StringType,
											Validators: []validator.Set{
												validate.UniqueCaseInsensitive(),
											},
										},
									},
								},
							},
						},
					},
					"dashboards": schema.SingleNestedBlock{
						Validators: []validator.Object{
							objectvalidator.AlsoRequires(
								path.MatchRoot("config").AtName("dashboards").AtName("node_groups"),
							),
						},
						Blocks: map[string]schema.Block{
							//NOTE: changed "set" to "list+customValidator" because https://github.com/hashicorp/terraform-plugin-sdk/issues/1210
							"node_groups": schema.ListNestedBlock{
								Validators: []validator.List{
									listvalidator.SizeAtLeast(1),
									validate.UniqueByField("name", func(x model.DashboardNode) string { return x.Name.ValueString() }),
								},
								NestedObject: schema.NestedBlockObject{
									Blocks: map[string]schema.Block{
										"resources": common_schema.NodeResource(),
									},
									Attributes: map[string]schema.Attribute{
										"name":        schema.StringAttribute{Required: true},
										"hosts_count": schema.Int64Attribute{Required: true},
										"zone_ids": schema.SetAttribute{
											Required:    true,
											ElementType: types.StringType,
										},
										"subnet_ids": schema.ListAttribute{
											Optional:    true,
											Computed:    true,
											ElementType: types.StringType,
										},

										"assign_public_ip": schema.BoolAttribute{Computed: true, Optional: true},
									},
								},
							},
						},
					},
					"access": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"data_transfer": schema.BoolAttribute{Optional: true},
							"serverless":    schema.BoolAttribute{Optional: true},
						},
					},
				},
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{Computed: true, Optional: true},
					"admin_password": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"maintenance_window": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf("ANYTIME", "WEEKLY"),
						},
					},
					"day": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf("MON", "TUE", "WED", "THU", "FRI", "SAT", "SUN"),
						},
					},
					"hour": schema.Int64Attribute{
						Optional: true,
						Validators: []validator.Int64{
							int64validator.Between(1, 24),
						},
					},
				},
			},
		},
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"folder_id": schema.StringAttribute{
				Computed:    true,
				Optional:    true,
				Description: "ID of the folder that the OpenSearch cluster belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name":        schema.StringAttribute{Required: true, Description: "Name of the OpenSearch cluster. The name must be unique within the folder."},
			"description": schema.StringAttribute{Optional: true, Description: "Description of the OpenSearch cluster"},
			"labels": schema.MapAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Custom labels for the OpenSearch cluster as `key:value` pairs.",
			},
			"environment": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Deployment environment of the OpenSearch cluster.",
			},
			"hosts": common_schema.Hosts(),
			"network_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "ID of the network that the cluster belongs to.",
			},
			"health": schema.StringAttribute{Computed: true, Description: "Aggregated cluster health."},
			"status": schema.StringAttribute{Computed: true, Description: "Current state of the cluster."},
			"security_group_ids": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "User security groups",
			},
			"service_account_id":  schema.StringAttribute{Optional: true},
			"deletion_protection": schema.BoolAttribute{Computed: true, Optional: true},
			"auth_settings": schema.SingleNestedAttribute{
				Description: "Authentification settings for dashboards",
				Optional:    true,
				Validators: []validator.Object{
					objectvalidator.AlsoRequires(
						path.MatchRoot("config").AtName("dashboards"),
						path.MatchRoot("auth_settings").AtName("saml"),
					),
				},
				Attributes: map[string]schema.Attribute{
					"saml": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"enabled":                   schema.BoolAttribute{Required: true},
							"idp_entity_id":             schema.StringAttribute{Required: true},
							"idp_metadata_file_content": schema.StringAttribute{Required: true},
							"sp_entity_id":              schema.StringAttribute{Required: true},
							"dashboards_url":            schema.StringAttribute{Required: true},
							"roles_key":                 schema.StringAttribute{Optional: true},
							"subject_key":               schema.StringAttribute{Optional: true},
						},
					},
				},
			},
		},
	}
}

func updateState(ctx context.Context, sdk *ycsdk.SDK, state *model.OpenSearch, diagnostics *diag.Diagnostics, createIfMissing bool) {
	clusterID := state.ID.ValueString()
	tflog.Debug(ctx, "Reading OpenSearch Cluster", log.IdFromStr(clusterID))
	cluster := request.GetCusterByID(ctx, sdk, diagnostics, clusterID)
	if diagnostics.HasError() {
		return
	}

	if cluster == nil {
		if createIfMissing {
			// To create a new cluster if missing
			state.ID = types.StringUnknown()
			return
		}

		diagnostics.AddError(
			"Failed to get cluster data",
			fmt.Sprintf("Cluster with id %s not found", clusterID))
		return
	}

	state.ClusterID = state.ID

	tflog.Debug(ctx, fmt.Sprintf("updateState: OpenSearch Cluster state: %+v", state))
	tflog.Debug(ctx, fmt.Sprintf("updateState: Received OpenSearch Cluster data: %+v", cluster))

	diags := model.ClusterToState(ctx, cluster, state)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	authSettings := request.GetAuthSettings(ctx, sdk, diagnostics, clusterID)
	if diagnostics.HasError() {
		return
	}

	state.AuthSettings, diags = model.AuthSettingsToState(ctx, authSettings, state.AuthSettings)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	hosts := request.GetHostsList(ctx, sdk, diagnostics, clusterID)
	if diagnostics.HasError() {
		return
	}

	state.Hosts, diags = model.HostsToState(ctx, hosts)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("updatedState: OpenSearch Cluster state: %+v", state))
}
