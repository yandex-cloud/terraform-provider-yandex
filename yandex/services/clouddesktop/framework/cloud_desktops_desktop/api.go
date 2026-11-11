package cloud_desktops_desktop

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

func readDesktopByID(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktopID string) (*clouddesktop.Desktop, bool) {
	desktop, err := apisdk.NewDesktopClient(sdk).Get(ctx, &clouddesktop.GetDesktopRequest{
		DesktopId: desktopID,
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
			"Error while requesting API to get Desktop: "+err.Error(),
		)
		return nil, isNotFound
	}

	return desktop, false
}

func readDesktopByNameAndFolderID(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, name, folderID string) *clouddesktop.Desktop {
	desktops, err := apisdk.NewDesktopClient(sdk).List(ctx, &clouddesktop.ListDesktopsRequest{
		FolderId: folderID,
	})
	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to List Desktops with folderID: "+err.Error(),
		)
		return nil
	}

	for len(desktops.Desktops) != 0 {
		for _, desktop := range desktops.Desktops {
			if desktop.Name == name {
				return desktop
			}
		}

		if desktops.NextPageToken == "" {
			break
		}
		desktops, err = apisdk.NewDesktopClient(sdk).List(ctx, &clouddesktop.ListDesktopsRequest{
			FolderId:  folderID,
			PageToken: desktops.NextPageToken,
		})

		if err != nil {
			diag.AddError(
				"Failed to Read resource",
				"Error while requesting API to List Desktops with folderID: "+err.Error(),
			)
			return nil
		}
	}

	diag.AddError(
		"Failed to Read resource",
		"API didn't return Desktop with such name and folderID",
	)
	return nil
}

func createDesktop(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktop *clouddesktop.CreateDesktopRequest) string {
	op, err := apisdk.NewDesktopClient(sdk).Create(ctx, desktop)

	if err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while requesting API to create Desktop: "+err.Error(),
		)
		return ""
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Create resource",
			"Error while waiting for operation to create Desktop: "+err.Error(),
		)
		return ""
	}

	resp := op.Metadata()
	return resp.DesktopId
}

func updateDesktop(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktopID string, desktop *clouddesktop.UpdatePropertiesRequest, updatePaths []string) {
	desktop.DesktopId = desktopID
	mask, err := fieldmaskpb.New(desktop, updatePaths...)
	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Update paths to update Desktop are incorrect: "+err.Error(),
		)
		return
	}
	desktop.UpdateMask = mask

	op, err := apisdk.NewDesktopClient(sdk).UpdateProperties(ctx, desktop)

	if err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while requesting API to update Desktop: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Update resource",
			"Error while waiting for API to update Desktop: "+err.Error(),
		)
	}
}

func deleteDesktop(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, desktopID string) {
	op, err := apisdk.NewDesktopClient(sdk).Delete(ctx, &clouddesktop.DeleteDesktopRequest{
		DesktopId: desktopID,
	})

	if err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while requesting API to delete Desktop: "+err.Error(),
		)
		return
	}

	if _, err = op.Wait(ctx); err != nil {
		diag.AddError(
			"Failed to Delete resource",
			"Error while waiting for operation to delete Desktop: "+err.Error(),
		)
	}
}
