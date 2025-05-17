package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	os_conn "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/object_storage_connection"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type objectStorageConnectionStrategy struct {
}

func (_ *objectStorageConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	objectStorageSetting := setting.GetObjectStorage()
	d.Set(os_conn.AttributeBucket, objectStorageSetting.GetBucket())
	return flattenYandexYQAuth(d, objectStorageSetting.GetAuth())
}

func (_ *objectStorageConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(os_conn.AttributeServiceAccountID).(string)
	bucket := d.Get(os_conn.AttributeBucket).(string)
	auth := parseServiceIDToIAMAuth(serviceAccountID)
	return &Ydb_FederatedQuery.ConnectionSetting{
		Connection: &Ydb_FederatedQuery.ConnectionSetting_ObjectStorage{
			ObjectStorage: &Ydb_FederatedQuery.ObjectStorageConnection{
				Bucket: bucket,
				Auth:   auth,
			},
		},
	}, nil
}

func NewObjectStorageConnectionStrategy() ConnectionStrategy {
	return &objectStorageConnectionStrategy{}
}

func resourceYandexYQObjectStorageConnection() *schema.Resource {
	return resourceYandexYQBaseConnection(NewObjectStorageConnectionStrategy(), os_conn.ResourceSchema())
}
