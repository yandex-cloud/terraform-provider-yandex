package cloudregistry_registry

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cloudregistry "github.com/yandex-cloud/go-genproto/yandex/cloud/cloudregistry/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	cloudRegistryResource = "yandex_cloudregistry_registry.test-registry"
	cloudRegistryKind     = "DOCKER"
	cloudRegistryType     = "LOCAL"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccCloudRegistryRegistry_basic(t *testing.T) {
	registryName := acctest.RandomWithPrefix("tf-registry")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistryRegistryConfig(registryName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists(cloudRegistryResource),
					resource.TestCheckResourceAttrSet(cloudRegistryResource, "id"),
					resource.TestCheckResourceAttrSet(cloudRegistryResource, "folder_id"),
					resource.TestCheckResourceAttr(cloudRegistryResource, "name", registryName),
					resource.TestCheckResourceAttr(cloudRegistryResource, "kind", cloudRegistryKind),
					resource.TestCheckResourceAttr(cloudRegistryResource, "type", cloudRegistryType),
					resource.TestCheckResourceAttrSet(cloudRegistryResource, "status"),
					resource.TestCheckResourceAttrSet(cloudRegistryResource, "created_at"),
				),
			},
			{
				ResourceName:      cloudRegistryResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudRegistryRegistry_patternFilter(t *testing.T) {
	registryName := acctest.RandomWithPrefix("tf-registry")

	initialInclude := []string{"alpine/*", "nginx/*"}
	initialExclude := []string{"debian/*"}

	updatedInclude := []string{"alpine/*", "nginx/*", "busybox/*"}
	updatedExclude := []string{"debian/*", "ubuntu/*"}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudRegistryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCloudRegistryRegistryConfigWithPatternFilter(registryName, initialInclude, initialExclude),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists(cloudRegistryResource),
					resource.TestCheckResourceAttr(cloudRegistryResource, "pattern_filter.include_patterns.#", fmt.Sprint(len(initialInclude))),
					resource.TestCheckResourceAttr(cloudRegistryResource, "pattern_filter.exclude_patterns.#", fmt.Sprint(len(initialExclude))),
					testAccCheckCloudRegistryPatternFilter(cloudRegistryResource, initialInclude, initialExclude),
				),
			},
			{
				Config: testAccCloudRegistryRegistryConfigWithPatternFilter(registryName, updatedInclude, updatedExclude),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists(cloudRegistryResource),
					resource.TestCheckResourceAttr(cloudRegistryResource, "pattern_filter.include_patterns.#", fmt.Sprint(len(updatedInclude))),
					resource.TestCheckResourceAttr(cloudRegistryResource, "pattern_filter.exclude_patterns.#", fmt.Sprint(len(updatedExclude))),
					testAccCheckCloudRegistryPatternFilter(cloudRegistryResource, updatedInclude, updatedExclude),
				),
			},
			{
				Config: testAccCloudRegistryRegistryConfigWithPatternFilter(registryName, updatedInclude, nil),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudRegistryExists(cloudRegistryResource),
					resource.TestCheckResourceAttr(cloudRegistryResource, "pattern_filter.include_patterns.#", fmt.Sprint(len(updatedInclude))),
					resource.TestCheckResourceAttr(cloudRegistryResource, "pattern_filter.exclude_patterns.#", "0"),
					testAccCheckCloudRegistryPatternFilter(cloudRegistryResource, updatedInclude, nil),
				),
			},
			{
				ResourceName:      cloudRegistryResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCloudRegistryRegistryConfig(registryName string) string {
	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "test-registry" {
  name = "%s"
  kind = "%s"
  type = "%s"
}
`, registryName, cloudRegistryKind, cloudRegistryType)
}

func testAccCloudRegistryRegistryConfigWithPatternFilter(registryName string, include, exclude []string) string {
	block := ""
	if include != nil || exclude != nil {
		lines := []string{"  pattern_filter {"}
		if include != nil {
			lines = append(lines, fmt.Sprintf("    include_patterns = [%s]", quoteJoin(include)))
		}
		if exclude != nil {
			lines = append(lines, fmt.Sprintf("    exclude_patterns = [%s]", quoteJoin(exclude)))
		}
		lines = append(lines, "  }")
		block = strings.Join(lines, "\n")
	}

	return fmt.Sprintf(`
resource "yandex_cloudregistry_registry" "test-registry" {
  name = "%s"
  kind = "%s"
  type = "%s"

%s
}
`, registryName, cloudRegistryKind, cloudRegistryType, block)
}

func quoteJoin(items []string) string {
	if len(items) == 0 {
		return ""
	}
	return `"` + strings.Join(items, `", "`) + `"`
}

func testAccCheckCloudRegistryPatternFilter(n string, wantInclude, wantExclude []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()
		found, err := config.SDK.CloudRegistry().Registry().Get(context.Background(), &cloudregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		gotInclude := append([]string(nil), found.GetPatternFilter().GetIncludePatterns()...)
		gotExclude := append([]string(nil), found.GetPatternFilter().GetExcludePatterns()...)
		want := func(v []string) []string {
			out := append([]string(nil), v...)
			sort.Strings(out)
			return out
		}
		sort.Strings(gotInclude)
		sort.Strings(gotExclude)

		if !reflect.DeepEqual(gotInclude, want(wantInclude)) {
			return fmt.Errorf("include_patterns mismatch: want %v, got %v", wantInclude, gotInclude)
		}
		if !reflect.DeepEqual(gotExclude, want(wantExclude)) {
			return fmt.Errorf("exclude_patterns mismatch: want %v, got %v", wantExclude, gotExclude)
		}

		return nil
	}
}

func testAccCheckCloudRegistryExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		found, err := config.SDK.CloudRegistry().Registry().Get(context.Background(), &cloudregistry.GetRegistryRequest{
			RegistryId: rs.Primary.ID,
		})
		if err != nil {
			return err
		}

		if found.GetId() != rs.Primary.ID {
			return fmt.Errorf("Cloud Registry %s not found", n)
		}

		return nil
	}
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

		return fmt.Errorf("Cloud Registry still exists")
	}

	return nil
}
