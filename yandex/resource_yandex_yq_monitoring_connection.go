package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/monitoring_connection"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type monitoringConnectionStrategy struct {
}

func (_ *monitoringConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	monitoringSetting := setting.GetMonitoring()
	d.Set(monitoring_connection.AttributeProject, monitoringSetting.GetProject())
	d.Set(monitoring_connection.AttributeCluster, monitoringSetting.GetCluster())
	return flattenYandexYQAuth(d, monitoringSetting.GetAuth())
}

func (_ *monitoringConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(AttributeServiceAccountID).(string)
	project := d.Get(monitoring_connection.AttributeProject).(string)
	cluster := d.Get(monitoring_connection.AttributeCluster).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_Monitoring{
			Monitoring: &Ydb_FederatedQuery.Monitoring{
				Project: project,
				Cluster: cluster,
				Auth:    auth,
			},
		},
	}, nil
}

func newMonitoringConnectionStrategy() ConnectionStrategy {
	return &monitoringConnectionStrategy{}
}

func resourceYandexYQMonitoringConnection() *schema.Resource {
	return resourceYandexYQBaseConnection(newMonitoringConnectionStrategy(), monitoring_connection.ResourceSchema())
}
