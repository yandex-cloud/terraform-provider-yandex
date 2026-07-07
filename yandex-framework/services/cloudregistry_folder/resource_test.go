package cloudregistry_folder

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cloudregistry "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	cloudregistryv1sdk "github.com/yandex-cloud/go-sdk/services/cloudregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	cloudRegistryFolderResource = "yandex_cloudregistry_folder.test-folder"
	cloudRegistryFolderPath     = "/common-artifacts/some-folder"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccCloudRegistryFolder_basic(t *testing.T) {
	registryName := acctest.RandomWithPrefix("tf-registry")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckCloudRegistryFolderDestroy,
			testAccCheckCloudRegistryDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistryFolderConfig(registryName, cloudRegistryFolderPath),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryFolderExists(cloudRegistryFolderResource),
					resource.TestCheckResourceAttrSet(cloudRegistryFolderResource, "id"),
					resource.TestCheckResourceAttrSet(cloudRegistryFolderResource, "artifact_id"),
					resource.TestCheckResourceAttrSet(cloudRegistryFolderResource, "registry_id"),
					resource.TestCheckResourceAttr(cloudRegistryFolderResource, "path", cloudRegistryFolderPath),
					resource.TestCheckResourceAttrSet(cloudRegistryFolderResource, "name"),
					resource.TestCheckResourceAttrSet(cloudRegistryFolderResource, "status"),
				),
			},
			{
				ResourceName:                         cloudRegistryFolderResource,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "artifact_id",
				ImportStateVerifyIgnore:              []string{"path", "registry_id", "with_history"},
			},
		},
	})
}

func testAccCloudRegistryFolderConfig(registryName, path string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "test-registry" {
  name = "%s"
  kind = "DOCKER"
  type = "LOCAL"
}

resource "yandex_cloudregistry_folder" "test-folder" {
  registry_id = yandex_cloudregistry_registry.test-registry.id
  path        = "%s"
}
`, registryName, path)
}

func testAccCheckCloudRegistryFolderExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := cloudregistryv1sdk.NewArtifactClient(config.SDKv2).Get(context.Background(), &cloudregistry.GetArtifactRequest{
			ArtifactId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.GetId() != rs.Primary.ID {
			return fmt.Errorf("Cloud Registry folder %s not found", n)
		}

		foundByPath, err := cloudregistryv1sdk.NewArtifactClient(config.SDKv2).GetByPath(context.Background(), &cloudregistry.GetArtifactByPathRequest{
			RegistryId: rs.Primary.Attributes["registry_id"],
			Path:       rs.Primary.Attributes["path"],
		})
		if err != nil {
			return err
		}

		if foundByPath.GetId() != rs.Primary.ID {
			return fmt.Errorf("Cloud Registry folder %s not found by path", n)
		}

		return nil
	}
}

func testAccCheckCloudRegistryFolderDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cloudregistry_folder" {
			continue
		}

		_, err := cloudregistryv1sdk.NewArtifactClient(config.SDKv2).Get(context.Background(), &cloudregistry.GetArtifactRequest{
			ArtifactId: rs.Primary.ID,
		})
		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				continue
			} else if ok {
				return fmt.Errorf("error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}

			return fmt.Errorf("Cloud Registry folder still exists")
		}

		return fmt.Errorf("Cloud Registry folder still exists")
	}

	return nil
}

func testAccCheckCloudRegistryDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_cloudregistry_registry" {
			continue
		}

		_, err := config.SDK.CloudRegistry().Registry().Get(context.Background(), &cloudregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				continue
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Cloud Registry still exists")
		}
	}

	return nil
}
