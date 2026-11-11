package yandex

import (
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	alb "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/apploadbalancer/pluginsdk"
	audit_trails "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/audittrails/pluginsdk"
	backup "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/backup/pluginsdk"
	cdn "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/cdn/pluginsdk"
	certificate_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/certificatemanager/pluginsdk"
	compute "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/compute/pluginsdk"
	container_registry "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/containerregistry/pluginsdk"
	data_proc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/dataproc/pluginsdk"
	dns "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/dns/pluginsdk"
	iam "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iam/pluginsdk"
	iot_core "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/iot_core/pluginsdk"
	managed_kubernetes "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/k8s/pluginsdk"
	kms "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/kms/pluginsdk"
	network_load_balancer "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/loadbalancer/pluginsdk"
	lockbox "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/lockbox/pluginsdk"
	cloud_logging "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/logging/pluginsdk"
	clickhouse "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/clickhouse/pluginsdk"
	greenplum "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/greenplum/pluginsdk"
	kafka "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/kafka/pluginsdk"
	mongodb "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mongodb/pluginsdk"
	mysql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/mysql/pluginsdk"
	postgresql "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/postgresql/pluginsdk"
	redis "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/mdb/redis/pluginsdk"
	message_queue "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/message_queue/pluginsdk"
	monitoring "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/monitoring/pluginsdk"
	organization_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/organizationmanager/pluginsdk"
	resource_manager "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/resourcemanager/pluginsdk"
	api_gateway "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/apigateway/pluginsdk"
	serverless_containers "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/containers/pluginsdk"
	eventrouter "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/eventrouter/pluginsdk"
	cloud_functions "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/functions/pluginsdk"
	triggers "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/serverless/triggers/pluginsdk"
	smart_web_security "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/smartwebsecurity/pluginsdk"
	object_storage "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/storage/pluginsdk"
	vpc "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/vpc/pluginsdk"
	ydb "github.com/yandex-cloud/terraform-provider-yandex/yandex/services/ydb/pluginsdk"
)

type serviceRegistration struct {
	service     string
	resources   map[string]*schema.Resource
	dataSources map[string]*schema.Resource
}

func registerServices(provider *schema.Provider) error {
	return mergeRegistrations(provider, []serviceRegistration{
		{service: "apploadbalancer", resources: alb.Resources(), dataSources: alb.DataSources()},
		{service: "serverless/apigateway", resources: api_gateway.Resources(), dataSources: api_gateway.DataSources()},
		{service: "audittrails", resources: audit_trails.Resources(), dataSources: audit_trails.DataSources()},
		{service: "backup", resources: backup.Resources(), dataSources: backup.DataSources()},
		{service: "cdn", resources: cdn.Resources(), dataSources: cdn.DataSources()},
		{service: "certificatemanager", resources: certificate_manager.Resources(), dataSources: certificate_manager.DataSources()},
		{service: "mdb/clickhouse", resources: clickhouse.Resources(), dataSources: clickhouse.DataSources()},
		{service: "serverless/functions", resources: cloud_functions.Resources(), dataSources: cloud_functions.DataSources()},
		{service: "logging", resources: cloud_logging.Resources(), dataSources: cloud_logging.DataSources()},
		{service: "compute", resources: compute.Resources(), dataSources: compute.DataSources()},
		{service: "containerregistry", resources: container_registry.Resources(), dataSources: container_registry.DataSources()},
		{service: "dataproc", resources: data_proc.Resources(), dataSources: data_proc.DataSources()},
		{service: "dns", resources: dns.Resources(), dataSources: dns.DataSources()},
		{service: "serverless/eventrouter", resources: eventrouter.Resources(), dataSources: eventrouter.DataSources()},
		{service: "mdb/greenplum", resources: greenplum.Resources(), dataSources: greenplum.DataSources()},
		{service: "iam", resources: iam.Resources(), dataSources: iam.DataSources()},
		{service: "iot_core", resources: iot_core.Resources(), dataSources: iot_core.DataSources()},
		{service: "mdb/kafka", resources: kafka.Resources(), dataSources: kafka.DataSources()},
		{service: "kms", resources: kms.Resources(), dataSources: kms.DataSources()},
		{service: "lockbox", resources: lockbox.Resources(), dataSources: lockbox.DataSources()},
		{service: "k8s", resources: managed_kubernetes.Resources(), dataSources: managed_kubernetes.DataSources()},
		{service: "message_queue", resources: message_queue.Resources(), dataSources: message_queue.DataSources()},
		{service: "mdb/mongodb", resources: mongodb.Resources(), dataSources: mongodb.DataSources()},
		{service: "monitoring", resources: monitoring.Resources(), dataSources: monitoring.DataSources()},
		{service: "mdb/mysql", resources: mysql.Resources(), dataSources: mysql.DataSources()},
		{service: "loadbalancer", resources: network_load_balancer.Resources(), dataSources: network_load_balancer.DataSources()},
		{service: "storage", resources: object_storage.Resources(), dataSources: object_storage.DataSources()},
		{service: "organizationmanager", resources: organization_manager.Resources(), dataSources: organization_manager.DataSources()},
		{service: "mdb/postgresql", resources: postgresql.Resources(), dataSources: postgresql.DataSources()},
		{service: "mdb/redis", resources: redis.Resources(), dataSources: redis.DataSources()},
		{service: "resourcemanager", resources: resource_manager.Resources(), dataSources: resource_manager.DataSources()},
		{service: "serverless/containers", resources: serverless_containers.Resources(), dataSources: serverless_containers.DataSources()},
		{service: "smartwebsecurity", resources: smart_web_security.Resources(), dataSources: smart_web_security.DataSources()},
		{service: "serverless/triggers", resources: triggers.Resources(), dataSources: triggers.DataSources()},
		{service: "vpc", resources: vpc.Resources(), dataSources: vpc.DataSources()},
		{service: "ydb", resources: ydb.Resources(), dataSources: ydb.DataSources()},
	})
}

func mergeRegistrations(provider *schema.Provider, services []serviceRegistration) error {
	resources := make(map[string]*schema.Resource, len(provider.ResourcesMap))
	dataSources := make(map[string]*schema.Resource, len(provider.DataSourcesMap))
	maps.Copy(resources, provider.ResourcesMap)
	maps.Copy(dataSources, provider.DataSourcesMap)
	resourceOwners := providerEntryOwners(resources)
	dataSourceOwners := providerEntryOwners(dataSources)
	for _, service := range services {
		if err := mergeServiceEntries(resources, service.resources, resourceOwners, service.service, "resource"); err != nil {
			return fmt.Errorf("register service %q: %w", service.service, err)
		}
		if err := mergeServiceEntries(dataSources, service.dataSources, dataSourceOwners, service.service, "data source"); err != nil {
			return fmt.Errorf("register service %q: %w", service.service, err)
		}
	}
	provider.ResourcesMap = resources
	provider.DataSourcesMap = dataSources
	return nil
}

func providerEntryOwners(entries map[string]*schema.Resource) map[string]string {
	owners := make(map[string]string, len(entries))
	for name := range entries {
		owners[name] = "provider"
	}
	return owners
}

func mergeServiceEntries(target, entries map[string]*schema.Resource, owners map[string]string, owner, kind string) error {
	for name, entry := range entries {
		if name == "" || entry == nil {
			return fmt.Errorf("invalid %s %q", kind, name)
		}
		if previous, exists := owners[name]; exists {
			return fmt.Errorf("%s %q is already registered by %q", kind, name, previous)
		}
		target[name] = entry
		owners[name] = owner
	}
	return nil
}
