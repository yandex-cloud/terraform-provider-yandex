package yandex

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const cloudPrefix = "tfacc"

func init() {
	resource.AddTestSweepers("yandex_resourcemanager_cloud", &resource.Sweeper{
		Name:         "yandex_resourcemanager_cloud",
		F:            testSweepClouds,
		Dependencies: []string{},
	})
}

func sweepCloudOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexResourceManagerCloudDeleteTimeout)
	defer cancel()

	op, err := conf.sdk.ResourceManager().Cloud().Delete(ctx, &resourcemanager.DeleteCloudRequest{
		CloudId:     id,
		DeleteAfter: timestamppb.Now(),
	})

	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepClouds(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &resourcemanager.ListCloudsRequest{OrganizationId: conf.OrganizationID}
	it := conf.sdk.ResourceManager().Cloud().CloudIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		if !strings.HasPrefix(it.Value().Name, cloudPrefix) {
			continue
		}
		id := it.Value().GetId()
		if !sweepWithRetry(sweepCloudOnce, conf, "Cloud", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Cloud %q", id))
		}
	}

	return result.ErrorOrNil()
}

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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudDestroy,
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

func testAccCheckCloudDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_resourcemanager_cloud" {
			continue
		}

		_, err := config.sdk.ResourceManager().Cloud().Get(context.Background(), &resourcemanager.GetCloudRequest{
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
	// language=tf
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
