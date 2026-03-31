package mdb_mysql_database_v2

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mysql/v1"
	mysqlv1sdk "github.com/yandex-cloud/go-sdk/services/mdb/mysql/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	yandexMDBMySQLDatabaseDefaultTimeout = 10 * time.Minute
)

type databaseResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &databaseResource{}
}

func (r *databaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mysql_database_v2"
}

func (r *databaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *databaseResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MySQL database within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/ru/docs/managed-mysql/operations/databases).",
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
				MarkdownDescription: "ID of the MySQL cluster. Provided by the client when the database is created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the database.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 63),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z0-9_-]*$`),
						"must contain only letters, numbers, underscores, and hyphens",
					),
				},
			},
			"deletion_protection_mode": schema.StringAttribute{
				MarkdownDescription: "Deletion Protection inhibits deletion of the database. Possible values: DELETION_PROTECTION_MODE_DISABLED (default), DELETION_PROTECTION_MODE_ENABLED, DELETION_PROTECTION_MODE_INHERITED.",
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
	}
}

func (r *databaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBMySQLDatabaseDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()

	var spec mysql.DatabaseSpec
	stateToSpec(&plan, &spec)

	createReq := &mysql.CreateDatabaseRequest{
		ClusterId:    cid,
		DatabaseSpec: &spec,
	}

	CreateDatabase(ctx, r.providerConfig, &resp.Diagnostics, createReq)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, plan.Name.ValueString()))

	db := ReadDatabase(ctx, r.providerConfig, &resp.Diagnostics, cid, plan.Name.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	specToState(db, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Database
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid, dbName, err := resourceid.Deconstruct(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse resource ID",
			fmt.Sprintf("Error parsing resource ID %q: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	db, err := mysqlv1sdk.NewDatabaseClient(r.providerConfig.SDKv2).Get(ctx, &mysql.GetDatabaseRequest{
		ClusterId:    cid,
		DatabaseName: dbName,
	})

	if err != nil {
		f := resp.Diagnostics.AddError
		if validate.IsStatusWithCode(err, codes.NotFound) {
			resp.State.RemoveResource(ctx)
			f = resp.Diagnostics.AddWarning
		}

		f(
			"Failed to Read resource",
			fmt.Sprintf("Error while requesting API to read MySQL database %q in cluster %q: %s", dbName, cid, err.Error()),
		)
		return
	}

	specToState(db, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *databaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Database
	var state Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBMySQLDatabaseDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	cid, dbName, err := resourceid.Deconstruct(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse resource ID",
			fmt.Sprintf("Error parsing resource ID %q: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	updateReq := &mysql.UpdateDatabaseRequest{
		ClusterId:    cid,
		DatabaseName: dbName,
	}

	var updatePaths []string

	if !plan.DeletionProtectionMode.Equal(state.DeletionProtectionMode) {
		updateReq.DeletionProtectionMode = getDeletionProtectionModeValue(plan.DeletionProtectionMode)
		updatePaths = append(updatePaths, "deletion_protection_mode")
	}

	if len(updatePaths) > 0 {
		updateReq.UpdateMask = &fieldmaskpb.FieldMask{
			Paths: updatePaths,
		}

		UpdateDatabase(ctx, r.providerConfig, &resp.Diagnostics, updateReq)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	db := ReadDatabase(ctx, r.providerConfig, &resp.Diagnostics, cid, dbName)
	if resp.Diagnostics.HasError() {
		return
	}

	specToState(db, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *databaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Database
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBMySQLDatabaseDefaultTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid, dbName, err := resourceid.Deconstruct(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse resource ID",
			fmt.Sprintf("Error parsing resource ID %q: %s", state.Id.ValueString(), err.Error()),
		)
		return
	}

	DeleteDatabase(ctx, r.providerConfig, &resp.Diagnostics, cid, dbName)
}

func (r *databaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, dbName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: <cluster_id>:<database_name>. Got: %q. Error: %s", req.ID, err.Error()),
		)
		return
	}

	db := ReadDatabase(ctx, r.providerConfig, &resp.Diagnostics, clusterId, dbName)
	if resp.Diagnostics.HasError() {
		return
	}

	var state Database
	specToState(db, &state)

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
