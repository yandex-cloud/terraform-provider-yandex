package yandex_resourcemanager_cloud_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagerv1sdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/internal/provider/framework"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

const cloudPrefix = "tfacc"

func newCloudInfo() *resourceCloudInfo {
	return &resourceCloudInfo{
		Name:        acctest.RandomWithPrefix(cloudPrefix),
		Description: acctest.RandString(20),
		LabelKey:    "label_key",
		LabelValue:  "label_value",
	}
}

func TestAccResourceManagerCloud_create(t *testing.T) {
	t.Parallel()

	cloudInfo := newCloudInfo()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckCloudDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceManagerCloud(cloudInfo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_resourcemanager_cloud.foobar", "name", cloudInfo.Name),
					resource.TestCheckResourceAttr("yandex_resourcemanager_cloud.foobar", "description", cloudInfo.Description),
					resource.TestCheckResourceAttr("yandex_resourcemanager_cloud.foobar", fmt.Sprintf("labels.%s", cloudInfo.LabelKey), cloudInfo.LabelValue),
				),
			},
			{
				ResourceName:      "yandex_resourcemanager_cloud.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceManagerCloud_UpgradeFromSDKv2(t *testing.T) {
	t.Parallel()

	cloudInfo := newCloudInfo()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { test.AccPreCheck(t) },
		CheckDestroy: testAccCheckCloudDestroy,
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccResourceManagerCloud(cloudInfo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_resourcemanager_cloud.foobar", "name", cloudInfo.Name),
					resource.TestCheckResourceAttr("yandex_resourcemanager_cloud.foobar", "description", cloudInfo.Description),
					resource.TestCheckResourceAttr("yandex_resourcemanager_cloud.foobar", fmt.Sprintf("labels.%s", cloudInfo.LabelKey), cloudInfo.LabelValue),
				),
			},
			{
				ProtoV6ProviderFactories: test.AccProviderFactories,
				Config:                   testAccResourceManagerCloud(cloudInfo),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckCloudDestroy(s *terraform.State) error {
	config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_resourcemanager_cloud" {
			continue
		}

		_, err := resourcemanagerv1sdk.NewCloudClient(config.SDKv2).Get(context.Background(), &resourcemanager.GetCloudRequest{
			CloudId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Cloud still exists")
		}
	}

	return nil
}

type resourceCloudInfo struct {
	Name        string
	Description string

	LabelKey   string
	LabelValue string
}

func testAccResourceManagerCloud(info *resourceCloudInfo) string {
	return fmt.Sprintf(`
resource "yandex_resourcemanager_cloud" "foobar" {
  name        = "%s"
  description = "%s"

  labels = {
    %s = "%s"
  }
}
`, info.Name, info.Description, info.LabelKey, info.LabelValue)
}
