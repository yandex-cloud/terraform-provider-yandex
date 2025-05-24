package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydsConnectionStrategy struct {
}

func (*ydsConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	dataStreamsSetting := setting.GetDataStreams()
	d.Set(AttributeDatabaseID, dataStreamsSetting.GetDatabaseId())
	d.Set(AttributeSharedReading, dataStreamsSetting.GetSharedReading())

	return flattenYandexYQAuth(d, dataStreamsSetting.GetAuth())
}

func (*ydsConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(AttributeServiceAccountID).(string)
	databaseID := d.Get(AttributeDatabaseID).(string)
	sharedReading := d.Get(AttributeSharedReading).(bool)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_DataStreams{
			DataStreams: &Ydb_FederatedQuery.DataStreams{
				DatabaseId:    databaseID,
				Auth:          auth,
				SharedReading: sharedReading,
			},
		},
	}, nil
}

func newYDSConnectionStrategy() ConnectionStrategy {
	return &ydsConnectionStrategy{}
}

func resourceYandexYQYDSConnection() *schema.Resource {
	description := "Yandex DataStreams connection. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#connection)."
	return resourceYandexYQBaseConnection(newYDSConnectionStrategy(), description, newYDSConnectionResourceSchema())
}
