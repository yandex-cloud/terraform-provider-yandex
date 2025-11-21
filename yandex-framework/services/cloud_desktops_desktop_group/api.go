package cloud_desktops_desktop_group

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
	apisdk "github.com/yandex-cloud/go-sdk/services/clouddesktop/v1/api"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func readDesktopGroupByID(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktopGroupID string) (*clouddesktop.DesktopGroup, bool) {
	desktopGroup, err := apisdk.NewDesktopGroupClient(sdk).Get(ctx, &clouddesktop.GetDesktopGroupRequest{
		DesktopGroupId: desktopGroupID,
	})
	if err != nil {
		isNotFound := false
		f := diag.AddError
		if validate.IsStatusWithCode(err, codes.NotFound) {
			f = diag.AddWarning
			isNotFound = true
		}

		f(
			"Failed to Read resource",
			"Error while requesting API to get Desktop Group: "+err.Error(),
		)
		return nil, isNotFound
	}

	return desktopGroup, false
}

func readDesktopGroupByNameAndFolderID(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, name, folderID string) (*clouddesktop.DesktopGroup, bool, error) {
	desktopGroups, err := apisdk.NewDesktopGroupClient(sdk).List(ctx, &clouddesktop.ListDesktopGroupsRequest{
		FolderId: folderID,
	})
	if err != nil {
		isNotFound := false
		f := diag.AddError
		if validate.IsStatusWithCode(err, codes.NotFound) {
			isNotFound = true
			f = diag.AddWarning
		}

		f(
			"Failed to Read resource",
			"Error while requesting API to List Desktop Groups: "+err.Error(),
		)
		return nil, isNotFound, err
	}

	for len(desktopGroups.DesktopGroups) != 0 {
		for _, group := range desktopGroups.DesktopGroups {
			if group.Name == name {
				return group, false, nil
			}
		}

		if desktopGroups.NextPageToken == "" {
			break
		}
		desktopGroups, err = apisdk.NewDesktopGroupClient(sdk).List(ctx, &clouddesktop.ListDesktopGroupsRequest{
			FolderId:  folderID,
			PageToken: desktopGroups.NextPageToken,
		})

		if err != nil {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to List Desktop Groups: "+err.Error(),
			)
			return nil, false, err
		}
	}

	diag.AddError(
		"Failed to Read resource",
		"API didn't return Desktop Group with such name and folderID",
	)
	return nil, true, err
}

func createDesktopGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktopGroup *clouddesktop.CreateDesktopGroupRequest) string {
	op, err := apisdk.NewDesktopGroupClient(sdk).Create(ctx, desktopGroup)

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create Desktop Group: "+err.Error(),
		)
		return ""
	}

	if _, err := op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create Desktop Group: "+err.Error(),
		)
		return ""
	}

	resp := op.Metadata()
	return resp.DesktopGroupId
}

func updateDesktopGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktopGroupID string, desktopGroup *clouddesktop.UpdateDesktopGroupRequest, updatePaths []string) {
	desktopGroup.DesktopGroupId = desktopGroupID
	mask, err := fieldmaskpb.New(desktopGroup, updatePaths...)
	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Update paths to update Desktop Group are incorrect: "+err.Error(),
		)
		return
	}
	desktopGroup.UpdateMask = mask

	op, err := apisdk.NewDesktopGroupClient(sdk).Update(ctx, desktopGroup)

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update Desktop Group: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for API to update Desktop Group: "+err.Error(),
		)
	}
}

func deleteDesktopGroup(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktopGroupID string) {
	op, err := apisdk.NewDesktopGroupClient(sdk).Delete(ctx, &clouddesktop.DeleteDesktopGroupRequest{
		DesktopGroupId: desktopGroupID,
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete Desktop Group: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Desktop Group: "+err.Error(),
		)
	}
}
