package acceptanceenv

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	resourcemanagersdk "github.com/yandex-cloud/go-sdk/services/resourcemanager/v1"
	iamsdk "github.com/yandex-cloud/go-sdk/v2/services/iam/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"

	"sync"

	"github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

var testCloudID = "not initialized"
var testOrganizationID = "not initialized"
var testCloudName = "not initialized"
var testFolderID = "not initialized"
var testFolderName = "not initialized"
var testRoleID = "resource-manager.clouds.member"
var testUserLogin1 = "no user login"
var testUserLogin2 = "no user login"
var testUserID1 = "no user id"
var testUserID2 = "no user id"
var testStorageEndpoint = "no.storage.endpoint"

var load = sync.OnceFunc(func() {
	if os.Getenv("TF_ACC") != "" {
		if err := setTestIDs(); err != nil {
			panic(fmt.Errorf("initialize acceptance environment: %w", err))
		}
	}
})

func Load() { load() }

func FolderName() string {
	Load()
	return testFolderName
}

func CloudName() string {
	Load()
	return testCloudName
}

func RoleID() string {
	Load()
	return testRoleID
}

func CloudID() string {
	Load()
	return testCloudID
}

func OrganizationID() string {
	Load()
	return testOrganizationID
}

func FolderID() string {
	Load()
	return testFolderID
}

func UserID1() string {
	Load()
	return testUserID1
}

func UserID2() string {
	Load()
	return testUserID2
}

func UserLogin1() string {
	Load()
	return testUserLogin1
}

func UserLogin2() string {
	Load()
	return testUserLogin2
}

func StorageEndpoint() string {
	Load()
	return testStorageEndpoint
}

func setTestIDs() error {
	// init sdk client based on env var
	envEndpoint := os.Getenv("YC_ENDPOINT")
	if envEndpoint == "" {
		envEndpoint = common.DefaultEndpoint
	}

	providerConfig := &providerconfig.Config{
		Endpoint:                       envEndpoint,
		Token:                          os.Getenv("YC_TOKEN"),
		ServiceAccountKeyFileOrContent: os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE"),
	}
	ctx := context.Background()
	if err := providerConfig.InitAndValidate(ctx, "acceptance-tests", false); err != nil {
		return err
	}

	// setup example ID values for test cases
	testCloudID = os.Getenv("YC_CLOUD_ID")
	testOrganizationID = os.Getenv("YC_ORGANIZATION_ID")

	testFolderID = os.Getenv("YC_FOLDER_ID")
	folder := getFolderByID(providerConfig, testFolderID)
	if folder != nil {
		testFolderName = folder.Name
		if testCloudID == "" {
			testCloudID = folder.CloudId
		} else if testCloudID != folder.CloudId {
			return fmt.Errorf("Invalid cloud id: %s != %s", testCloudID, folder.CloudId)
		}
	} else {
		testFolderName = "no folder name detected"
	}
	testCloudName = getCloudNameByID(providerConfig, testCloudID)

	testUserLogin1 = os.Getenv("YC_LOGIN")
	testUserLogin2 = os.Getenv("YC_LOGIN_2")

	testUserID1 = loginToUserID(providerConfig, testUserLogin1)
	testUserID2 = loginToUserID(providerConfig, testUserLogin2)

	testStorageEndpoint = os.Getenv("YC_STORAGE_ENDPOINT_URL")

	return nil
}

func getCloudNameByID(config *providerconfig.Config, cloudID string) string {
	client := resourcemanagersdk.NewCloudClient(config.SDK)
	cloud, err := client.Get(context.Background(), &resourcemanager.GetCloudRequest{
		CloudId: cloudID,
	})
	if err != nil {
		log.Printf("could not get cloud name for %s: %s", cloudID, err)
		if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return "no cloud name detected"
	}
	return cloud.Name
}

func getFolderByID(config *providerconfig.Config, folderID string) *resourcemanager.Folder {
	client := resourcemanagersdk.NewFolderClient(config.SDK)
	folder, err := client.Get(context.Background(), &resourcemanager.GetFolderRequest{
		FolderId: folderID,
	})
	if err != nil {
		log.Printf("could not get folder name for %s: %s", folderID, err)
		if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return nil
	}
	return folder
}

func loginToUserID(config *providerconfig.Config, loginName string) (userID string) {
	client := iamsdk.NewYandexPassportUserAccountClient(config.SDK)
	account, err := client.GetByLogin(context.Background(), &iam.GetUserAccountByLoginRequest{
		Login: loginName,
	})
	if err != nil {
		log.Printf("could not get user Id for %s: %s", loginName, err)
		if reqID, ok := sdkcommon.IsRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return "not initialized"
	}
	return account.Id
}
