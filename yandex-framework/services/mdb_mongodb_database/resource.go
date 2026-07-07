package mdb_mongodb_database

import (
	"context"
	"fmt"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/mongodb/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/resourceid"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/wrappers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

const (
	yandexMDBMongoDBDatabaseCreateTimeout = time.Hour
	yandexMDBMongoDBDatabaseUpdateTimeout = time.Hour
	yandexMDBMongoDBDatabaseDeleteTimeout = time.Hour
)

type bindingResource struct {
	providerConfig *provider_config.Config
}

func NewResource() resource.Resource {
	return &bindingResource{}
}

func (r *bindingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mdb_mongodb_database"
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
		MarkdownDescription: "Manages a MongoDB Database within the Yandex Cloud. For more information, see [the official documentation](https://yandex.cloud/docs/managed-mongodb/).",
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
				MarkdownDescription: "The ID of MongoDB Cluster.",
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
			},
			"deletion_protection": schema.BoolAttribute{
				MarkdownDescription: "Inhibits deletion of the database.",
				Optional:            true,
			},
		},
	}
}

func (r *bindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Database
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	db, err := r.providerConfig.SDK.MDB().MongoDB().Database().Get(ctx, &mongodb.GetDatabaseRequest{
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
			"Error while requesting API to get MongoDB database:"+err.Error(),
		)
		return
	}
	databaseToState(db, &state)

	state.Id = types.StringValue(resourceid.Construct(cid, dbName))
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Database
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, yandexMDBMongoDBDatabaseCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	dbName := plan.Name.ValueString()
	spec := &mongodb.DatabaseSpec{
		Name:               dbName,
		DeletionProtection: wrappers.BoolFromTF(plan.DeletionProtection),
	}
	createDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, spec)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, dbName))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Update changes mutable fields of the database (currently only deletion_protection;
// cluster_id and name require replacement).
func (r *bindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan Database
	var state Database
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, yandexMDBMongoDBDatabaseUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	cid := plan.ClusterID.ValueString()
	dbName := plan.Name.ValueString()

	if !plan.DeletionProtection.Equal(state.DeletionProtection) {
		updateDatabase(ctx, r.providerConfig, &resp.Diagnostics, &mongodb.UpdateDatabaseRequest{
			ClusterId:          cid,
			DatabaseName:       dbName,
			DeletionProtection: wrappers.BoolFromTF(plan.DeletionProtection),
			UpdateMask:         &fieldmaskpb.FieldMask{Paths: []string{"deletion_protection"}},
		})
		if resp.Diagnostics.HasError() {
			return
		}
	}

	plan.Id = types.StringValue(resourceid.Construct(cid, dbName))
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *bindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Database
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, yandexMDBMongoDBDatabaseDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	cid := state.ClusterID.ValueString()
	dbName := state.Name.ValueString()
	deleteDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, cid, dbName)
}

func (r *bindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterId, dbName, err := resourceid.Deconstruct(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			err.Error(),
		)
		return
	}
	db := readDatabase(ctx, r.providerConfig.SDK, &resp.Diagnostics, clusterId, dbName)
	if resp.Diagnostics.HasError() {
		return
	}
	var state Database
	databaseToState(db, &state)
	state.Id = types.StringValue(resourceid.Construct(clusterId, dbName))

	state.Timeouts = timeouts.Value{
		Object: types.ObjectNull(map[string]attr.Type{
			"create": types.StringType,
			"update": types.StringType,
			"delete": types.StringType,
		}),
	}

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}
