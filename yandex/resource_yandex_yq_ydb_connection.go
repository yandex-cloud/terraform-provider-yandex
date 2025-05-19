package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/ydb_connection"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydbConnectionStrategy struct {
}

func (_ *ydbConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	ydbSetting := setting.GetYdbDatabase()
	d.Set(ydb_connection.AttributeDatabaseID, ydbSetting.GetDatabaseId())

	return flattenYandexYQAuth(d, ydbSetting.GetAuth())
}

func (_ *ydbConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(AttributeServiceAccountID).(string)
	databaseID := d.Get(ydb_connection.AttributeDatabaseID).(string)

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
	return resourceYandexYQBaseConnection(newYDBConnectionStrategy(), ydb_connection.ResourceSchema())
}
