package yandex

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/yandex-cloud/terraform-provider-yandex/common"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/grpc/codes"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
)

const (
	defaultZoneForSweepers = "ru-central1-a"
	sweepRetryTimeout      = 5 * time.Second
)

type sweeperFunc func(*Config, string) error

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func configForSweepers() (*Config, error) {
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

	conf := &Config{
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

	err = conf.initAndValidate(context.Background(), "", true)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func sweepWithRetry(sf sweeperFunc, conf *Config, resource, id string) bool {
	return sweepWithRetryByFunc(conf, fmt.Sprintf("%s '%s'", resource, id), func(conf *Config) error {
		return sf(conf, id)
	})
}

func sweepWithRetryByFunc(conf *Config, message string, sf func(conf *Config) error) bool {
	debugLog("started sweeping %s", message)
	for i := 1; i <= conf.MaxRetries; i++ {
		err := sf(conf)
		if err != nil {
			debugLog("[%s] delete try #%d: %v", message, i, err)
		} else {
			debugLog("[%s] delete try #%d: deleted", message, i)
			return true
		}
		time.Sleep(sweepRetryTimeout)
	}

	debugLog("failed to sweep %s", message)
	return false
}

func handleSweepOperation(ctx context.Context, conf *Config, op *operation.Operation, err error) error {
	sdkop, err := conf.sdk.WrapOperation(op, err)
	if err != nil {
		if isStatusWithCode(err, codes.NotFound) {
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

func debugLog(format string, v ...interface{}) {
	log.Printf("[DEBUG] "+format, v...)
}
