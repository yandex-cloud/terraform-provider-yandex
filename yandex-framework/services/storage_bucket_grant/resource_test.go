package storage_bucket_grant_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccStorageBucketGrantResource_basic(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_basic(bucketName, test.GetExampleFolderID(), test.GetExampleUserID1()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "1"),
				),
			},
		},
	})
}

func testAccStorageBucketGrantConfig_basic(bucketName string, folderID string, userID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  grant {
    id          = "` + userID + `"
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }
}
`
}
