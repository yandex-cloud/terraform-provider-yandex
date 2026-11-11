package sweep

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
	loggingsdk "github.com/yandex-cloud/go-sdk/services/logging/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
)

func registerResourceYandexLoggingGroup(add func(string, *resource.Sweeper)) {
	add("yandex_logging_group", &resource.Sweeper{
		Name: "yandex_logging_group",
		F:    sweepAllYandexLoggingGroup,
	})
}

func sweepAllYandexLoggingGroup(_ string) error {
	if os.Getenv("YC_SKIP_LOGGING_GROUP_SWEEPING") == "1" {
		log.Print("[INFO] Skipping yandex_logging_group sweeper: YC_SKIP_LOGGING_GROUP_SWEEPING=1")
		return nil
	}

	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	client := loggingsdk.NewLogGroupClient(conf.SDK)
	resp, err := client.List(conf.Context(), &logging.ListLogGroupsRequest{
		FolderId: conf.FolderID,
		PageSize: 1000,
	})
	if err != nil {
		return fmt.Errorf("error getting log group: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Groups {
		if !sweepYandexLoggingGroup(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Yandex Cloud Logging group %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepYandexLoggingGroup(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepYandexLoggingGroupOnce, conf, "Yandex Cloud Logging group", id)
}

func sweepYandexLoggingGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(30 * time.Second)
	defer cancel()

	client := loggingsdk.NewLogGroupClient(conf.SDK)
	op, err := client.Delete(ctx, &logging.DeleteLogGroupRequest{
		LogGroupId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}
