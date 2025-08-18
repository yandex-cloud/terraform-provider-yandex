package storage_bucket_grant_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
	storage "github.com/yandex-cloud/terraform-provider-yandex/pkg/storage/s3"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	storage_bucket_grant "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/storage_bucket_grant"
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
    permissions = ["READ", "WRITE"]
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
    permissions = ["WRITE", "READ"]
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

func TestDetectACLFromGrants(t *testing.T) {
	stringPtr := func(s string) *string {
		return &s
	}

	tests := []struct {
		name     string
		grants   []*s3.Grant
		expected string
	}{
		{
			name:     "Empty grants should return private",
			grants:   []*s3.Grant{},
			expected: storage.BucketACLPrivate,
		},
		{
			name:     "Nil grants should return private",
			grants:   nil,
			expected: storage.BucketACLPrivate,
		},
		{
			name: "AllUsers with READ permission should return public-read",
			grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: stringPtr(storage.PermissionRead),
				},
			},
			expected: storage.BucketCannedACLPublicRead,
		},
		{
			name: "AllUsers with READ and WRITE permissions should return public-read-write",
			grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: stringPtr(storage.PermissionRead),
				},
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: stringPtr(storage.PermissionWrite),
				},
			},
			expected: storage.BucketCannedACLPublicReadWrite,
		},
		{
			name: "AuthenticatedUsers with READ permission should return authenticated-read",
			grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AuthenticatedUsers"),
					},
					Permission: stringPtr(storage.PermissionRead),
				},
			},
			expected: storage.BucketCannedACLAuthenticatedRead,
		},
		{
			name: "AllUsers with extra permissions should return empty string",
			grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: stringPtr(storage.PermissionRead),
				},
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: stringPtr(storage.PermissionWrite),
				},
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: stringPtr(storage.PermissionFullControl),
				},
			},
			expected: "",
		},
		{
			name: "Multiple grantees should return empty string",
			grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: stringPtr(storage.PermissionRead),
				},
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AuthenticatedUsers"),
					},
					Permission: stringPtr(storage.PermissionRead),
				},
			},
			expected: "",
		},
		{
			name: "CanonicalUser grants should return empty string",
			grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeCanonicalUser),
						ID:   stringPtr("user-id-123"),
					},
					Permission: stringPtr(storage.PermissionRead),
				},
			},
			expected: "",
		},
		{
			name: "Grants with nil grantee should be ignored",
			grants: []*s3.Grant{
				{
					Grantee:    nil,
					Permission: stringPtr(storage.PermissionRead),
				},
			},
			expected: "",
		},
		{
			name: "Grants with nil permission should be ignored",
			grants: []*s3.Grant{
				{
					Grantee: &s3.Grantee{
						Type: stringPtr(storage.TypeGroup),
						URI:  stringPtr("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: nil,
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := storage_bucket_grant.DetectACLFromGrants(tt.grants)
			assert.Equal(t, tt.expected, result)
		})
	}
}
