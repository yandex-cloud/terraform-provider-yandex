package kubernetes_marketplace_helm_release_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	k8s_marketplace "github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/marketplace/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/k8s/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

var defaultZone = "ru-central1-d"

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccK8sMarketplaceHelmRelease_basic(t *testing.T) {
	appName := "gatekeeper"
	versionID := "f2ecif2vt62k2637tgus" // Gatekeeper 3.12.0
	helmReleaseResourceFullName := "yandex_kubernetes_marketplace_helm_release.example"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {Source: "hashicorp/time"},
		},
		CheckDestroy: testAccCheckHelmReleaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccK8sMarketplaceHelmRelease_basic(appName, versionID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHelmReleaseExists(helmReleaseResourceFullName),
				),
			},
			{
				ResourceName:      helmReleaseResourceFullName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckHelmReleaseExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.KubernetesMarketplace().HelmRelease().Get(context.Background(), &k8s_marketplace.GetHelmReleaseRequest{
			Id: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.GetId() != rs.Primary.ID {
			return fmt.Errorf("Helm Release %s not found", n)
		}

		cf := resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(n, "product_version", found.GetProductVersion()),
			resource.TestCheckResourceAttr(n, "cluster_id", found.GetClusterId()),
			resource.TestCheckResourceAttr(n, "name", found.GetAppName()),
			resource.TestCheckResourceAttr(n, "namespace", found.GetAppNamespace()),
			resource.TestCheckResourceAttr(n, "product_id", found.GetProductId()),
			resource.TestCheckResourceAttr(n, "product_name", found.GetProductName()),
			resource.TestCheckResourceAttr(n, "status", found.GetStatus().String()),
			test.AccCheckCreatedAtAttr(n),
		)

		return cf(s)
	}
}

func testAccCheckHelmReleaseDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_kubernetes_marketplace_helm_release" {
			continue
		}

		_, err := config.SDK.KubernetesMarketplace().HelmRelease().Get(context.Background(), &k8s_marketplace.GetHelmReleaseRequest{
			Id: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex.Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("Helm Release still exists")
		}
	}

	return nil
}

func testAccK8sMarketplaceHelmRelease_basic(appName, versionID string) string {
	deps := testAccKubernetesCluster()
	return deps + fmt.Sprintf(`
resource "yandex_kubernetes_marketplace_helm_release" "example" {
  cluster_id = yandex_kubernetes_cluster.test.id
  product_version = "%s"
  name            = "%s"
  namespace       = "default"
}
`, versionID, appName)
}

func testAccKubernetesCluster() string {
	deps := testAccKubernetesClusterDeps()
	return deps + fmt.Sprintf(`
resource "yandex_kubernetes_cluster" "test" {
  depends_on = [
    yandex_resourcemanager_folder_iam_member.cluster_sa,
  ]

  release_channel = "%s"

  master {
    zonal {
      zone = "%s"
    }

    public_ip = true
  }

  service_account_id      = yandex_iam_service_account.cluster-sa.id
  node_service_account_id = yandex_iam_service_account.cluster-sa.id

  network_id = data.yandex_vpc_network.common.id
}

resource "yandex_kubernetes_node_group" "test" {
  cluster_id = yandex_kubernetes_cluster.test.id

  instance_template {
    platform_id = "standard-v2"

    network_interface {
      nat        = true
      subnet_ids = [data.yandex_vpc_subnet.common-sub.id]
    }

    resources {
      memory = 4
      cores  = 2
    }

    boot_disk {
      type = "network-hdd"
      size = 64
    }

    scheduling_policy {
      preemptible = false
    }
  }

  scale_policy {
    fixed_scale {
      size = 1
    }
  }

  allocation_policy {
    location {
      zone = "%s"
    }
  }
}
`, k8s.ReleaseChannel_REGULAR.String(), defaultZone, defaultZone)
}

func testAccKubernetesClusterDeps() string {
	return fmt.Sprintf(`
locals {
	common_folder_id = "b1gl18oo8atfpg1m7f2l"
}

resource "yandex_iam_service_account" "cluster-sa" {
  name = "test-cluster-sa"
}

resource "yandex_resourcemanager_folder_iam_member" "cluster_sa" {
  folder_id = "%s"
  role      = "admin"
  member    = "serviceAccount:${yandex_iam_service_account.cluster-sa.id}"
}

data "yandex_vpc_network" "common" {
  folder_id = local.common_folder_id
  name = "k8s-marketplace-tf-acceptance-net-fc32"
}

data "yandex_vpc_subnet" "common-sub" {
  folder_id = local.common_folder_id
  name = "k8s-marketplace-tf-acceptance-net-fc32-%s"
}
`, test.GetExampleFolderID(), defaultZone)
}
