package yandex

import (
	"fmt"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

func init() {
	resource.AddTestSweepers("yandex_logging_group", &resource.Sweeper{
		Name: "yandex_logging_group",
		F:    testSweepYandexLoggingGroup,
	})
}

func testSweepYandexLoggingGroup(_ string) error {
	conf, err := configForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := conf.sdk.Logging().LogGroup().List(conf.Context(), &logging.ListLogGroupsRequest{
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

func sweepYandexLoggingGroup(conf *Config, id string) bool {
	return sweepWithRetry(sweepYandexLoggingGroupOnce, conf, "Yandex Cloud Logging group", id)
}

func sweepYandexLoggingGroupOnce(conf *Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(30 * time.Second)
	defer cancel()

	op, err := conf.sdk.Logging().LogGroup().Delete(ctx, &logging.DeleteLogGroupRequest{
		LogGroupId: id,
	})
	return handleSweepOperation(ctx, conf, op, err)
}
