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

const folderPrefix = "tfacc"

func init() {
	resource.AddTestSweepers("yandex_resourcemanager_folder", &resource.Sweeper{
		Name:         "yandex_resourcemanager_folder",
		F:            testSweepFolders,
		Dependencies: []string{},
	})
}

func sweepFolderOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(yandexResourceManagerFolderDeleteTimeout)
	defer cancel()

	op, err := conf.sdk.ResourceManager().Folder().Delete(ctx, &resourcemanager.DeleteFolderRequest{
		FolderId:    id,
		DeleteAfter: timestamppb.Now(),
	})

	return handleSweepOperation(ctx, conf, op, err)
}

func testSweepFolders(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &resourcemanager.ListFoldersRequest{CloudId: conf.CloudID}
	it := conf.sdk.ResourceManager().Folder().FolderIterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		if !strings.HasPrefix(it.Value().Name, folderPrefix) {
			continue
		}
		id := it.Value().GetId()
		if !sweepWithRetry(sweepFolderOnce, conf, "Folder", id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Folder %q", id))
		}
	}

	return result.ErrorOrNil()
}

func newFolderInfo() *resourceFolderInfo {
	return &resourceFolderInfo{
		Name:        acctest.RandomWithPrefix(folderPrefix),
		Description: acctest.RandString(20),
		LabelKey:    "label_key",
		LabelValue:  "label_value",
	}
}

func TestAccResourceManagerFolder_create(t *testing.T) {
	t.Parallel()

	folderInfo := newFolderInfo()

	// TODO: remove me once the problem is resolved.
	t.Log(testAccResourceManagerFolder(folderInfo))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFolderDestroy,
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

func testAccCheckFolderDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "yandex_resourcemanager_folder" {
			continue
		}

		_, err := config.sdk.ResourceManager().Folder().Get(context.Background(), &resourcemanager.GetFolderRequest{
			FolderId: rs.Primary.ID,
		})
		if err == nil {
			return fmt.Errorf("Folder still exists")
		}
	}

	return nil
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
