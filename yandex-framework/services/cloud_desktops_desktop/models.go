package cloud_desktops_desktop

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
)

type Desktop struct {
	Id types.String `tfsdk:"id"`

	DesktopId        types.String     `tfsdk:"desktop_id"`
	Name             types.String     `tfsdk:"name"`
	DesktopGroupId   types.String     `tfsdk:"desktop_group_id"`
	NetworkInterface NetworkInterface `tfsdk:"network_interface"`
	Users            types.List       `tfsdk:"members"`

	Labels   types.Map      `tfsdk:"labels"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type DesktopDataSource struct {
	DesktopId      types.String `tfsdk:"desktop_id"`
	Name           types.String `tfsdk:"name"`
	FolderId       types.String `tfsdk:"folder_id"`
	DesktopGroupId types.String `tfsdk:"desktop_group_id"`
	Users          types.List   `tfsdk:"members"`

	Labels types.Map `tfsdk:"labels"`
}

type NetworkInterface struct {
	SubnetId types.String `tfsdk:"subnet_id"`
}

type User struct {
	SubjectId   types.String `tfsdk:"subject_id"`
	SubjectType types.String `tfsdk:"subject_type"`
}

var userType = types.ObjectType{
	AttrTypes: map[string]attr.Type{
		"subject_id":   types.StringType,
		"subject_type": types.StringType,
	},
}

func desktopToState(ctx context.Context, desktop *clouddesktop.Desktop, state *Desktop) diag.Diagnostics {
	state.DesktopId = types.StringValue(desktop.Id)
	state.Name = types.StringValue(desktop.Name)
	state.DesktopGroupId = types.StringValue(desktop.DesktopGroupId)

	var diag diag.Diagnostics
	state.Labels, diag = types.MapValueFrom(ctx, types.StringType, desktop.Labels)
	if diag.HasError() {
		return diag
	}

	var userValues []attr.Value
	for _, user := range desktop.Users {
		// sometimes api doesn't return the type of the subject, it's pretty obvious that it's a user then
		subjectType := user.SubjectType
		if subjectType == "" {
			subjectType = "userAccount"
		}

		userValue, diagInt := types.ObjectValue(userType.AttrTypes, map[string]attr.Value{
			"subject_id":   types.StringValue(user.SubjectId),
			"subject_type": types.StringValue(subjectType),
		})

		userValues = append(userValues, userValue)
		diag.Append(diagInt...)
	}
	if diag.HasError() {
		return diag
	}

	state.Users, diag = types.ListValue(userType, userValues)
	return diag
}

func desktopToDataSourceState(ctx context.Context, desktop *clouddesktop.Desktop, state *DesktopDataSource) diag.Diagnostics {
	state.DesktopId = types.StringValue(desktop.Id)
	state.Name = types.StringValue(desktop.Name)
	state.FolderId = types.StringValue(desktop.FolderId)
	state.DesktopGroupId = types.StringValue(desktop.DesktopGroupId)

	var diag diag.Diagnostics
	state.Labels, diag = types.MapValueFrom(ctx, types.StringType, desktop.Labels)
	if diag.HasError() {
		return diag
	}

	var userValues []attr.Value
	for _, user := range desktop.Users {
		// sometimes api doesn't return the type of the subject, it's pretty obvious that it's a user then
		subjectType := user.SubjectType
		if subjectType == "" {
			subjectType = "userAccount"
		}

		userValue, diagInt := types.ObjectValue(userType.AttrTypes, map[string]attr.Value{
			"subject_id":   types.StringValue(user.SubjectId),
			"subject_type": types.StringValue(subjectType),
		})

		userValues = append(userValues, userValue)
		diag.Append(diagInt...)
	}
	if diag.HasError() {
		return diag
	}

	state.Users, diag = types.ListValue(userType, userValues)
	return diag
}

func planToCreateDesktop(ctx context.Context, plan *Desktop) (*clouddesktop.CreateDesktopRequest, diag.Diagnostics) {
	create := &clouddesktop.CreateDesktopRequest{}

	create.DesktopGroupId = plan.DesktopGroupId.ValueString()
	create.SubnetId = plan.NetworkInterface.SubnetId.ValueString()

	users := []User{}
	diag := plan.Users.ElementsAs(ctx, &users, false)
	if diag.HasError() {
		return nil, diag
	}

	for _, user := range users {
		create.Users = append(create.Users, &clouddesktop.User{
			SubjectId:   user.SubjectId.ValueString(),
			SubjectType: user.SubjectType.ValueString(),
		})
	}

	return create, diag
}

func planToUpdateDesktop(ctx context.Context, plan *Desktop) (*clouddesktop.UpdatePropertiesRequest, diag.Diagnostics) {
	update := &clouddesktop.UpdatePropertiesRequest{}

	update.Name = plan.Name.ValueString()

	update.Labels = make(map[string]string, 0)
	diag := plan.Labels.ElementsAs(ctx, &update.Labels, false)
	if diag.HasError() {
		return nil, diag
	}

	return update, diag
}
func ConstructID(desktopID, subnetID string) string {
	return strings.Join([]string{desktopID, subnetID}, ",")
}

func DeconstructID(ID string) (string, string, error) {
	vals := strings.Split(ID, ",")
	if len(vals) != 2 {
		return "", "", fmt.Errorf("Invalid resource id format: %q", ID)
	}

	return vals[0], vals[1], nil
}
