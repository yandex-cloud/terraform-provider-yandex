package sweep

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"google.golang.org/genproto/protobuf/field_mask"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/kafka/v1"
	kafkasdk "github.com/yandex-cloud/go-sdk/services/mdb/kafka/v1"
	sdkcommon "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sdkcommon"
	sweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/sweep"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex/providerconfig"
	svc_kafka "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/kafka/pluginsdk"
)

func registerResourceYandexMdbKafkaCluster(add func(string, *resource.Sweeper)) {
	add(kfResourceType, &resource.Sweeper{
		Name: kfResourceType,
		F:    sweepAllMDBKafkaCluster,
	})
}

const (
	currentDefaultKafkaVersion = "3.9"

	kfResourceType = "yandex_mdb_kafka_cluster"
	kfResourceFoo  = kfResourceType + ".foo"

	kfVPCDependencies = `
resource "yandex_vpc_network" "mdb-kafka-test-net" {}

resource "yandex_vpc_subnet" "mdb-kafka-test-subnet-a" {
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.mdb-kafka-test-net.id
  v4_cidr_blocks = ["10.1.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-kafka-test-subnet-b" {
  zone           = "ru-central1-b"
  network_id     = yandex_vpc_network.mdb-kafka-test-net.id
  v4_cidr_blocks = ["10.2.0.0/24"]
}

resource "yandex_vpc_subnet" "mdb-kafka-test-subnet-d" {
  zone           = "ru-central1-d"
  network_id     = yandex_vpc_network.mdb-kafka-test-net.id
  v4_cidr_blocks = ["10.3.0.0/24"]
}
`

	kfDiskEncryptionKeyResource = `
resource "yandex_kms_symmetric_key" "disk_encrypt" {}
`
)

func sweepAllMDBKafkaCluster(_ string) error {
	conf, err := sweep.ConfigForSweepers()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	resp, err := kafkasdk.NewClusterClient(conf.SDK).List(conf.Context(), &kafka.ListClustersRequest{
		FolderId: conf.FolderID,
		PageSize: sdkcommon.DefaultMDBPageSize,
	})
	if err != nil {
		return fmt.Errorf("error getting Kafka clusters: %s", err)
	}

	result := &multierror.Error{}
	for _, c := range resp.Clusters {
		if !sweepMDBKafkaCluster(conf, c.Id) {
			result = multierror.Append(result, fmt.Errorf("failed to sweep Kafka cluster %q", c.Id))
		}
	}

	return result.ErrorOrNil()
}

func sweepMDBKafkaCluster(conf *providerconfig.Config, id string) bool {
	return sweep.SweepWithRetry(sweepMDBKafkaClusterOnce, conf, "Kafka cluster", id)
}

func sweepMDBKafkaClusterOnce(conf *providerconfig.Config, id string) error {
	ctx, cancel := conf.ContextWithTimeout(svc_kafka.YandexMDBKafkaClusterDeleteTimeout)
	defer cancel()

	mask := field_mask.FieldMask{Paths: []string{"deletion_protection"}}
	op, err := kafkasdk.NewClusterClient(conf.SDK).Update(ctx, &kafka.UpdateClusterRequest{
		ClusterId:          id,
		DeletionProtection: false,
		UpdateMask:         &mask,
	})
	err = sweep.HandleSweepOperationV2(ctx, op, err)
	if err != nil && !strings.EqualFold(sdkcommon.ErrorMessage(err), "no changes detected") {
		return err
	}

	deleteOp, err := kafkasdk.NewClusterClient(conf.SDK).Delete(ctx, &kafka.DeleteClusterRequest{
		ClusterId: id,
	})
	return sweep.HandleSweepOperationV2(ctx, deleteOp, err)
}
