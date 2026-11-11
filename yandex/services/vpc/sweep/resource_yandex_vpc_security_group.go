package sweep

import (
	"fmt"

	"github.com/hashicorp/go-multierror"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	vpcsdk "github.com/yandex-cloud/go-sdk/services/vpc/v1"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
)

func sweepAllVPCSecurityGroups(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	req := &vpc.ListSecurityGroupsRequest{FolderId: conf.FolderID}
	client := vpcsdk.NewSecurityGroupClient(conf.SDK)

	it := client.Iterator(conf.Context(), req)
	result := &multierror.Error{}
	for it.Next() {
		id := it.Value().GetId()
		if !sweepVPCSecurityGroup(conf, id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep VPC security group %q", id))
		}
	}

	return result.ErrorOrNil()
}

func sweepVPCSecurityGroup(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepVPCSecurityGroupOnce, conf, "VPC Security Group", id)
}

func sweepVPCSecurityGroupOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_vpc.YandexVPCNetworkDefaultTimeout)
	defer cancel()
	client := vpcsdk.NewSecurityGroupClient(conf.SDK)

	sg, err := client.Get(ctx, &vpc.GetSecurityGroupRequest{
		SecurityGroupId: id,
	})
	if err != nil {
		return err
	}

	if sg.DefaultForNetwork {
		return nil
	}

	op, err := client.Delete(ctx, &vpc.DeleteSecurityGroupRequest{
		SecurityGroupId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, op, err)
}

func getYandexVPCSecurityGroupSweeperDeps() []string {
	return []string{
		"yandex_alb_load_balancer",
		"yandex_compute_instance",
		"yandex_compute_instance_group",
		"yandex_dataproc_cluster",
		"yandex_kubernetes_node_group",
		"yandex_kubernetes_cluster",
		"yandex_mdb_clickhouse_cluster",
		"yandex_mdb_mongodb_cluster",
		"yandex_mdb_mysql_cluster",
		"yandex_mdb_opensearch_cluster",
		"yandex_mdb_postgresql_cluster",
		"yandex_mdb_greenplum_cluster",
		"yandex_mdb_redis_cluster",
		"yandex_mdb_kafka_cluster",
		"yandex_mdb_kafka_cluster",
	}
}

func registerResourceYandexVpcSecurityGroup(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_security_group", &resource.Sweeper{
		Name:         "yandex_vpc_security_group",
		F:            sweepAllVPCSecurityGroups,
		Dependencies: getYandexVPCSecurityGroupSweeperDeps(),
	})
}
