package mdb_mongodb_user

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/go-sdk/services/mdb/mongodb/v1"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"

	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
)

const (
	yandexMDBMongoDBUserCreateTimeout = time.Hour
	yandexMDBMongoDBUserDeleteTimeout = time.Hour
	yandexMDBMongoDBUserUpdateTimeout = 2 * time.Hour
)

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mongodb_user"
}

func (r *bindingResource) Configure(_ context.Context,
	req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *bindingResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MongoDB user within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/).",
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
				MarkdownDescription: "The ID of the cluster to which user belongs to.",
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
				MarkdownDescription: "The password of the user. Either `password` or `password_wo` is required for users with `PASSWORD` authentication and both must be omitted for users with `IAM` authentication.",
				Optional:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("password_wo")),
				},
			},
			"password_wo": schema.StringAttribute{
				MarkdownDescription: "The password of the user. This attribute is write-only and is not stored in state. Requires `password_wo_version` to trigger updates. Write-only arguments are supported in Terraform 1.11 and later. Must be omitted for users with `IAM` authentication.",
				Optional:            true,
				Sensitive:           true,
				WriteOnly:           true,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo_version")),
				},
			},
			"password_wo_version": schema.Int64Attribute{
				MarkdownDescription: "A version number for the write-only password. Increment this to trigger a password update.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRelative().AtParent().AtName("password_wo")),
				},
			},
			"auth_type": schema.StringAttribute{
				MarkdownDescription: "The authentication type of the user. Either `PASSWORD` (default) or `IAM`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(authTypePassword),
				Validators: []validator.String{
					stringvalidator.OneOf(authTypePassword, authTypeIam),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "Inhibits deletion of the user.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"permission": schema.SetNestedBlock{
				MarkdownDescription: "Set of permissions granted to the user.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"database_name": schema.StringAttribute{
							MarkdownDescription: "The name of the database that the permission grants access to.",
							Required:            true,
						},
						"roles": schema.SetAttribute{
							MarkdownDescription: "The roles of the user in this database. For more information see [the official documentation](https://yandex.cloud/docs/managed-mongodb/concepts/users-and-roles).",
							Optional:            true,
							Computed:            true,
							ElementType:         basetypes.StringType{},
							Default:             setdefault.StaticValue(types.SetValueMust(types.StringType, []attr.Value{})),
						},
					},
				},
			},
		},
	}
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state User
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	cid := state.ClusterID.ValueString()
	userName := state.Name.ValueString()
	user, err := mongodbsdk.NewUserClient(r.providerConfig.SDKv2).Get(ctx, &mongodb.GetUserRequest{
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
			"Error while requesting API to get MongoDB user:"+err.Error(),
		)
		return
	}
	resp.Diagnostics.Append(userToState(user, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Id = types.StringValue(resourceid.Construct(cid, userName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan User
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	var passwordWo types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("password_wo"), &passwordWo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBMongoDBUserCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	userPlan, diags := userFromState(ctx, &plan, mongodbUserPasswordForCreate(&plan, passwordWo))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createUser(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, cid, userPlan)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, userPlan.Name))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func getUpdatePaths(plan, state *mongodb.UserSpec, passwordChanged bool) []string {
	var updatePaths []string
	if passwordChanged {
		updatePaths = append(updatePaths, "password")
	}
	if fmt.Sprintf("%v", state.Permissions) != fmt.Sprintf("%v", plan.Permissions) {
		updatePaths = append(updatePaths, "permissions")
	}
	if plan.DeletionProtection != state.DeletionProtection {
		updatePaths = append(updatePaths, "deletion_protection")
	}
	return updatePaths
}

func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan User
	var state User
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	var passwordWo types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("password_wo"), &passwordWo)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBMongoDBUserUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	password, passwordChanged, passwordDiags := mongodbUserPasswordChange(&plan, &state, passwordWo)
	resp.Diagnostics.Append(passwordDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	userState, diags := userFromState(ctx, &state, state.Password.ValueString())
	resp.Diagnostics.Append(diags...)
	userPlan, diags := userFromState(ctx, &plan, password)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	updatePaths := getUpdatePaths(userPlan, userState, passwordChanged)

	if len(updatePaths) > 0 {
		updateUser(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, cid, userPlan, updatePaths)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, userPlan.Name))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state User
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBMongoDBUserDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	deleteUser(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, cid, dbName)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, userName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	user := readUser(ctx, r.providerConfig.SDKv2, &resp.Diagnostics, clusterId, userName)
	if resp.Diagnostics.HasError() {
		return
	}
	var state User
	resp.Diagnostics.Append(userToState(user, &state)...)

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"delete": types.StringType,
			"update": types.StringType,
		}),
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func mongodbUserPasswordForCreate(plan *User, passwordWo types.String) string {
	if !passwordWo.IsNull() && !passwordWo.IsUnknown() {
		return passwordWo.ValueString()
	}
	return plan.Password.ValueString()
}

func mongodbUserPasswordChange(plan, state *User, passwordWo types.String) (string, bool, diag.Diagnostics) {
	password := plan.Password.ValueString()
	passwordChanged := !plan.Password.IsNull() && !plan.Password.Equal(state.Password)

	if plan.PasswordWoVersion.IsNull() || plan.PasswordWoVersion.Equal(state.PasswordWoVersion) {
		return password, passwordChanged, nil
	}
	if passwordWo.IsNull() || passwordWo.IsUnknown() {
		diagnostics := diag.Diagnostics{}
		diagnostics.AddAttributeError(
			path.Root("password_wo"),
			"Missing MongoDB user password",
			"password_wo must be configured when password_wo_version changes",
		)
		return "", false, diagnostics
	}

	return passwordWo.ValueString(), true, nil
}
