package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydsBindingStrategy struct {
}

func (*ydsBindingStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.BindingSetting) error {
	ydsSetting := setting.GetDataStreams()

	d.Set(AttributeStream, ydsSetting.GetStreamName())
	d.Set(AttributeFormat, ydsSetting.GetFormat())
	d.Set(AttributeCompression, ydsSetting.GetCompression())
	d.Set(AttributeFormatSetting, ydsSetting.GetFormatSetting())

	schema, err := flattenSchema(ydsSetting.GetSchema())
	if err != nil {
		return err
	}

	d.Set(AttributeColumn, schema)
	return nil
}

func (*ydsBindingStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.BindingSetting, error) {
	format := d.Get(AttributeFormat).(string)
	compression := d.Get(AttributeCompression).(string)
	stream := d.Get(AttributeStream).(string)
	formatSetting, err := expandLabels(d.Get(AttributeFormatSetting))
	if err != nil {
		return nil, err
	}

	columns, err := parseColumns(d)
	if err != nil {
		return nil, err
	}

	schema := &Ydb_FederatedQuery.Schema{
		Column: columns,
	}

	return &Ydb_FederatedQuery.BindingSetting{
		Binding: &Ydb_FederatedQuery.BindingSetting_DataStreams{
			DataStreams: &Ydb_FederatedQuery.DataStreamsBinding{
				Format:        format,
				Compression:   compression,
				StreamName:    stream,
				Schema:        schema,
				FormatSetting: formatSetting,
			},
		},
	}, nil
}

func newYDSBindingStrategy() BindingStrategy {
	return &ydsBindingStrategy{}
}

func resourceYandexYQYDSBinding() *schema.Resource {
	return resourceYandexYQBaseBinding(newYDSBindingStrategy(), newYDSBindingResourceSchema())
}
