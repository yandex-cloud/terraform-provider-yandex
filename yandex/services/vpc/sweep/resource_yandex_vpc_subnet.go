package sweep

import (
	"fmt"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
)

func sweepVPCSubnetOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCNetworkDefaultTimeout)
	defer cancel()

	client := vpcsdk.NewSubnetClient(conf.SDK)

	op, err := client.Delete(ctx, &vpc.DeleteSubnetRequest{
		SubnetId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}

func SweepVPCSubnet(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepVPCSubnetOnce, conf, "VPC Subnet", id)
}

// NOTE(dxan): function may return non-empty string and non-nil error. Example:
// Resource is successfully created, but wait fails: the function returns id and wait error
func CreateVPCSubnetForSweeper(conf *providerconfig.Config, networkID string) (string, error) {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCSubnetDefaultTimeout)
	defer cancel()
	client := vpcsdk.NewSubnetClient(conf.SDK)

	op, err := client.Create(ctx, &vpc.CreateSubnetRequest{
		Name:         acctest.RandomWithPrefix("sweeper"),
		Description:  "created by sweeper",
		ZoneId:       conf.Zone,
		FolderId:     conf.FolderID,
		NetworkId:    networkID,
		V4CidrBlocks: []string{"10.1.0.0/24"},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create subnet: %v", err)
	}

	md := op.Metadata()
	sweep.DebugLog("Subnet '%s' was created, waiting for complete operation '%s'", md.GetSubnetId(), op.ID())

	_, err = op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("error while waiting for create subnet operation: %v", err)
	}

	return md.SubnetId, nil
}

func registerResourceYandexVpcSubnet(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_subnet", &resource.Sweeper{
		Name: "yandex_vpc_subnet",
		F:    sweepAllVPCSubnets,
		Dependencies: []string{
			"yandex_airflow_cluster",
			"yandex_alb_load_balancer",
			"yandex_compute_instance",
			"yandex_compute_instance_group",
			"yandex_dataproc_cluster",
			"yandex_kubernetes_cluster",
			"yandex_kubernetes_node_group",
			"yandex_lb_target_group",
			"yandex_mdb_clickhouse_cluster",
			"yandex_mdb_greenplum_cluster",
			"yandex_mdb_kafka_cluster",
			"yandex_mdb_mongodb_cluster",
			"yandex_mdb_mysql_cluster",
			"yandex_mdb_opensearch_cluster",
			"yandex_mdb_postgresql_cluster",
			"yandex_mdb_postgresql_cluster_v2",
			"yandex_mdb_redis_cluster",
			"yandex_mdb_sharded_postgresql_cluster",
			"yandex_metastore_cluster",
			"yandex_spark_cluster",
			"yandex_trino_cluster",
			"yandex_vpc_private_endpoint",
			"yandex_vpc_security_group",
			"yandex_ydb_database_dedicated",
			"yandex_ydb_database_serverless",
		},
	})
}

func sweepAllVPCSubnets(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListSubnetsRequest{FolderId: conf.FolderID}
	client := vpcsdk.NewSubnetClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !SweepVPCSubnet(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC subnet %q", id))
		}
	}

	return result.ErrorOrNil()
}
