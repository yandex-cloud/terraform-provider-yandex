package yandex

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type objectStorageBindingStrategy struct {
}

func (*objectStorageBindingStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.BindingSetting) error {
	objectStorageSetting := setting.GetObjectStorage()
	if len(objectStorageSetting.Subset) != 1 {
		return fmt.Errorf("unexpected empty subsets")
	}

	subset := objectStorageSetting.Subset[0]
	d.Set(AttributePathPattern, subset.GetPathPattern())
	d.Set(AttributeFormat, subset.GetFormat())
	d.Set(AttributeCompression, subset.GetCompression())
	d.Set(AttributeFormatSetting, subset.GetFormatSetting())
	d.Set(AttributePartitionedBy, subset.GetPartitionedBy())
	d.Set(AttributeProjection, subset.GetProjection())

	schema, err := flattenSchema(subset.GetSchema())
	if err != nil {
		return err
	}

	d.Set(AttributeColumn, schema)
	return nil
}

func (*objectStorageBindingStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.BindingSetting, error) {
	format := d.Get(AttributeFormat).(string)
	compression := d.Get(AttributeCompression).(string)
	pathPattern := d.Get(AttributePathPattern).(string)
	formatSetting, err := expandLabels(d.Get(AttributeFormatSetting))
	if err != nil {
		return nil, err
	}

	projection, err := expandLabels(d.Get(AttributeProjection))
	if err != nil {
		return nil, err
	}

	partitionedBy := expandStringList(d.Get(AttributePartitionedBy))
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
	return resourceYandexYQBaseBinding(newObjectStorageBindingStrategy(), newObjectStorageBindingResourceSchema())
}
