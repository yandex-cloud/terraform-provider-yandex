package acceptance

import (
	"fmt"
	"testing"

	resourcemanagerfixtures "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/testfixtures"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	acceptance "github.com/yandex-cloud/terraform-provider-yandex/internal/acceptance"
)

const folderPrefix = "tfacc"

func newFolderInfo() *resourceFolderInfo {
	return &resourceFolderInfo{
		Name:        acctest.RandomWithPrefix(folderPrefix),
		Description: acctest.RandString(20),
		LabelKey:    "label_key",
		LabelValue:  "label_value",
	}
}

func TestAccResourceManagerFolder_create(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	folderInfo := newFolderInfo()

	// TODO: remove me once the problem is resolved.
	t.Log(testAccResourceManagerFolder(folderInfo))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acceptance.PreCheck(t) },
		ProtoV6ProviderFactories: suite.Protocol6Factories(),
		CheckDestroy:             resourcemanagerfixtures.CheckFolderDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceManagerFolder(folderInfo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_resourcemanager_folder.foobar", "name", folderInfo.Name),
					resource.TestCheckResourceAttr("yandex_resourcemanager_folder.foobar", "description", folderInfo.Description),
					resource.TestCheckResourceAttr("yandex_resourcemanager_folder.foobar", fmt.Sprintf("labels.%s", folderInfo.LabelKey), folderInfo.LabelValue),
				),
			},
		},
	})
}

func TestAccResourceManagerFolder_UpgradeFromSDKv2(t *testing.T) {
	suite := testSuite{acceptance.New(t)}
	t.Parallel()

	folderInfo := newFolderInfo()

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acceptance.PreCheck(t) },
		CheckDestroy: resourcemanagerfixtures.CheckFolderDestroy(suite.Config),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"yandex": {
						VersionConstraint: "0.150.0",
						Source:            "yandex-cloud/yandex",
					},
				},
				Config: testAccResourceManagerFolder(folderInfo),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_resourcemanager_folder.foobar", "name", folderInfo.Name),
					resource.TestCheckResourceAttr("yandex_resourcemanager_folder.foobar", "description", folderInfo.Description),
					resource.TestCheckResourceAttr("yandex_resourcemanager_folder.foobar", fmt.Sprintf("labels.%s", folderInfo.LabelKey), folderInfo.LabelValue),
				),
			},
			{
				ProtoV6ProviderFactories: suite.Protocol6Factories(),
				Config:                   testAccResourceManagerFolder(folderInfo),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

type resourceFolderInfo struct {
	Name        string
	Description string

	LabelKey   string
	LabelValue string
}

func testAccResourceManagerFolder(info *resourceFolderInfo) string {
	// language=tf
	return fmt.Sprintf(`
resource "yandex_resourcemanager_folder" "foobar" {
  name        = "%s"
  description = "%s"

  labels = {
    %s = "%s"
  }
}
`, info.Name, info.Description, info.LabelKey, info.LabelValue)
}
