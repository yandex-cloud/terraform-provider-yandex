package sweep

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/monitoring/v3"
	monitoringsdk "github.com/yandex-cloud/go-sdk/services/monitoring/v3"
	"google.golang.org/grpc/codes"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_monitoring "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/monitoring/pluginsdk"
)

func registerResourceYandexMonitoringDashboard(add func(string, *resource.Sweeper)) {
	add("yandex_monitoring_dashboard", &resource.Sweeper{
		Name: "yandex_monitoring_dashboard",
		F:    sweepAllMonitoringDashboard,
	})
}

func sweepAllMonitoringDashboard(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	ctx := context.Background()
	req := &monitoring.ListDashboardsRequest{
		Container: &monitoring.ListDashboardsRequest_FolderId{
			FolderId: conf.FolderID,
		},
	}
	resp, err := monitoringsdk.NewDashboardClient(conf.SDK).List(ctx, req)
	if err != nil {
		return fmt.Errorf("error getting monitoring dashboards: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Dashboards {
		if !sweepMonitoringDashboard(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep YDB monitoring dashboard %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMonitoringDashboard(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMonitoringDashboardOnce, conf, "Monitoring dashboard", id)
}

func sweepMonitoringDashboardOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_monitoring.YandexMonitoringDashboardDefaultTimeout)
	defer cancel()
	op, err := monitoringsdk.NewDashboardClient(conf.SDK).Delete(ctx, &monitoring.DeleteDashboardRequest{
		DashboardId: id,
	})
	if err != nil {
		if sdkcommon.IsStatusWithCode(err, codes.NotFound) {
			return nil
		}
		return err
	}
	_, err = op.Wait(ctx)
	return err
}
