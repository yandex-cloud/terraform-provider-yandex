package testhelpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

var AccProviders map[string]tfprotov6.ProviderServer
var AccProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

// WARNING!!!! do not use testAccProviderEmptyFolder in tests, that use testAccCheck***Destroy functions.
// testAccCheck***Destroy functions tend to use static testAccProviderServer
var testAccProviderEmptyFolder map[string]tfprotov6.ProviderServer

var testAccProviderServer tfprotov6.ProviderServer
var AccProvider provider.Provider

var AccEnvVars = []string{
	"YC_FOLDER_ID",
	"YC_ZONE",
	"YC_LOGIN",
	"YC_LOGIN_2",
	"YC_STORAGE_ENDPOINT_URL",
	"YC_MESSAGE_QUEUE_ENDPOINT",
}

var AccForAuthEnvVars = []string{
	"YC_TOKEN",
	"YC_SERVICE_ACCOUNT_KEY_FILE",
}

var cloudID = "not initialized"
var organizationID = "not initialized"
var cloudName = "not initialized"
var folderID = "not initialized"
var folderName = "not initialized"
var roleID = "resource-manager.clouds.member"
var userLogin1 = "no user login"
var userLogin2 = "no user login"
var userID1 = "no user id"
var userID2 = "no user id"
var storageEndpoint = "no.storage.endpoint"

func NewFrameworkProviderServer(ctx context.Context) (func() tfprotov6.ProviderServer, error) {
	upgradedSdkProvider, _ := tf5to6server.UpgradeServer(
		context.Background(),
		yandex.NewSDKProvider().GRPCProvider,
	)
	providers := []func() tfprotov6.ProviderServer{
		providerserver.NewProtocol6(AccProvider),
		func() tfprotov6.ProviderServer {
			return upgradedSdkProvider
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		return nil, err
	}

	return muxServer.ProviderServer, nil
}

func init() {
	AccProvider = yandex_framework.NewFrameworkProvider()
	testAccProviderFunc, _ := NewFrameworkProviderServer(context.Background())
	testAccProviderServer = testAccProviderFunc()
	AccProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"yandex": func() (tfprotov6.ProviderServer, error) {
			return testAccProviderServer, nil
		},
	}

	if os.Getenv("TF_ACC") != "" {
		if err := setTestIDs(); err != nil {
			panic(err)
		}
	}
	//testAccProviderEmptyFolder = map[string]provider.Provider{
	//	"yandex": NewFrameworkProvider(),
	//}
}

func AccPreCheck(t *testing.T) {
	for _, varName := range AccEnvVars {
		if val := os.Getenv(varName); val == "" {
			t.Fatalf("%s must be set for acceptance tests", varName)
		}
	}

	for _, varName := range AccForAuthEnvVars {
		if val := os.Getenv(varName); val != "" {
			return
		}
	}
	t.Fatalf("one of the variables: %s must be set for acceptance tests", strings.Join(AccForAuthEnvVars, ", "))
}

func GetExampleFolderName() string {
	return folderName
}

func GetExampleCloudName() string {
	return cloudName
}

func GetExampleRoleID() string {
	return roleID
}

func GetExampleCloudID() string {
	return cloudID
}

func GetExampleOrganizationID() string {
	return organizationID
}

func GetExampleFolderID() string {
	return folderID
}

func GetExampleUserID1() string {
	return userID1
}

func GetExampleUserID2() string {
	return userID2
}

func GetExampleUserLogin1() string {
	return userLogin1
}

func GetExampleUserLogin2() string {
	return userLogin2
}

func GetExampleStorageEndpoint() string {
	return storageEndpoint
}

func GetBillingAccountId() string {
	return os.Getenv("YC_BILLING_TEST_ACCOUNT_ID_1")
}

func setTestIDs() error {
	// init sdk client based on env var
	envEndpoint := os.Getenv("YC_ENDPOINT")
	if envEndpoint == "" {
		envEndpoint = common.DefaultEndpoint
	}
	ctx := context.Background()

	providerConfig := &config.Config{
		ProviderState: config.State{
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
	cloudID = os.Getenv("YC_CLOUD_ID")
	organizationID = os.Getenv("YC_ORGANIZATION_ID")

	folderID = os.Getenv("YC_FOLDER_ID")
	folder := getFolderByID(sdk, folderID)
	if folder != nil {
		folderName = folder.Name
		if cloudID == "" {
			cloudID = folder.CloudId
		} else if cloudID != folder.CloudId {
			return fmt.Errorf("Invalid cloud id: %s != %s", cloudID, folder.CloudId)
		}
	} else {
		folderName = "no folder name detected"
	}
	cloudName = getCloudNameByID(sdk, cloudID)

	userLogin1 = os.Getenv("YC_LOGIN")
	userLogin2 = os.Getenv("YC_LOGIN_2")

	userID1 = loginToUserID(sdk, userLogin1)
	userID2 = loginToUserID(sdk, userLogin2)

	storageEndpoint = os.Getenv("YC_STORAGE_ENDPOINT_URL")

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
