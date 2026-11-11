package yandex_cloud_desktops_image

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/clouddesktop/v1"
	apisdk "github.com/yandex-cloud/go-sdk/services/clouddesktop/v1/api"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
)

func readImageByNameAndFolderID(ctx context.Context, sdk *ycsdk.SDK, diag *diag.Diagnostics, name, folderID string) *clouddesktop.DesktopImage {
	images, err := apisdk.NewDesktopImageClient(sdk).List(ctx, &clouddesktop.ListDesktopImagesRequest{
		FolderId: folderID,
		Filter:   fmt.Sprintf("name=\"%s\"", name),
		PageSize: 1,
	})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to list Desktop Image: "+err.Error(),
		)
		return nil
	}

	if len(images.DesktopImages) == 0 {
		diag.AddError(
			"Failed to Read resource",
			"No Desktop Image with such FolderID and Name found",
		)
		return nil
	}

	// have to do this to get labels
	id := images.DesktopImages[0].Id
	res, err := apisdk.NewDesktopImageClient(sdk).Get(ctx, &clouddesktop.GetDesktopImageRequest{ImageId: id})

	if err != nil {
		diag.AddError(
			"Failed to Read resource",
			"Error while requesting API to get Desktop Image: "+err.Error(),
		)
		return nil
	}
	return res
}
