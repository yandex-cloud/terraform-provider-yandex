package storage_bucket_policy_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	storage "github.com/yandex-cloud/terraform-provider-yandex/pkg/storage/s3"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
)

// TestMain - add sweepers flag to the go test command
// important for sweepers run.
func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccStorageBucketResourcePolicy(t *testing.T) {
	var (
		bucketName         = test.ResourceName(63)
		policyResourceName = "yandex_storage_bucket_policy.test-bucket-policy"
		testPolicy         = testBucketPolicy(bucketName)
		updatedTestPolicy  = testBucketPolicyUpdated(bucketName)
	)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             test.AccCheckBucketDestroy(bucketName),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketPolicyConfig(bucketName, test.GetExampleFolderID(), testPolicy),
				Check: resource.ComposeTestCheckFunc(
					test.BucketExists(bucketName),
					testAccStorageBucketPolicyExists(policyResourceName, bucketName),
					resource.TestCheckResourceAttr(policyResourceName, "bucket", bucketName),
					resource.TestCheckResourceAttrSet(policyResourceName, "policy"),
				),
			},
			{
				Config: testAccStorageBucketPolicyConfig(bucketName, test.GetExampleFolderID(), updatedTestPolicy),
				Check: resource.ComposeTestCheckFunc(
					test.BucketExists(bucketName),
					testAccStorageBucketPolicyExists(policyResourceName, bucketName),
					resource.TestCheckResourceAttr(policyResourceName, "bucket", bucketName),
					resource.TestCheckResourceAttrSet(policyResourceName, "policy"),
				),
			},
		},
	})
}

func testAccStorageBucketPolicyConfig(bucketName, folderID, policy string) string {
	return fmt.Sprintf(`
resource "yandex_storage_bucket" "test-bucket" {
  bucket = "%s"
  folder_id = "%s"
}

resource "yandex_storage_bucket_policy" "test-bucket-policy" {
  bucket = yandex_storage_bucket.test-bucket.bucket
  policy = <<POLICY
%s
POLICY
}
`, bucketName, folderID, policy)
}

func testAccStorageBucketPolicyExists(resourceName, bucketName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		config := test.AccProvider.(*yandex_framework.Provider).GetConfig()

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", resourceName)
		}

		bucket := rs.Primary.Attributes["bucket"]
		if bucket != bucketName {
			return fmt.Errorf("bucket name mismatch: expected %s, got %s", bucketName, bucket)
		}

		s3Client, err := storage.GetS3Client(context.Background(), "", "", &config)
		if err != nil {
			return fmt.Errorf("error getting S3 client: %s", err)
		}

		policy, err := s3Client.GetBucketPolicy(context.Background(), bucket)
		if err != nil {
			return fmt.Errorf("error getting bucket policy: %s", err)
		}

		if policy == "" {
			return fmt.Errorf("bucket policy is empty")
		}

		return nil
	}
}

func testBucketPolicy(bucketName string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": [
        "arn:aws:s3:::%s"
      ]
    }
  ]
}`, bucketName)
}

func testBucketPolicyUpdated(bucketName string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": "*",
      "Action": "*",
      "Resource": [
        "arn:aws:s3:::%s/*",
        "arn:aws:s3:::%s"
      ]
    },
    {
      "Effect": "Deny",
      "Principal": "*",
      "Action": "s3:PutObject",
      "Resource": [
        "arn:aws:s3:::%s/*"
      ]
    }
  ]
}`, bucketName, bucketName, bucketName)
}
