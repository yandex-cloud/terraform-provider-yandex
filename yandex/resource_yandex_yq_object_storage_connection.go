package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type objectStorageConnectionStrategy struct {
}

func (*objectStorageConnectionStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.ConnectionSetting) error {
	objectStorageSetting := setting.GetObjectStorage()
	if objectStorageSetting == nil {
		return nil
	}

	d.Set(AttributeBucket, objectStorageSetting.GetBucket())
	return flattenYandexYQAuth(d, objectStorageSetting.GetAuth())
}

func (*objectStorageConnectionStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.ConnectionSetting, error) {
	serviceAccountID := d.Get(AttributeServiceAccountID).(string)
	bucket := d.Get(AttributeBucket).(string)
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

func newObjectStorageConnectionStrategy() ConnectionStrategy {
	return &objectStorageConnectionStrategy{}
}

func resourceYandexYQObjectStorageConnection() *schema.Resource {
	description := "Object Storage connection. For more information, see [the official documentation](https://yandex.cloud/docs/query/concepts/glossary#connection)."
	return resourceYandexYQBaseConnection(newObjectStorageConnectionStrategy(), description, newObjectStorageConnectionResourceSchema())
}
