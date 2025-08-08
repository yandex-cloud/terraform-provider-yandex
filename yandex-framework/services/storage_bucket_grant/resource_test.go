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

func TestAccStorageBucketGrantResource_acl_private(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_acl_private(bucketName, test.GetExampleFolderID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "acl", "private"),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "0"),
				),
			},
		},
	})
}

func TestAccStorageBucketGrantResource_permissions_order(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_permissions_order_initial(bucketName, test.GetExampleFolderID(), test.GetExampleUserID1()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("yandex_storage_bucket_grant.test-bucket-grant", "grant.*", map[string]string{
						"permissions.#": "2",
					}),
				),
			},
			{
				Config:             testAccStorageBucketGrantConfig_permissions_order_reordered(bucketName, test.GetExampleFolderID(), test.GetExampleUserID1()),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccStorageBucketGrantResource_grants_to_acl_transition(t *testing.T) {
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
			{
				Config: testAccStorageBucketGrantConfig_acl_private(bucketName, test.GetExampleFolderID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "acl", "private"),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "0"),
				),
			},
		},
	})
}

func TestAccStorageBucketGrantResource_acl_to_grants_transition(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_acl_private(bucketName, test.GetExampleFolderID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "acl", "private"),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "0"),
				),
			},
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

func TestAccStorageBucketGrantResource_acl_public_read_stable(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_acl_public_read(bucketName, test.GetExampleFolderID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "acl", "public-read"),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "0"),
				),
			},
			{
				Config:             testAccStorageBucketGrantConfig_acl_public_read(bucketName, test.GetExampleFolderID()),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false, // Should be stable - no drift from implicit grants
			},
		},
	})
}

func TestAccStorageBucketGrantResource_acl_public_read_write_stable(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_acl_public_read_write(bucketName, test.GetExampleFolderID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "acl", "public-read-write"),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "0"),
				),
			},
			{
				Config:             testAccStorageBucketGrantConfig_acl_public_read_write(bucketName, test.GetExampleFolderID()),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false, // Should be stable - no drift from implicit grants
			},
		},
	})
}

func TestAccStorageBucketGrantResource_acl_authenticated_read_stable(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_acl_authenticated_read(bucketName, test.GetExampleFolderID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "acl", "authenticated-read"),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "0"),
				),
			},
			{
				Config:             testAccStorageBucketGrantConfig_acl_authenticated_read(bucketName, test.GetExampleFolderID()),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false, // Should be stable - no drift from implicit grants
			},
		},
	})
}

func TestAccStorageBucketGrantResource_acl_bucket_owner_full_control_stable(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_acl_bucket_owner_full_control(bucketName, test.GetExampleFolderID()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "acl", "bucket-owner-full-control"),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "0"),
				),
			},
			{
				Config:             testAccStorageBucketGrantConfig_acl_bucket_owner_full_control(bucketName, test.GetExampleFolderID()),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false, // Should be stable - no drift from implicit grants
			},
		},
	})
}

func TestAccStorageBucketGrantResource_grants_equivalent_to_public_read(t *testing.T) {
	bucketName := test.ResourceName(63)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_grants_public_read_equivalent(bucketName, test.GetExampleFolderID(), test.GetExampleUserID1()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("yandex_storage_bucket_grant.test-bucket-grant", "grant.*", map[string]string{
						"type":          "Group",
						"uri":           "http://acs.amazonaws.com/groups/global/AllUsers",
						"permissions.#": "1",
					}),
				),
			},
			{
				Config:             testAccStorageBucketGrantConfig_grants_public_read_equivalent(bucketName, test.GetExampleFolderID(), test.GetExampleUserID1()),
				PlanOnly:           true,
				ExpectNonEmptyPlan: false, // Should be stable
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

func testAccStorageBucketGrantConfig_permissions_order_initial(bucketName string, folderID string, userID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  grant {
    id          = "` + userID + `"
    permissions = ["READ", "WRITE"]
    type        = "CanonicalUser"
  }
}
`
}

func testAccStorageBucketGrantConfig_permissions_order_reordered(bucketName string, folderID string, userID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  grant {
    id          = "` + userID + `"
    permissions = ["READ", "WRITE"]
    type        = "CanonicalUser"
  }
}
`
}

func testAccStorageBucketGrantConfig_acl_private(bucketName string, folderID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  acl    = "private"
}
`
}

func testAccStorageBucketGrantConfig_acl_public_read(bucketName string, folderID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  acl    = "public-read"
}
`
}

func testAccStorageBucketGrantConfig_acl_public_read_write(bucketName string, folderID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  acl    = "public-read-write"
}
`
}

func testAccStorageBucketGrantConfig_acl_authenticated_read(bucketName string, folderID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  acl    = "authenticated-read"
}
`
}

func testAccStorageBucketGrantConfig_acl_bucket_owner_full_control(bucketName string, folderID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  acl    = "bucket-owner-full-control"
}
`
}

func testAccStorageBucketGrantConfig_grants_public_read_equivalent(bucketName string, folderID string, userID string) string {
	return `
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "` + bucketName + `"
  folder_id = "` + folderID + `"
}

resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = yandex_storage_bucket.test-bucket.bucket
    
  # Public read grant (equivalent to public-read ACL)
  grant {
    uri         = "http://acs.amazonaws.com/groups/global/AllUsers"
    permissions = ["READ"]
    type        = "Group"
  }
}
`
}
