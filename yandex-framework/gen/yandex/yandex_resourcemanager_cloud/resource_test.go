package yandex_resourcemanager_cloud_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagerv1sdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
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

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sweepCloudOnce(conf *provider_config.Config, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	op, err := resourcemanagerv1sdk.NewCloudClient(conf.SDKv2).Delete(ctx, &resourcemanager.DeleteCloudRequest{
		CloudId:     id,
		DeleteAfter: timestamppb.Now(),
	})
	_, err = op.Wait(ctx)
	return err
}

func testSweepClouds(_ string) error {
	if os.Getenv("YC_ENABLE_CLOUD_SWEEPING") != "1" {
		return nil
	}

	fmt.Println("Sweeping Clouds")

	conf, err := test.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &resourcemanager.ListCloudsRequest{OrganizationId: conf.ProviderState.OrganizationID.ValueString()}

	resp, err := resourcemanagerv1sdk.NewCloudClient(conf.SDKv2).List(context.Background(), req)
	if err != nil {
		return fmt.Errorf("error getting clouds: %s", err)
	}
	result := &multierror.Error{}
	for _, cloud := range resp.Clouds {
		if !strings.HasPrefix(cloud.Name, cloudPrefix) {
			continue
		}
		id := cloud.GetId()
		if !test.SweepWithRetry(sweepCloudOnce, conf, "Cloud", id) {
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
