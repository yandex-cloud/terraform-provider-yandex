package storage_bucket_iam_binding_test

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	iam_binding "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/storage_bucket_iam_binding"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccStorageBucketResourceIamBinding(t *testing.T) {
	var (
		bucketName          = test.ResourceName(63)
		bindingResourceName = "yandex_storage_bucket_iam_binding.test-bucket-binding"

		userID = "allUsers"
		role   = "storage.admin"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketIamBindingConfig(bucketName, test.GetExampleFolderID(), role, userID),
				Check: resource.ComposeTestCheckFunc(
					test.BucketExists(bucketName),
					testAccStorageBucketProjectIam(bindingResourceName, role, []string{"system:" + userID}),
				),
			},
		},
	})
}

func TestAccStorageBucketResourceIamBindingUpdate(t *testing.T) {
	var (
		bucketName          = test.ResourceName(63)
		bindingResourceName = "yandex_storage_bucket_iam_binding.test-bucket-binding"

		userID1 = "allUsers"
		userID2 = "allAuthenticatedUsers"
		role1   = "storage.admin"
		role2   = "storage.viewer"
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			// Create binding with one user and role
			{
				Config: testAccStorageBucketIamBindingConfig(bucketName, test.GetExampleFolderID(), role1, userID1),
				Check: resource.ComposeTestCheckFunc(
					test.BucketExists(bucketName),
					testAccStorageBucketProjectIam(bindingResourceName, role1, []string{"system:" + userID1}),
				),
			},
			// Update binding - change role
			{
				Config: testAccStorageBucketIamBindingConfig(bucketName, test.GetExampleFolderID(), role2, userID1),
				Check: resource.ComposeTestCheckFunc(
					test.BucketExists(bucketName),
					testAccStorageBucketProjectIam(bindingResourceName, role2, []string{"system:" + userID1}),
					// Check that old role is removed
					testAccStorageBucketProjectIamNotExists(bindingResourceName, role1),
				),
			},
			// Update binding - add second user
			{
				Config: testAccStorageBucketIamBindingConfigMultipleMembers(bucketName, test.GetExampleFolderID(), role2, userID1, userID2),
				Check: resource.ComposeTestCheckFunc(
					test.BucketExists(bucketName),
					testAccStorageBucketProjectIam(bindingResourceName, role2, []string{"system:" + userID1, "system:" + userID2}),
				),
			},
		},
	})
}

func testAccStorageBucketIamBindingConfig(bucketName, folderID, role, userID string) string {
	return fmt.Sprintf(`
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "%s"
  folder_id = "%s"
}

resource "yandex_storage_bucket_iam_binding" "test-bucket-binding" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  role = "%s"
  members = ["system:%s"]
}
`, bucketName, folderID, role, userID)
}

func testAccStorageBucketProjectIam(resourceName, role string, members []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		bucketName := rs.Primary.Attributes["bucket"]
		bucketResolver := sdkresolvers.BucketResolver(bucketName)
		if err := config.SDK.Resolve(context.Background(), bucketResolver); err != nil {
			return fmt.Errorf("Cannot get ResourceId for bucket %s", bucketName)
		}
		resourceId := bucketResolver.ID()

		bucketUpdater := iam_binding.BucketIAMUpdater{
			ResourceId:     resourceId,
			Bucket:         bucketName,
			ProviderConfig: &config,
		}

		bindings, err := bucketUpdater.GetAccessBindings(context.Background(), resourceId)
		if err != nil {
			return err
		}

		var roleMembers []string
		for _, binding := range bindings {
			if binding.RoleId == role {
				member := binding.Subject.Type + ":" + binding.Subject.Id
				roleMembers = append(roleMembers, member)
			}
		}
		sort.Strings(members)
		sort.Strings(roleMembers)

		if reflect.DeepEqual(members, roleMembers) {
			return nil
		}

		return fmt.Errorf("binding found but expected members is %v, got %v", members, roleMembers)
	}
}

func testAccStorageBucketIamBindingConfigMultipleMembers(bucketName, folderID, role, userID1, userID2 string) string {
	return fmt.Sprintf(`
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "%s"
  folder_id = "%s"
}

resource "yandex_storage_bucket_iam_binding" "test-bucket-binding" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  role = "%s"
  members = ["system:%s", "system:%s"]
}
`, bucketName, folderID, role, userID1, userID2)
}

func testAccStorageBucketProjectIamNotExists(resourceName, role string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		bucketName := rs.Primary.Attributes["bucket"]
		bucketResolver := sdkresolvers.BucketResolver(bucketName)
		if err := config.SDK.Resolve(context.Background(), bucketResolver); err != nil {
			return fmt.Errorf("Cannot get ResourceId for bucket %s", bucketName)
		}
		resourceId := bucketResolver.ID()

		bucketUpdater := iam_binding.BucketIAMUpdater{
			ResourceId:     resourceId,
			Bucket:         bucketName,
			ProviderConfig: &config,
		}

		bindings, err := bucketUpdater.GetAccessBindings(context.Background(), resourceId)
		if err != nil {
			return err
		}

		// Check that role does NOT exist
		for _, binding := range bindings {
			if binding.RoleId == role {
				return fmt.Errorf("role %s should not exist, but found binding with members", role)
			}
		}

		return nil
	}
}
