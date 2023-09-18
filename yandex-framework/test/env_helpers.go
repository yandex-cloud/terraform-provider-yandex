package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider-config"
)

const providerDefaultValueInsecure = false
const providerDefaultValuePlaintext = false
const providerDefaultValueEndpoint = "api.cloud.yandex.net:443"

var testAccProviders map[string]tfprotov5.ProviderServer
var testAccProviderFactories map[string]func() (tfprotov5.ProviderServer, error)

// WARNING!!!! do not use testAccProviderEmptyFolder in tests, that use testAccCheck***Destroy functions.
// testAccCheck***Destroy functions tend to use static testAccProviderServer
var testAccProviderEmptyFolder map[string]tfprotov5.ProviderServer

var testAccProviderServer tfprotov5.ProviderServer
var testAccProvider provider.Provider

var testAccEnvVars = []string{
	"YC_FOLDER_ID",
	"YC_ZONE",
	"YC_LOGIN",
	"YC_LOGIN_2",
	"YC_STORAGE_ENDPOINT_URL",
	"YC_MESSAGE_QUEUE_ENDPOINT",
}

var testAccForAuthEnvVars = []string{
	"YC_TOKEN",
	"YC_SERVICE_ACCOUNT_KEY_FILE",
}

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

func NewFrameworkProividerServer(ctx context.Context) (func() tfprotov5.ProviderServer, error) {
	providers := []func() tfprotov5.ProviderServer{
		providerserver.NewProtocol5(testAccProvider),
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}

func init() {
	testAccProvider = yandex_framework.NewFrameworkProvider()
	testAccProviderFunc, _ := NewFrameworkProividerServer(context.Background())
	testAccProviderServer = testAccProviderFunc()
	testAccProviderFactories = map[string]func() (tfprotov5.ProviderServer, error){
		"yandex": func() (tfprotov5.ProviderServer, error) {
			return testAccProviderServer, nil
		},
	}
	/*
		testAccProviderEmptyFolder = map[string]provider.Provider{
			"yandex": NewFrameworkProvider(),
		}
	*/
	if os.Getenv("TF_ACC") != "" {
		if err := setTestIDs(); err != nil {
			panic(err)
		}
	}
}

func testAccPreCheck(t *testing.T) {
	for _, varName := range testAccEnvVars {
		if val := os.Getenv(varName); val == "" {
			t.Fatalf("%s must be set for acceptance tests", varName)
		}
	}

	for _, varName := range testAccForAuthEnvVars {
		if val := os.Getenv(varName); val != "" {
			return
		}
	}
	t.Fatalf("one of the variables: %s must be set for acceptance tests", strings.Join(testAccForAuthEnvVars, ", "))
}

func getExampleFolderName() string {
	return testFolderName
}

func getExampleCloudName() string {
	return testCloudName
}

func getExampleRoleID() string {
	return testRoleID
}

func getExampleCloudID() string {
	return testCloudID
}

func getExampleOrganizationID() string {
	return testOrganizationID
}

func getExampleFolderID() string {
	return testFolderID
}

func getExampleUserID1() string {
	return testUserID1
}

func getExampleUserID2() string {
	return testUserID2
}

func getExampleUserLogin1() string {
	return testUserLogin1
}

func getExampleUserLogin2() string {
	return testUserLogin2
}

func getExampleStorageEndpoint() string {
	return testStorageEndpoint
}

func setTestIDs() error {
	// init sdk client based on env var
	envEndpoint := os.Getenv("YC_ENDPOINT")
	if envEndpoint == "" {
		envEndpoint = common.DefaultEndpoint
	}
	ctx := context.Background()

	providerConfig := &provider_config.Config{
		ProviderState: provider_config.State{
			Token:                          types.StringValue(os.Getenv("YC_TOKEN")),
			ServiceAccountKeyFileOrContent: types.StringValue(os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE")),
		},
	}
	credentials, err := providerConfig.Credentials(ctx)
	if err != nil {
		return err
	}

	config := &ycsdk.Config{
		Credentials: credentials,
		Endpoint:    envEndpoint,
	}

	sdk, err := ycsdk.Build(ctx, *config)
	if err != nil {
		return err
	}

	// setup example ID values for test cases
	testCloudID = os.Getenv("YC_CLOUD_ID")
	testOrganizationID = os.Getenv("YC_ORGANIZATION_ID")

	testFolderID = os.Getenv("YC_FOLDER_ID")
	folder := getFolderByID(sdk, testFolderID)
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
	testCloudName = getCloudNameByID(sdk, testCloudID)

	testUserLogin1 = os.Getenv("YC_LOGIN")
	testUserLogin2 = os.Getenv("YC_LOGIN_2")

	testUserID1 = loginToUserID(sdk, testUserLogin1)
	testUserID2 = loginToUserID(sdk, testUserLogin2)

	testStorageEndpoint = os.Getenv("YC_STORAGE_ENDPOINT_URL")

	return nil
}

func getCloudNameByID(sdk *ycsdk.SDK, cloudID string) string {
	cloud, err := sdk.ResourceManager().Cloud().Get(context.Background(), &resourcemanager.GetCloudRequest{
		CloudId: cloudID,
	})
	if err != nil {
		log.Printf("could not get cloud name for %s: %s", cloudID, err)
		if reqID, ok := isRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return "no cloud name detected"
	}
	return cloud.Name
}

func getFolderByID(sdk *ycsdk.SDK, folderID string) *resourcemanager.Folder {
	folder, err := sdk.ResourceManager().Folder().Get(context.Background(), &resourcemanager.GetFolderRequest{
		FolderId: folderID,
	})
	if err != nil {
		log.Printf("could not get folder name for %s: %s", folderID, err)
		if reqID, ok := isRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return nil
	}
	return folder
}

func loginToUserID(sdk *ycsdk.SDK, loginName string) (userID string) {
	account, err := sdk.IAM().YandexPassportUserAccount().GetByLogin(context.Background(), &iam.GetUserAccountByLoginRequest{
		Login: loginName,
	})
	if err != nil {
		log.Printf("could not get user Id for %s: %s", loginName, err)
		if reqID, ok := isRequestIDPresent(err); ok {
			log.Printf("[DEBUG] request ID is %s\n", reqID)
		}
		return "not initialized"
	}
	return account.Id
}

func saveAndUnsetEnvVars(varNames []string) map[string]string {
	storage := make(map[string]string, len(varNames))

	for _, v := range varNames {
		storage[v] = os.Getenv(v)
		os.Unsetenv(v)
	}

	return storage
}

func restoreEnvVars(storage map[string]string) error {
	for varName, varValue := range storage {
		if err := os.Setenv(varName, varValue); err != nil {
			return err
		}
	}
	return nil
}

func isRequestIDPresent(err error) (string, bool) {
	st, ok := status.FromError(err)
	if ok {
		for _, d := range st.Details() {
			if reqInfo, ok := d.(*errdetails.RequestInfo); ok {
				return reqInfo.RequestId, true
			}
		}
	}
	return "", false
}
