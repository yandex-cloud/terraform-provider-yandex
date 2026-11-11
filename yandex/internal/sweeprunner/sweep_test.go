package sweeprunner

import (
	"fmt"
	"log"
	"testing"

	cloud_registrysweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cloudregistry/sweep"
	connection_managersweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/connectionmanager/sweep"
	container_registrysweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/containerregistry/sweep"
	data_catalogsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datacatalog/sweep"
	dnssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/dns/sweep"
	workflowssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/workflows/sweep"
	ytsaurussweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/ytsaurus/sweep"

	billingsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/billing/sweep"
	datalenssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datalens/sweep"
	dataspheresweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datasphere/sweep"
	managed_gitlabsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/gitlab/sweep"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	airflowsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/airflow/sweep"
	albsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/apploadbalancer/sweep"
	audit_trailssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/audittrails/sweep"
	backupsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/backup/sweep"
	cdnsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cdn/sweep"
	certificate_managersweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/certificatemanager/sweep"
	computesweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/sweep"
	data_procsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/dataproc/sweep"
	datatransfersweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/datatransfer/sweep"
	iamsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/sweep"
	iot_coresweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iot_core/sweep"
	managed_kubernetessweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/k8s/sweep"
	kmssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/sweep"
	network_load_balancersweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/loadbalancer/sweep"
	lockboxsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/lockbox/sweep"
	cloud_loggingsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/logging/sweep"
	clickhousesweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/sweep"
	greenplumsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/greenplum/sweep"
	kafkasweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/kafka/sweep"
	mongodbsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mongodb/sweep"
	mysqlsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/sweep"
	opensearchsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/opensearch/sweep"
	postgresqlsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/postgresql/sweep"
	redissweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/redis/sweep"
	sharded_postgresqlsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/spqr/sweep"
	metastoresweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/metastore/sweep"
	monitoringsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/monitoring/sweep"
	organization_managersweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/sweep"
	resource_managersweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/sweep"
	api_gatewaysweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/apigateway/sweep"
	serverless_containerssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/containers/sweep"
	eventroutersweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/eventrouter/sweep"
	cloud_functionssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/functions/sweep"
	triggerssweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/triggers/sweep"
	smartcaptchasweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/smartcaptcha/sweep"
	smart_web_securitysweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/smartwebsecurity/sweep"
	sparksweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/spark/sweep"
	object_storagesweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/sweep"
	trinosweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/trino/sweep"
	vpcsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/sweep"
	ydbsweep "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/ydb/sweep"
)

func TestMain(m *testing.M) {
	sweepers := make(map[string]*resource.Sweeper)
	add := func(name string, sweeper *resource.Sweeper) {
		if sweeper == nil || sweeper.Name != name || sweeper.F == nil {
			log.Fatalf("invalid sweeper registration for %q", name)
		}
		if _, exists := sweepers[name]; exists {
			log.Fatalf("duplicate sweeper %q", name)
		}
		sweepers[name] = sweeper
	}

	ytsaurussweep.Register(add)
	workflowssweep.Register(add)
	dnssweep.Register(add)
	data_catalogsweep.Register(add)
	container_registrysweep.Register(add)
	connection_managersweep.Register(add)
	cloud_registrysweep.Register(add)
	managed_gitlabsweep.Register(add)
	dataspheresweep.Register(add)
	datalenssweep.Register(add)
	billingsweep.Register(add)
	airflowsweep.Register(add)
	albsweep.Register(add)
	api_gatewaysweep.Register(add)
	audit_trailssweep.Register(add)
	backupsweep.Register(add)
	cdnsweep.Register(add)
	certificate_managersweep.Register(add)
	clickhousesweep.Register(add)
	cloud_functionssweep.Register(add)
	cloud_loggingsweep.Register(add)
	computesweep.Register(add)
	data_procsweep.Register(add)
	datatransfersweep.Register(add)
	eventroutersweep.Register(add)
	greenplumsweep.Register(add)
	iamsweep.Register(add)
	iot_coresweep.Register(add)
	kafkasweep.Register(add)
	kmssweep.Register(add)
	lockboxsweep.Register(add)
	managed_kubernetessweep.Register(add)
	metastoresweep.Register(add)
	mongodbsweep.Register(add)
	monitoringsweep.Register(add)
	mysqlsweep.Register(add)
	network_load_balancersweep.Register(add)
	object_storagesweep.Register(add)
	opensearchsweep.Register(add)
	organization_managersweep.Register(add)
	postgresqlsweep.Register(add)
	redissweep.Register(add)
	resource_managersweep.Register(add)
	serverless_containerssweep.Register(add)
	sharded_postgresqlsweep.Register(add)
	smart_web_securitysweep.Register(add)
	smartcaptchasweep.Register(add)
	sparksweep.Register(add)
	triggerssweep.Register(add)
	trinosweep.Register(add)
	vpcsweep.Register(add)
	ydbsweep.Register(add)

	if err := validateDependencies(sweepers); err != nil {
		log.Fatal(err)
	}
	for name, sweeper := range sweepers {
		resource.AddTestSweepers(name, sweeper)
	}
	resource.TestMain(m)
}

func validateDependencies(sweepers map[string]*resource.Sweeper) error {
	const (
		visiting = iota + 1
		visited
	)
	states := make(map[string]int, len(sweepers))
	var visit func(string) error
	visit = func(name string) error {
		switch states[name] {
		case visiting:
			return fmt.Errorf("sweeper dependency cycle at %q", name)
		case visited:
			return nil
		}
		sweeper, exists := sweepers[name]
		if !exists {
			return fmt.Errorf("unknown sweeper %q", name)
		}
		states[name] = visiting
		for _, dependency := range sweeper.Dependencies {
			if err := visit(dependency); err != nil {
				return fmt.Errorf("sweeper %q: %w", name, err)
			}
		}
		states[name] = visited
		return nil
	}
	for name := range sweepers {
		if err := visit(name); err != nil {
			return err
		}
	}
	return nil
}
