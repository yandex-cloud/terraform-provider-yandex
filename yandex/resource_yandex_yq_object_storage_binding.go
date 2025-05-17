package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	os_binding "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/object_storage_binding"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type objectStorageBindingStrategy struct {
}

func (_ *objectStorageBindingStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.BindingSetting) error {
	objectStorageSetting := setting.GetObjectStorage()
	if len(objectStorageSetting.Subset) != 1 {
		return fmt.Errorf("unexpected empty subsets")
	}

	subset := objectStorageSetting.Subset[0]
	d.Set(os_binding.AttributePathPattern, subset.GetPathPattern())
	d.Set(os_binding.AttributeFormat, subset.GetFormat())
	d.Set(os_binding.AttributeCompression, subset.GetCompression())
	d.Set(os_binding.AttributeFormatSetting, subset.GetFormatSetting())
	d.Set(os_binding.AttributePartitionedBy, subset.GetPartitionedBy())
	d.Set(os_binding.AttributeProjection, subset.GetProjection())

	schema, err := flattenSchema(subset.GetSchema())
	if err != nil {
		return err
	}

	d.Set(os_binding.AttributeColumn, schema)
	return nil
}

func (_ *objectStorageBindingStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.BindingSetting, error) {
	format := d.Get(os_binding.AttributeFormat).(string)
	compression := d.Get(os_binding.AttributeCompression).(string)
	pathPattern := d.Get(os_binding.AttributePathPattern).(string)
	formatSetting, err := expandLabels(d.Get(os_binding.AttributeFormatSetting))
	if err != nil {
		return nil, err
	}

	projection, err := expandLabels(d.Get(os_binding.AttributeProjection))
	if err != nil {
		return nil, err
	}

	partitionedBy := expandStringList(d.Get(os_binding.AttributePartitionedBy))
	columns, err := parseColumns(d)
	if err != nil {
		return nil, err
	}

	schema := &Ydb_FederatedQuery.Schema{
		Column: columns,
	}

	return &Ydb_FederatedQuery.BindingSetting{
		Binding: &Ydb_FederatedQuery.BindingSetting_ObjectStorage{
			ObjectStorage: &Ydb_FederatedQuery.ObjectStorageBinding{
				Subset: []*Ydb_FederatedQuery.ObjectStorageBinding_Subset{
					{
						Format:        format,
						Compression:   compression,
						PathPattern:   pathPattern,
						Schema:        schema,
						FormatSetting: formatSetting,
						Projection:    projection,
						PartitionedBy: partitionedBy,
					},
				},
			},
		},
	}, nil
}

func newObjectStorageBindingStrategy() BindingStrategy {
	return &objectStorageBindingStrategy{}
}

func resourceYandexYQObjectStorageBinding() *schema.Resource {
	return resourceYandexYQBaseBinding(newObjectStorageBindingStrategy(), os_binding.ResourceSchema())
}
