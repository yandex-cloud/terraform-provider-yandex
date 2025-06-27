package storage_bucket_grant_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	storagepb "github.com/yandex-cloud/go-genproto/yandex/cloud/storage/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	test "github.com/yandex-cloud/terraform-provider-yandex/pkg/testhelpers"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestAccStorageBucketGrantResource_basic(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf(
			"Acceptance tests skipped unless env '%s' set",
			resource.EnvTfAcc)
		return
	}

	bucketName := test.ResourceName(63)

	envEndpoint := os.Getenv("YC_ENDPOINT")
	if envEndpoint == "" {
		envEndpoint = common.DefaultEndpoint
	}
	ctx := context.Background()

	providerConfig := &provider_config.Config{
		ProviderState: provider_config.State{
			Token:                          types.StringValue(os.Getenv("YC_TOKEN")),
			ServiceAccountKeyFileOrContent: types.StringValue(os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE")),
			Endpoint:                       types.StringValue(envEndpoint),
			StorageEndpoint:                types.StringValue(os.Getenv("YC_STORAGE_ENDPOINT_URL")),
		},
	}

	credentials, err := providerConfig.Credentials(ctx)
	if err != nil {
		t.Fatalf("Failed to init credentials: %s", err)
	}

	config := &ycsdk.Config{
		Credentials: credentials,
		Endpoint:    envEndpoint,
	}

	sdk, err := ycsdk.Build(ctx, *config)
	if err != nil {
		t.Fatalf("Failed to init sdk: %s", err)
	}

	_, err = sdk.StorageAPI().Bucket().Create(context.Background(), &storagepb.CreateBucketRequest{
		Name:     bucketName,
		FolderId: test.GetExampleFolderID(),
	})
	if err != nil {
		t.Fatalf("Failed to create bucket: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test.AccPreCheck(t) },
		ProtoV6ProviderFactories: test.AccProviderFactories,
		CheckDestroy:             testAccCheckStorageBucketGrantDestroy(sdk),
		Steps: []resource.TestStep{
			{
				Config: testAccStorageBucketGrantConfig_basic(bucketName, test.GetExampleUserID1()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "bucket", bucketName),
					resource.TestCheckResourceAttr("yandex_storage_bucket_grant.test-bucket-grant", "grant.#", "1"),
				),
			},
		},
	})

	_, err = sdk.StorageAPI().Bucket().Delete(context.Background(), &storagepb.DeleteBucketRequest{
		Name: bucketName,
	})
	if err != nil {
		t.Fatalf("Failed to delete bucket: %s", err)
	}
}

func testAccStorageBucketGrantConfig_basic(bucketName string, userID string) string {
	return `
resource "yandex_storage_bucket_grant" "test-bucket-grant" {
  bucket = "` + bucketName + `"
  grant {
    id          = "` + userID + `"
    permissions = ["FULL_CONTROL"]
    type        = "CanonicalUser"
  }
}
`
}

func testAccCheckStorageBucketGrantDestroy(sdk *ycsdk.SDK) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "yandex_storage_bucket_grant" {
				continue
			}

			bucketName := rs.Primary.Attributes["bucket"]
			if bucketName == "" {
				continue
			}

			aclOutput, err := sdk.StorageAPI().Bucket().Get(context.Background(), &storagepb.GetBucketRequest{
				Name: bucketName,
				View: storagepb.GetBucketRequest_VIEW_ACL,
			})

			if err != nil {
				continue
			}

			if len(aclOutput.Acl.Grants) > 0 {
				return fmt.Errorf("storage bucket grant still exists for bucket %s", bucketName)
			}
		}

		return nil
	}
}
