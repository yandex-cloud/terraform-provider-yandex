package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	yds_binding "github.com/yandex-cloud/terraform-provider-yandex/yandex/internal/yq/sdk/yds_binding"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

type ydsBindingStrategy struct {
}

func (_ *ydsBindingStrategy) FlattenSetting(d *schema.ResourceData, setting *Ydb_FederatedQuery.BindingSetting) error {
	ydsSetting := setting.GetDataStreams()

	d.Set(yds_binding.AttributeStream, ydsSetting.GetStreamName())
	d.Set(yds_binding.AttributeFormat, ydsSetting.GetFormat())
	d.Set(yds_binding.AttributeCompression, ydsSetting.GetCompression())
	d.Set(yds_binding.AttributeFormatSetting, ydsSetting.GetFormatSetting())

	schema, err := flattenSchema(ydsSetting.GetSchema())
	if err != nil {
		return err
	}

	d.Set(yds_binding.AttributeColumn, schema)
	return nil
}

func (_ *ydsBindingStrategy) ExpandSetting(d *schema.ResourceData) (*Ydb_FederatedQuery.BindingSetting, error) {
	format := d.Get(yds_binding.AttributeFormat).(string)
	compression := d.Get(yds_binding.AttributeCompression).(string)
	stream := d.Get(yds_binding.AttributeStream).(string)
	formatSetting, err := expandLabels(d.Get(yds_binding.AttributeFormatSetting))
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
	return resourceYandexYQBaseBinding(newYDSBindingStrategy(), yds_binding.ResourceSchema())
}
