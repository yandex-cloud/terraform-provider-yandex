package testfixtures

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	computesdk "github.com/yandex-cloud/go-sdk/services/compute/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CheckImageDestroy(config *providerconfig.Config, s *terraform.State) error {
	return checkImageDestroy(computesdk.NewImageClient(config.SDK), s)
}

type imageGetter interface {
	Get(context.Context, *compute.GetImageRequest, ...grpc.CallOption) (*compute.Image, error)
}

func checkImageDestroy(client imageGetter, s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_compute_image" {
			continue
		}
		found, err := client.Get(context.Background(), &compute.GetImageRequest{
			ImageId: rs.Primary.ID,
		})
		if status.Code(err) == codes.NotFound {
			continue
		}
		if err != nil {
			return fmt.Errorf("check yandex_compute_image %q is deleted: %w", rs.Primary.ID, err)
		}
		// Public images are shared fixtures and must not be deleted by acceptance tests.
		if found.FolderId == "standard-images" {
			continue
		}
		return fmt.Errorf("image still exists: %s", rs.Primary.ID)
	}
	return nil
}
