package testfixtures

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CheckFolderDestroy(config func() (*providerconfig.Config, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cfg, err := config()
		if err != nil {
			return err
		}
		return checkFolderDestroy(resourcemanagersdk.NewFolderClient(cfg.SDK), s)
	}
}

type folderGetter interface {
	Get(context.Context, *resourcemanager.GetFolderRequest, ...grpc.CallOption) (*resourcemanager.Folder, error)
}

func checkFolderDestroy(client folderGetter, s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_resourcemanager_folder" {
			continue
		}
		_, err := client.Get(context.Background(), &resourcemanager.GetFolderRequest{
			FolderId: rs.Primary.ID,
		})
		if status.Code(err) == codes.NotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("check yandex_resourcemanager_folder %q is deleted: %w", rs.Primary.ID, err)
		}
		return fmt.Errorf("folder still exists: %s", rs.Primary.ID)
	}
	return nil
}
