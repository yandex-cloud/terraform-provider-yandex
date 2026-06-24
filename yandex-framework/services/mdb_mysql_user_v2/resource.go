package mdb_mysql_user_v2

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	mysql "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	mysqlv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	yandexMDBMySQLUserDefaultTimeout = 10 * time.Minute
)

type userResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &userResource{}
}

func (r *userResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mysql_user_v2"
}

func (r *userResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerConfig, ok := req.ProviderData.(*provider_config.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf(
				"Expected *provider_config.Config, got: %T. Please report this issue to the provider developers.",
				req.ProviderData,
			),
		)
		return
	}
	r.providerConfig = providerConfig
}

func (r *userResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MySQL user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/ru/docs/managed-mysql/operations/cluster-users).",
		Attributes: map[string]schema.Attribute{
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
			"id": schema.StringAttribute{
				MarkdownDescription: common.ResourceDescriptions["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cluster_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the MySQL cluster.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the user.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password of the user.",
				Optional:            true,
				Sensitive:           true,
			},
			"generate_password": schema.BoolAttribute{
				MarkdownDescription: "Generate password using Connection Manager. Used only during creation.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"global_permissions": schema.SetAttribute{
				MarkdownDescription: "List of the user's global permissions.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
			},
			"authentication_plugin": schema.StringAttribute{
				MarkdownDescription: "Authentication plugin.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"MYSQL_NATIVE_PASSWORD",
						"CACHING_SHA2_PASSWORD",
						"SHA256_PASSWORD",
						"MYSQL_NO_LOGIN",
						"MDB_IAMPROXY_AUTH",
					),
				},
			},
			"connection_manager": schema.MapAttribute{
				MarkdownDescription: "Connection Manager connection configuration. Filled in by the server automatically.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"deletion_protection_mode": schema.StringAttribute{
				MarkdownDescription: "Deletion Protection inhibits deletion of the user.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("DELETION_PROTECTION_MODE_DISABLED"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"DELETION_PROTECTION_MODE_DISABLED",
						"DELETION_PROTECTION_MODE_ENABLED",
						"DELETION_PROTECTION_MODE_INHERITED",
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.SetNestedBlock{
				MarkdownDescription: "Set of permissions granted to the user.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"database_name": schema.StringAttribute{
							MarkdownDescription: "The name of the database.",
							Required:            true,
						},
						"roles": schema.ListAttribute{
							MarkdownDescription: "List of roles.",
							ElementType:         types.StringType,
							Optional:            true,
						},
					},
				},
			},
			"connection_limits": schema.ListNestedBlock{
				MarkdownDescription: "User's connection limits.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_questions_per_hour": schema.Int64Attribute{
							MarkdownDescription: "Max questions per hour.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(-1),
						},
						"max_updates_per_hour": schema.Int64Attribute{
							MarkdownDescription: "Max updates per hour.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(-1),
						},
						"max_connections_per_hour": schema.Int64Attribute{
							MarkdownDescription: "Max connections per hour.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(-1),
						},
						"max_user_connections": schema.Int64Attribute{
							MarkdownDescription: "Max user connections.",
							Optional:            true,
							Computed:            true,
							Default:             int64default.StaticInt64(-1),
						},
					},
				},
			},
		},
	}
}

func (r *userResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan User
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBMySQLUserDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	userSpec := stateToSpec(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	userSpec.DeletionProtectionMode = getDeletionProtectionModeValue(plan.DeletionProtection)

	CreateUser(ctx, r.providerConfig, &resp.Diagnostics, &mysql.CreateUserRequest{
		ClusterId: cid,
		UserSpec:  userSpec,
	})
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, plan.Name.ValueString()))

	user := ReadUser(ctx, r.providerConfig, &resp.Diagnostics, cid, plan.Name.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	specToState(ctx, user, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state User
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid, userName, err := resourceid.Deconstruct(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse resource ID",
			fmt.Sprintf("Error parsing resource ID %q: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	user, err := mysqlv1sdk.NewUserClient(r.providerConfig.SDKv2).Get(ctx, &mysql.GetUserRequest{
		ClusterId: cid,
		UserName:  userName,
	})
	if err != nil {
		f := resp.Diagnostics.AddError
		if validate.IsStatusWithCode(err, codes.NotFound) {
			resp.State.RemoveResource(ctx)
			f = resp.Diagnostics.AddWarning
		}
		f(
			"Failed to Read resource",
			fmt.Sprintf(
				"Error while requesting API to read MySQL user %q in cluster %q: %s",
				userName, cid, err.Error(),
			),
		)
		return
	}

	specToState(ctx, user, &state, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *userResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state User
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBMySQLUserDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	cid, userName, err := resourceid.Deconstruct(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse resource ID",
			fmt.Sprintf("Error parsing resource ID %q: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	userSpec := stateToSpec(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var updatePaths []string

	if !plan.Password.IsNull() && !plan.Password.Equal(state.Password) {
		updatePaths = append(updatePaths, "password")
	}
	if !plan.Permissions.Equal(state.Permissions) {
		updatePaths = append(updatePaths, "permissions")
	}
	if !plan.GlobalPermissions.Equal(state.GlobalPermissions) {
		updatePaths = append(updatePaths, "global_permissions")
	}
	if !plan.ConnectionLimits.Equal(state.ConnectionLimits) {
		updatePaths = append(updatePaths,
			"connection_limits.max_questions_per_hour",
			"connection_limits.max_updates_per_hour",
			"connection_limits.max_connections_per_hour",
			"connection_limits.max_user_connections",
		)
	}
	if !plan.AuthenticationPlugin.Equal(state.AuthenticationPlugin) {
		updatePaths = append(updatePaths, "authentication_plugin")
	}
	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		updatePaths = append(updatePaths, "deletion_protection_mode")
	}

	if len(updatePaths) > 0 {
		updateReq := &mysql.UpdateUserRequest{
			ClusterId:              cid,
			UserName:               userName,
			Password:               userSpec.Password,
			Permissions:            userSpec.Permissions,
			GlobalPermissions:      userSpec.GlobalPermissions,
			ConnectionLimits:       userSpec.ConnectionLimits,
			AuthenticationPlugin:   userSpec.AuthenticationPlugin,
			DeletionProtectionMode: getDeletionProtectionModeValue(plan.DeletionProtection),
			UpdateMask:             &fieldmaskpb.FieldMask{Paths: updatePaths},
		}

		UpdateUser(ctx, r.providerConfig, &resp.Diagnostics, updateReq)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	user := ReadUser(ctx, r.providerConfig, &resp.Diagnostics, cid, userName)
	if resp.Diagnostics.HasError() {
		return
	}

	specToState(ctx, user, &plan, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *userResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state User
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBMySQLUserDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid, userName, err := resourceid.Deconstruct(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse resource ID",
			fmt.Sprintf("Error parsing resource ID %q: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	DeleteUser(ctx, r.providerConfig, &resp.Diagnostics, cid, userName)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *userResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterID, userName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf(
				"Expected import identifier with format: <cluster_id>:<user_name>. Got: %q. Error: %s",
				req.ID, err.Error(),
			),
		)
		return
	}

	user := ReadUser(ctx, r.providerConfig, &resp.Diagnostics, clusterID, userName)
	if resp.Diagnostics.HasError() {
		return
	}

	var state User
	specToState(ctx, user, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Password = types.StringNull()
	state.GeneratePassword = types.BoolValue(false)

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
