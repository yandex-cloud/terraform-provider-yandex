package yandex_mdb_kafka_cluster_change_freeze

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/changefreeze"
)

var _ resource.ResourceWithModifyPlan = (*yandexMdbKafkaClusterChangeFreezeResource)(nil)

func (r *yandexMdbKafkaClusterChangeFreezeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	changefreeze.PreventActiveReplacement(ctx, req, resp)
}

func customChangeFreezeImporter(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	changefreeze.ImportState(ctx, req, resp)
}

func ignoreChangeFreezeTerminateError(err error) bool {
	return changefreeze.IgnoreTerminateError(err)
}
