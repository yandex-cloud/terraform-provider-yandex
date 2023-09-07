package test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider-config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	defaultZoneForSweepers = "ru-central1-a"
)

type sweeperFunc func(*provider_config.Config, string) error

func configForSweepers() (*provider_config.Config, error) {
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
			StorageEndpoint:                types.StringValue(os.Getenv("YC_STORAGE_ENDPOINT_URL")),
		},
	}

	err = conf.InitAndValidate(context.Background(), "", true)
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func sweepWithRetry(sf sweeperFunc, conf *provider_config.Config, resource, id string) bool {
	return sweepWithRetryByFunc(conf, fmt.Sprintf("%s '%s'", resource, id), func(conf *provider_config.Config) error {
		return sf(conf, id)
	})
}

func sweepWithRetryByFunc(conf *provider_config.Config, message string, sf func(conf *provider_config.Config) error) bool {
	debugLog("started sweeping %s", message)
	for i := 1; i <= int(conf.ProviderState.MaxRetries.ValueInt64()); i++ {
		err := sf(conf)
		if err != nil {
			debugLog("[%s] delete try #%d: %v", message, i, err)
		} else {
			debugLog("[%s] delete try #%d: deleted", message, i)
			return true
		}
	}

	debugLog("failed to sweep %s", message)
	return false
}

func handleSweepOperation(ctx context.Context, conf *provider_config.Config, op *operation.Operation, err error) error {
	sdkop, err := conf.SDK.WrapOperation(op, err)
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

func isStatusWithCode(err error, code codes.Code) bool {
	grpcStatus, ok := status.FromError(err)
	return ok && grpcStatus.Code() == code
}
