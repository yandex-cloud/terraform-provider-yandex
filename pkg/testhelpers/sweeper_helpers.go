package testhelpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/validate"
	provider_config "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"google.golang.org/grpc/codes"
)

const (
	defaultZoneForSweepers = "ru-central1-a"
	AccTestsUser           = "yc.terraform.acctest-sa"
	testResourceNamePrefix = "yc-tf-acc-tests"
	sweepRetryTimeout      = 5 * time.Second
)

type sweeperFunc func(*provider_config.Config, string) error

func ConfigForSweepers() (*provider_config.Config, error) {
	token, saKeyFile := os.Getenv("YC_TOKEN"), os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE")
	if token == "" && saKeyFile == "" {
		return nil, fmt.Errorf("environmental variables YC_TOKEN or YC_SERVICE_ACCOUNT_KEY_FILE must be set")
	}
	cloudID, folderID := os.Getenv("YC_CLOUD_ID"), os.Getenv("YC_FOLDER_ID")
	if folderID == "" {
		return nil, fmt.Errorf("environmental variable: YC_FOLDER_ID must be set")
	}

	insecure, err := strconv.ParseBool(strings.ToLower(os.Getenv("YC_INSECURE")))
	if err != nil {
		insecure = false
	}

	maxRetries, err := strconv.Atoi(os.Getenv("YC_MAX_RETRIES"))
	if err != nil {
		maxRetries = common.DefaultMaxRetries
	}

	zone := os.Getenv("YC_ZONE")
	if zone == "" {
		zone = defaultZoneForSweepers
	}

	conf := &provider_config.Config{
		ProviderState: provider_config.State{
			Zone:                           types.StringValue(zone),
			Insecure:                       types.BoolValue(insecure),
			MaxRetries:                     types.Int64Value(int64(maxRetries)),
			Token:                          types.StringValue(token),
			ServiceAccountKeyFileOrContent: types.StringValue(saKeyFile),
			CloudID:                        types.StringValue(cloudID),
			FolderID:                       types.StringValue(folderID),
			Endpoint:                       types.StringValue(os.Getenv("YC_ENDPOINT")),
			YQEndpoint:                     types.StringValue(common.DefaultYQEndpoint),
			StorageEndpoint:                types.StringValue(os.Getenv("YC_STORAGE_ENDPOINT_URL")),
		},
	}

	diags := diag.Diagnostics{}
	diags.Append(conf.InitAndValidate(context.Background(), "", true, diag.Diagnostics{})...)
	if diags.HasError() {
		return nil, fmt.Errorf(diags.Errors()[0].Detail())
	}

	return conf, nil
}

func SweepWithRetry(sf sweeperFunc, conf *provider_config.Config, resource, id string) bool {
	return SweepWithRetryByFunc(conf, fmt.Sprintf("%s '%s'", resource, id), func(conf *provider_config.Config) error {
		return sf(conf, id)
	})
}

func SweepWithRetryByFunc(conf *provider_config.Config, message string, sf func(conf *provider_config.Config) error) bool {
	DebugLog("started sweeping %s", message)
	for i := 1; i <= int(conf.ProviderState.MaxRetries.ValueInt64()); i++ {
		err := sf(conf)
		if err != nil {
			DebugLog("[%s] delete try #%d: %v", message, i, err)
		} else {
			DebugLog("[%s] delete try #%d: deleted", message, i)
			return true
		}
		time.Sleep(sweepRetryTimeout)
	}

	DebugLog("failed to sweep %s", message)
	return false
}

func HandleSweepOperation(ctx context.Context, conf *provider_config.Config, op *operation.Operation, err error) error {
	sdkop, err := conf.SDK.WrapOperation(op, err)
	if err != nil {
		if validate.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}

	err = sdkop.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = sdkop.Response()
	return err
}

func DebugLog(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

func IsTestResourceName(name string) bool {
	return strings.HasPrefix(name, testResourceNamePrefix)
}

// ResourceName - return randomized name for resource with common prefix
func ResourceName(length int) string {
	saltLen := length - len(testResourceNamePrefix) - 1
	return fmt.Sprintf(
		"%s-%s",
		testResourceNamePrefix,
		acctest.RandStringFromCharSet(saltLen, acctest.CharSetAlpha),
	)
}
