package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type monitoringConnectionStrategy struct {
}

func (*monitoringConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	monitoringSetting := setting.GetMonitoring()
	d.Set(AttributeProject, monitoringSetting.GetProject())
	d.Set(AttributeCluster, monitoringSetting.GetCluster())
	return flattenYandexYQAuth(d, monitoringSetting.GetAuth())
}

func (*monitoringConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(AttributeServiceAccountID).(string)
	project := d.Get(AttributeProject).(string)
	cluster := d.Get(AttributeCluster).(string)

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
	return resourceYandexYQBaseConnection(newMonitoringConnectionStrategy(), newMonitoringConnectionResourceSchema())
}
