package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	yds_conn "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/yds_connection"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydsConnectionStrategy struct {
}

func (_ *ydsConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	dataStreamsSetting := setting.GetDataStreams()
	d.Set(yds_conn.AttributeDatabaseID, dataStreamsSetting.GetDatabaseId())
	//d.Set(yds_conn.AttributeSharedReading, dataStreamsSetting.GetSharedReading())

	return flattenYandexYQAuth(d, dataStreamsSetting.GetAuth())
}

func (_ *ydsConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(AttributeServiceAccountID).(string)
	databaseID := d.Get(yds_conn.AttributeDatabaseID).(string)
	//	sharedReading := d.Get(yds_conn.AttributeSharedReading).(bool)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_DataStreams{
			DataStreams: &Ydb_FederatedQuery.DataStreams{
				DatabaseId: databaseID,
				Auth:       auth,
				// SharedReading: sharedReading,
			},
		},
	}, nil
}

func newYDSConnectionStrategy() ConnectionStrategy {
	return &ydsConnectionStrategy{}
}

func resourceYandexYQYDSConnection() *schema.Resource {
	return resourceYandexYQBaseConnection(newYDSConnectionStrategy(), yds_conn.ResourceSchema())
}
