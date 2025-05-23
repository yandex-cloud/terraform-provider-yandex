package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydbConnectionStrategy struct {
}

func (*ydbConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	ydbSetting := setting.GetYdbDatabase()
	if ydbSetting == nil {
		return nil
	}

	d.Set(AttributeDatabaseID, ydbSetting.GetDatabaseId())

	return flattenYandexYQAuth(d, ydbSetting.GetAuth())
}

func (*ydbConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(AttributeServiceAccountID).(string)
	databaseID := d.Get(AttributeDatabaseID).(string)

	auth := parseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_YdbDatabase{
			YdbDatabase: &Ydb_FederatedQuery.YdbDatabase{
				DatabaseId: databaseID,
				Auth:       auth,
			},
		},
	}, nil
}

func newYDBConnectionStrategy() ConnectionStrategy {
	return &ydbConnectionStrategy{}
}

func resourceYandexYQYDBConnection() *schema.Resource {
	return resourceYandexYQBaseConnection(newYDBConnectionStrategy(), newYDBConnectionResourceSchema())
}
