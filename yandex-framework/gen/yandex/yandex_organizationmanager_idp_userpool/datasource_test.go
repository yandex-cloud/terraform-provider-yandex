package yandex_organizationmanager_idp_userpool_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	idp "github.com/yandex-cloud/go-genproto/yandex/cloud/organizationmanager/v1/idp"
	idpsdk "github.com/yandex-cloud/go-sdk/services/organizationmanager/v1/idp"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAccDataSourceOrganizationManagerIdpUserpool_byID(t *testing.T) {
	userpoolName := acctest.RandomWithPrefix("tf-userpool")
	organizationID := test.GetExampleOrganizationID()
	testSubdomain := acctest.RandomWithPrefix("tf-acc-test-subdomain")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckIdpUserpoolDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceIdpUserpoolConfig(userpoolName, organizationID, testSubdomain, true),
				Check: resource.ComposeTestCheckFunc(
					test.AccCheckResourceIDField("data.yandex_organizationmanager_idp_userpool.source", "userpool_id"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_userpool.source", "name", userpoolName),
					resource.TestCheckResourceAttrSet("data.yandex_organizationmanager_idp_userpool.source", "id"),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_userpool.source", "organization_id", organizationID),
					resource.TestCheckResourceAttr("data.yandex_organizationmanager_idp_userpool.source", "labels.test_label", "example-label-value"),
					test.AccCheckCreatedAtAttr("data.yandex_organizationmanager_idp_userpool.source"),
				),
			},
		},
	})
}

func testAccDataSourceIdpUserpoolConfig(name, organizationID, defaultSubdomain string, useID bool) string {
	if useID {
		return testAccDataSourceIdpUserpoolResourceConfig(name, organizationID, defaultSubdomain) + idpUserpoolDataByIDConfig
	}

	return testAccDataSourceIdpUserpoolResourceConfig(name, organizationID, defaultSubdomain)
}

func testAccDataSourceIdpUserpoolResourceConfig(name, organizationID, defaultSubdomain string) string {
	return fmt.Sprintf(`
resource "yandex_organizationmanager_idp_userpool" "foobar" {
  name              = "%s"
  organization_id   = "%s"
  default_subdomain = "%s"

  labels = {
    test_label = "example-label-value"
  }
}
`, name, organizationID, defaultSubdomain)
}

const idpUserpoolDataByIDConfig = `
data "yandex_organizationmanager_idp_userpool" "source" {
  userpool_id = "${yandex_organizationmanager_idp_userpool.foobar.userpool_id}"
}
`

func testAccCheckIdpUserpoolDataSourceDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_organizationmanager_idp_userpool" {
			continue
		}

		_, err := idpsdk.NewUserpoolClient(config.SDKv2).Get(context.Background(), &idp.GetUserpoolRequest{
			UserpoolId: rs.Primary.ID,
		})

		if err != nil {
			if grpcStatus, ok := status.FromError(err); ok && grpcStatus != nil && grpcStatus.Code() == codes.NotFound {
				return nil
			} else if ok {
				return fmt.Errorf("Error while requesting Yandex Cloud: grpc code error : %d, http message error: %s", grpcStatus.Code(), grpcStatus.Message())
			}
			return fmt.Errorf("IdpUserpool still exists")
		}
	}

	return nil
}
