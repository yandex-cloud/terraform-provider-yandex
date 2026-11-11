package sweep

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/common"

	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	"google.golang.org/grpc/codes"
)

func ConfigForSweepers() (*providerconfig.Config, error) {
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

	conf := &providerconfig.Config{
		Zone:                           zone,
		Insecure:                       insecure,
		MaxRetries:                     maxRetries,
		Token:                          token,
		ServiceAccountKeyFileOrContent: saKeyFile,
		CloudID:                        cloudID,
		FolderID:                       folderID,
		Endpoint:                       os.Getenv("YC_ENDPOINT"),
		StorageEndpoint:                os.Getenv("YC_STORAGE_ENDPOINT_URL"),
	}

	err = conf.InitAndValidate(context.Background(), "", true)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

const (
	defaultZoneForSweepers = "ru-central1-a"
	sweepRetryTimeout      = 5 * time.Second
)

func SweepWithRetry(sf sweeperFunc, conf *providerconfig.Config, resource, id string) bool {
	return SweepWithRetryByFunc(conf, fmt.Sprintf("%s '%s'", resource, id), func(conf *providerconfig.Config) error {
		return sf(conf, id)
	})
}

type sweeperFunc func(*providerconfig.Config, string) error

func SweepWithRetryByFunc(conf *providerconfig.Config, message string, sf func(conf *providerconfig.Config) error) bool {
	DebugLog("started sweeping %s", message)
	for i := 1; i <= conf.MaxRetries; i++ {
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

func DebugLog(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}

func HandleSweepOperationV2(ctx context.Context, op sdkcommon.SdkV2Operation, err error) error {
	if err != nil {
		if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}
	_, err = op.Abstract().Wait(ctx)
	return err
}
