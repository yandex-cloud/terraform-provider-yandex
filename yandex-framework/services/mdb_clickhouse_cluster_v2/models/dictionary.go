package models

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	clickhouseConfig "github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/clickhouse/v1/config"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datasize"
	"github.com/yandex-cloud/terraform-provider-yandex/pkg/mdbcommon"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_clickhouse_cluster_v2/utils"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type ExternalDictionary struct {
	Structure types.Object `tfsdk:"structure"`
	Layout    types.Object `tfsdk:"layout"`
	Lifetime  types.Object `tfsdk:"lifetime"`
	Source    types.Object `tfsdk:"source"`
}

var ExternalDictionaryAttrTypes = map[string]attr.Type{
	"structure": types.ObjectType{AttrTypes: DictionaryStructureAttrTypes},
	"layout":    types.ObjectType{AttrTypes: DictionaryLayoutAttrTypes},
	"lifetime":  types.ObjectType{AttrTypes: DictionaryLifetimeAttrTypes},
	"source":    types.ObjectType{AttrTypes: DictionarySourceAttrTypes},
}

func flattenExternalDictionary(ctx context.Context, apiDict *clickhouseConfig.ClickhouseConfig_ExternalDictionary, prevDict types.Object, diags *diag.Diagnostics) types.Object {
	if apiDict == nil {
		return types.ObjectNull(ExternalDictionaryAttrTypes)
	}

	if apiDict.Source == nil {
		diags.AddError(
			"Missing dictionary source",
			"External dictionary must have exactly one source defined",
		)
		return types.ObjectNull(ExternalDictionaryAttrTypes)
	}

	if apiDict.Lifetime == nil {
		diags.AddError(
			"Missing dictionary lifetime",
			"External dictionary must have exactly one lifetime option defined",
		)
		return types.ObjectNull(ExternalDictionaryAttrTypes)
	}

	lifetime := DictionaryLifetime{
		Range: types.ObjectNull(DictionaryLifetimeRangeAttrTypes),
	}
	switch lt := apiDict.Lifetime.(type) {
	case *clickhouseConfig.ClickhouseConfig_ExternalDictionary_FixedLifetime:
		lifetime.FixedLifetime = types.Int64Value(lt.FixedLifetime)
	case *clickhouseConfig.ClickhouseConfig_ExternalDictionary_LifetimeRange:
		lifetime.Range = flattenDictionaryLifetimeRange(ctx, lt.LifetimeRange, diags)
	default:
		diags.AddError(
			"Invalid lifetime type",
			fmt.Sprintf("Unknown lifetime type: %T", apiDict.Lifetime),
		)
		return types.ObjectNull(ExternalDictionaryAttrTypes)
	}

	lifetimeObj, d := types.ObjectValueFrom(ctx, DictionaryLifetimeAttrTypes, lifetime)
	diags.Append(d...)

	prevSrc := DictionarySource{
		HttpSource:       types.ObjectNull(DictionaryHttpSourceAttrTypes),
		ClickhouseSource: types.ObjectNull(DictionaryClickhouseSourceAttrTypes),
		MongodbSource:    types.ObjectNull(DictionaryMongodbSourceAttrTypes),
		PostgresqlSource: types.ObjectNull(DictionaryPostgresqlSourceAttrTypes),
		MysqlSource:      types.ObjectNull(DictionaryMysqlSourceAttrTypes),
	}
	if !prevDict.IsNull() && !prevDict.IsUnknown() {
		var prev ExternalDictionary
		diags.Append(prevDict.As(ctx, &prev, datasize.DefaultOpts)...)
		if !prev.Source.IsNull() && !prev.Source.IsUnknown() {
			diags.Append(prev.Source.As(ctx, &prevSrc, datasize.DefaultOpts)...)
		}
	}

	source := DictionarySource{
		HttpSource:       types.ObjectNull(DictionaryHttpSourceAttrTypes),
		ClickhouseSource: types.ObjectNull(DictionaryClickhouseSourceAttrTypes),
		MongodbSource:    types.ObjectNull(DictionaryMongodbSourceAttrTypes),
		PostgresqlSource: types.ObjectNull(DictionaryPostgresqlSourceAttrTypes),
		MysqlSource:      types.ObjectNull(DictionaryMysqlSourceAttrTypes),
	}

	switch src := apiDict.Source.(type) {
	case *clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_:
		source.HttpSource = flattenDictionaryHttpSource(ctx, src.HttpSource, diags)
	case *clickhouseConfig.ClickhouseConfig_ExternalDictionary_ClickhouseSource_:
		source.ClickhouseSource = flattenDictionaryClickHouseSource(ctx, src.ClickhouseSource, prevSrc.ClickhouseSource, diags)
	case *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MongodbSource_:
		source.MongodbSource = flattenDictionaryMongoDbSource(ctx, src.MongodbSource, prevSrc.MongodbSource, diags)
	case *clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSource_:
		source.PostgresqlSource = flattenDictionaryPostgreSqlSource(ctx, src.PostgresqlSource, prevSrc.PostgresqlSource, diags)
	case *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_:
		source.MysqlSource = flattenDictionaryMySqlSource(ctx, src.MysqlSource, prevSrc.MysqlSource, diags)
	default:
		diags.AddError(
			"Unsupported source type",
			fmt.Sprintf("Source type %T is not supported yet", apiDict.Source),
		)
		return types.ObjectNull(ExternalDictionaryAttrTypes)
	}

	sourceObj, d := types.ObjectValueFrom(ctx, DictionarySourceAttrTypes, source)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(
		ctx, ExternalDictionaryAttrTypes, ExternalDictionary{
			Structure: flattenDictionaryStructure(ctx, apiDict.Structure, diags),
			Layout:    flattenDictionaryLayout(ctx, apiDict.Layout, diags),
			Lifetime:  lifetimeObj,
			Source:    sourceObj,
		},
	)
	diags.Append(d...)

	return obj
}

func FlattenExternalDictionaries(ctx context.Context, dicts []*clickhouseConfig.ClickhouseConfig_ExternalDictionary, prevDicts types.Map, diags *diag.Diagnostics) types.Map {
	if dicts == nil {
		return types.MapNull(types.ObjectType{AttrTypes: ExternalDictionaryAttrTypes})
	}

	prevElems := prevDicts.Elements()
	tfDicts := make(map[string]attr.Value, len(dicts))
	for _, d := range dicts {
		prevDictObj := types.ObjectNull(ExternalDictionaryAttrTypes)
		if prev, ok := prevElems[d.Name]; ok {
			prevDictObj = prev.(types.Object)
		}
		tfDicts[d.Name] = flattenExternalDictionary(ctx, d, prevDictObj, diags)
	}

	m, d := types.MapValue(types.ObjectType{AttrTypes: ExternalDictionaryAttrTypes}, tfDicts)
	diags.Append(d...)

	return m
}

func ExpandExternalDictionaries(ctx context.Context, c types.Map, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_ExternalDictionary {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var dictionaries map[string]ExternalDictionary
	diags.Append(c.ElementsAs(ctx, &dictionaries, false)...)
	if diags.HasError() {
		return nil
	}

	names := make([]string, 0, len(dictionaries))
	for name := range dictionaries {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]*clickhouseConfig.ClickhouseConfig_ExternalDictionary, 0, len(dictionaries))
	for _, name := range names {
		dictionary := dictionaries[name]
		proto := &clickhouseConfig.ClickhouseConfig_ExternalDictionary{
			Name:      name,
			Structure: expandDictionaryStructure(ctx, dictionary.Structure, diags),
			Layout:    expandDictionaryLayout(ctx, dictionary.Layout, diags),
		}
		expandDictionaryLifetime(ctx, dictionary.Lifetime, proto, diags)
		expandDictionarySource(ctx, dictionary.Source, proto, diags)
		result = append(result, proto)
	}

	return result
}

type DictionaryLifetime struct {
	FixedLifetime types.Int64  `tfsdk:"fixed_lifetime"`
	Range         types.Object `tfsdk:"range"`
}

var DictionaryLifetimeAttrTypes = map[string]attr.Type{
	"fixed_lifetime": types.Int64Type,
	"range":          types.ObjectType{AttrTypes: DictionaryLifetimeRangeAttrTypes},
}

type DictionarySource struct {
	HttpSource       types.Object `tfsdk:"http_source"`
	ClickhouseSource types.Object `tfsdk:"clickhouse_source"`
	MongodbSource    types.Object `tfsdk:"mongodb_source"`
	PostgresqlSource types.Object `tfsdk:"postgresql_source"`
	MysqlSource      types.Object `tfsdk:"mysql_source"`
}

var DictionarySourceAttrTypes = map[string]attr.Type{
	"http_source":       types.ObjectType{AttrTypes: DictionaryHttpSourceAttrTypes},
	"clickhouse_source": types.ObjectType{AttrTypes: DictionaryClickhouseSourceAttrTypes},
	"mongodb_source":    types.ObjectType{AttrTypes: DictionaryMongodbSourceAttrTypes},
	"postgresql_source": types.ObjectType{AttrTypes: DictionaryPostgresqlSourceAttrTypes},
	"mysql_source":      types.ObjectType{AttrTypes: DictionaryMysqlSourceAttrTypes},
}

type DictionaryStructure struct {
	Id         types.Object `tfsdk:"id"`
	Key        types.Object `tfsdk:"key"`
	RangeMin   types.Object `tfsdk:"range_min"`
	RangeMax   types.Object `tfsdk:"range_max"`
	Attributes types.List   `tfsdk:"attributes"`
}

var DictionaryStructureAttrTypes = map[string]attr.Type{
	"id":         types.ObjectType{AttrTypes: DictionaryIdAttrTypes},
	"key":        types.ObjectType{AttrTypes: DictionaryKeyAttrTypes},
	"range_min":  types.ObjectType{AttrTypes: DictionaryAttributeAttrTypes},
	"range_max":  types.ObjectType{AttrTypes: DictionaryAttributeAttrTypes},
	"attributes": types.ListType{ElemType: types.ObjectType{AttrTypes: DictionaryAttributeAttrTypes}},
}

func flattenDictionaryStructure(ctx context.Context, apiStruct *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure, diags *diag.Diagnostics) types.Object {
	if apiStruct == nil {
		return types.ObjectNull(DictionaryStructureAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryStructureAttrTypes, DictionaryStructure{
			Id:         flattenDictionaryId(ctx, apiStruct.Id, diags),
			Key:        flattenDictionaryKey(ctx, apiStruct.Key, diags),
			RangeMin:   flattenDictionaryAttribute(ctx, apiStruct.RangeMin, diags),
			RangeMax:   flattenDictionaryAttribute(ctx, apiStruct.RangeMax, diags),
			Attributes: flattenDictionaryAttributes(ctx, apiStruct.Attributes, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryStructure(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var structure DictionaryStructure
	diags.Append(c.As(ctx, &structure, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure{
		Id:         expandDictionaryId(ctx, structure.Id, diags),
		Key:        expandDictionaryKey(ctx, structure.Key, diags),
		RangeMin:   expandDictionaryAttribute(ctx, structure.RangeMin, diags),
		RangeMax:   expandDictionaryAttribute(ctx, structure.RangeMax, diags),
		Attributes: expandDictionaryAttributes(ctx, structure.Attributes, diags),
	}
}

type DictionaryId struct {
	Name types.String `tfsdk:"name"`
}

var DictionaryIdAttrTypes = map[string]attr.Type{
	"name": types.StringType,
}

func flattenDictionaryId(ctx context.Context, dictId *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Id, diags *diag.Diagnostics) types.Object {
	if dictId == nil {
		return types.ObjectNull(DictionaryIdAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryIdAttrTypes, DictionaryId{
			Name: types.StringValue(dictId.Name),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryId(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Id {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var id DictionaryId
	diags.Append(c.As(ctx, &id, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Id{
		Name: id.Name.ValueString(),
	}
}

type DictionaryKey struct {
	Attributes types.List `tfsdk:"attributes"`
}

var DictionaryKeyAttrTypes = map[string]attr.Type{
	"attributes": types.ListType{ElemType: types.ObjectType{AttrTypes: DictionaryAttributeAttrTypes}},
}

func flattenDictionaryKey(ctx context.Context, dictKey *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Key, diags *diag.Diagnostics) types.Object {
	if dictKey == nil {
		return types.ObjectNull(DictionaryKeyAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryKeyAttrTypes, DictionaryKey{
			Attributes: flattenDictionaryAttributes(ctx, dictKey.Attributes, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryKey(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Key {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var key DictionaryKey
	diags.Append(c.As(ctx, &key, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Key{
		Attributes: expandDictionaryAttributes(ctx, key.Attributes, diags),
	}
}

type DictionaryAttribute struct {
	Name         types.String `tfsdk:"name"`
	Type         types.String `tfsdk:"type"`
	NullValue    types.String `tfsdk:"null_value"`
	Expression   types.String `tfsdk:"expression"`
	Hierarchical types.Bool   `tfsdk:"hierarchical"`
	Injective    types.Bool   `tfsdk:"injective"`
}

var DictionaryAttributeAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"type":         types.StringType,
	"null_value":   types.StringType,
	"expression":   types.StringType,
	"hierarchical": types.BoolType,
	"injective":    types.BoolType,
}

func flattenDictionaryAttribute(ctx context.Context, dictAttr *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Attribute, diags *diag.Diagnostics) types.Object {
	if dictAttr == nil {
		return types.ObjectNull(DictionaryAttributeAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryAttributeAttrTypes, DictionaryAttribute{
			Name:         types.StringValue(dictAttr.Name),
			Type:         types.StringValue(dictAttr.Type),
			NullValue:    types.StringValue(dictAttr.NullValue),
			Expression:   types.StringValue(dictAttr.Expression),
			Hierarchical: types.BoolValue(dictAttr.Hierarchical),
			Injective:    types.BoolValue(dictAttr.Injective),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryAttribute(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Attribute {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var attribute DictionaryAttribute
	diags.Append(c.As(ctx, &attribute, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Attribute{
		Name:         attribute.Name.ValueString(),
		Type:         attribute.Type.ValueString(),
		NullValue:    attribute.NullValue.ValueString(),
		Expression:   attribute.Expression.ValueString(),
		Hierarchical: attribute.Hierarchical.ValueBool(),
		Injective:    attribute.Injective.ValueBool(),
	}
}

func flattenDictionaryAttributes(ctx context.Context, dictAttrs []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Attribute, diags *diag.Diagnostics) types.List {
	if dictAttrs == nil {
		return types.ListNull(types.ObjectType{AttrTypes: DictionaryAttributeAttrTypes})
	}

	tfAttrs := make([]types.Object, len(dictAttrs))
	for i, attr := range dictAttrs {
		tfAttrs[i] = flattenDictionaryAttribute(ctx, attr, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DictionaryAttributeAttrTypes}, tfAttrs)
	diags.Append(d...)

	return list
}

func expandDictionaryAttributes(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Attribute {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	elems := c.Elements()
	result := make([]*clickhouseConfig.ClickhouseConfig_ExternalDictionary_Structure_Attribute, 0, len(elems))
	for _, e := range elems {
		result = append(result, expandDictionaryAttribute(ctx, e.(types.Object), diags))
		if diags.HasError() {
			return nil
		}
	}

	return result
}

type DictionaryLayout struct {
	Type                               types.String `tfsdk:"type"`
	SizeInCells                        types.Int64  `tfsdk:"size_in_cells"`
	AllowReadExpiredKeys               types.Bool   `tfsdk:"allow_read_expired_keys"`
	MaxUpdateQueueSize                 types.Int64  `tfsdk:"max_update_queue_size"`
	UpdateQueuePushTimeoutMilliseconds types.Int64  `tfsdk:"update_queue_push_timeout_milliseconds"`
	QueryWaitTimeoutMilliseconds       types.Int64  `tfsdk:"query_wait_timeout_milliseconds"`
	MaxThreadsForUpdates               types.Int64  `tfsdk:"max_threads_for_updates"`
	InitialArraySize                   types.Int64  `tfsdk:"initial_array_size"`
	MaxArraySize                       types.Int64  `tfsdk:"max_array_size"`
	AccessToKeyFromAttributes          types.Bool   `tfsdk:"access_to_key_from_attributes"`
	BlockSize                          types.Int64  `tfsdk:"block_size"`
	FileSize                           types.Int64  `tfsdk:"file_size"`
	ReadBufferSize                     types.Int64  `tfsdk:"read_buffer_size"`
	WriteBufferSize                    types.Int64  `tfsdk:"write_buffer_size"`
}

var DictionaryLayoutAttrTypes = map[string]attr.Type{
	"type":                                   types.StringType,
	"size_in_cells":                          types.Int64Type,
	"allow_read_expired_keys":                types.BoolType,
	"max_update_queue_size":                  types.Int64Type,
	"update_queue_push_timeout_milliseconds": types.Int64Type,
	"query_wait_timeout_milliseconds":        types.Int64Type,
	"max_threads_for_updates":                types.Int64Type,
	"initial_array_size":                     types.Int64Type,
	"max_array_size":                         types.Int64Type,
	"access_to_key_from_attributes":          types.BoolType,
	"block_size":                             types.Int64Type,
	"file_size":                              types.Int64Type,
	"read_buffer_size":                       types.Int64Type,
	"write_buffer_size":                      types.Int64Type,
}

func flattenDictionaryLayout(ctx context.Context, layout *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Layout, diags *diag.Diagnostics) types.Object {
	if layout == nil {
		return types.ObjectNull(DictionaryLayoutAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryLayoutAttrTypes, DictionaryLayout{
			Type:                               types.StringValue(layout.Type.Enum().String()),
			SizeInCells:                        types.Int64Value(layout.SizeInCells),
			AllowReadExpiredKeys:               flattenBoolWrapperOrFalse(layout.AllowReadExpiredKeys),
			MaxUpdateQueueSize:                 types.Int64Value(layout.MaxUpdateQueueSize),
			UpdateQueuePushTimeoutMilliseconds: types.Int64Value(layout.UpdateQueuePushTimeoutMilliseconds),
			QueryWaitTimeoutMilliseconds:       types.Int64Value(layout.QueryWaitTimeoutMilliseconds),
			MaxThreadsForUpdates:               types.Int64Value(layout.MaxThreadsForUpdates),
			InitialArraySize:                   types.Int64Value(layout.InitialArraySize),
			MaxArraySize:                       types.Int64Value(layout.MaxArraySize),
			AccessToKeyFromAttributes:          flattenBoolWrapperOrFalse(layout.AccessToKeyFromAttributes),
			BlockSize:                          types.Int64Value(layout.BlockSize),
			FileSize:                           types.Int64Value(layout.FileSize),
			ReadBufferSize:                     types.Int64Value(layout.ReadBufferSize),
			WriteBufferSize:                    types.Int64Value(layout.WriteBufferSize),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryLayout(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Layout {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var layout DictionaryLayout
	diags.Append(c.As(ctx, &layout, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	layoutType := utils.ExpandEnum("type", layout.Type.ValueString(), clickhouseConfig.ClickhouseConfig_ExternalDictionary_Layout_Type_value, diags)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Layout{
		Type:                               clickhouseConfig.ClickhouseConfig_ExternalDictionary_Layout_Type(*layoutType),
		SizeInCells:                        layout.SizeInCells.ValueInt64(),
		AllowReadExpiredKeys:               mdbcommon.ExpandBoolWrapper(ctx, layout.AllowReadExpiredKeys, diags),
		MaxUpdateQueueSize:                 layout.MaxUpdateQueueSize.ValueInt64(),
		UpdateQueuePushTimeoutMilliseconds: layout.UpdateQueuePushTimeoutMilliseconds.ValueInt64(),
		QueryWaitTimeoutMilliseconds:       layout.QueryWaitTimeoutMilliseconds.ValueInt64(),
		MaxThreadsForUpdates:               layout.MaxThreadsForUpdates.ValueInt64(),
		InitialArraySize:                   layout.InitialArraySize.ValueInt64(),
		MaxArraySize:                       layout.MaxArraySize.ValueInt64(),
		AccessToKeyFromAttributes:          mdbcommon.ExpandBoolWrapper(ctx, layout.AccessToKeyFromAttributes, diags),
		BlockSize:                          layout.BlockSize.ValueInt64(),
		FileSize:                           layout.FileSize.ValueInt64(),
		ReadBufferSize:                     layout.ReadBufferSize.ValueInt64(),
		WriteBufferSize:                    layout.WriteBufferSize.ValueInt64(),
	}
}

func expandDictionaryLifetime(ctx context.Context, c types.Object, proto *clickhouseConfig.ClickhouseConfig_ExternalDictionary, diags *diag.Diagnostics) {
	if c.IsNull() || c.IsUnknown() {
		return
	}

	var lifetime DictionaryLifetime
	diags.Append(c.As(ctx, &lifetime, datasize.DefaultOpts)...)
	if diags.HasError() {
		return
	}

	if !lifetime.FixedLifetime.IsNull() && !lifetime.FixedLifetime.IsUnknown() {
		proto.Lifetime = &clickhouseConfig.ClickhouseConfig_ExternalDictionary_FixedLifetime{
			FixedLifetime: lifetime.FixedLifetime.ValueInt64(),
		}
		return
	}

	if !lifetime.Range.IsNull() && !lifetime.Range.IsUnknown() {
		proto.Lifetime = expandDictionaryLifetimeRange(ctx, lifetime.Range, diags)
		return
	}

	diags.AddError(
		"Invalid dictionary lifetime",
		"Either fixed_lifetime or range must be set in lifetime block",
	)
}

func expandDictionarySource(ctx context.Context, c types.Object, proto *clickhouseConfig.ClickhouseConfig_ExternalDictionary, diags *diag.Diagnostics) {
	if c.IsNull() || c.IsUnknown() {
		diags.AddError(
			"Invalid dictionary source",
			"External dictionary must have exactly one source defined",
		)
		return
	}

	var source DictionarySource
	diags.Append(c.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return
	}

	if !source.HttpSource.IsNull() && !source.HttpSource.IsUnknown() {
		proto.Source = &clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_{
			HttpSource: expandDictionaryHttpSource(ctx, source.HttpSource, diags),
		}
		return
	}

	if !source.ClickhouseSource.IsNull() && !source.ClickhouseSource.IsUnknown() {
		proto.Source = &clickhouseConfig.ClickhouseConfig_ExternalDictionary_ClickhouseSource_{
			ClickhouseSource: expandDictionaryClickHouseSource(ctx, source.ClickhouseSource, diags),
		}
		return
	}

	if !source.MongodbSource.IsNull() && !source.MongodbSource.IsUnknown() {
		proto.Source = &clickhouseConfig.ClickhouseConfig_ExternalDictionary_MongodbSource_{
			MongodbSource: expandDictionaryMongoDbSource(ctx, source.MongodbSource, diags),
		}
		return
	}

	if !source.PostgresqlSource.IsNull() && !source.PostgresqlSource.IsUnknown() {
		proto.Source = &clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSource_{
			PostgresqlSource: expandDictionaryPostgreSqlSource(ctx, source.PostgresqlSource, diags),
		}
		return
	}

	if !source.MysqlSource.IsNull() && !source.MysqlSource.IsUnknown() {
		proto.Source = &clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_{
			MysqlSource: expandDictionaryMySqlSource(ctx, source.MysqlSource, diags),
		}
		return
	}

	diags.AddError(
		"Invalid dictionary source",
		"External dictionary must have exactly one source defined (http_source, clickhouse_source, mongodb_source, postgresql_source, mysql_source)",
	)
}

type DictionaryLifetimeRange struct {
	Min types.Int64 `tfsdk:"min"`
	Max types.Int64 `tfsdk:"max"`
}

var DictionaryLifetimeRangeAttrTypes = map[string]attr.Type{
	"min": types.Int64Type,
	"max": types.Int64Type,
}

func flattenDictionaryLifetimeRange(ctx context.Context, lifetimeRange *clickhouseConfig.ClickhouseConfig_ExternalDictionary_Range, diags *diag.Diagnostics) types.Object {
	if lifetimeRange == nil {
		return types.ObjectNull(DictionaryLifetimeRangeAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryLifetimeRangeAttrTypes, DictionaryLifetimeRange{
			Min: types.Int64Value(lifetimeRange.Min),
			Max: types.Int64Value(lifetimeRange.Max),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryLifetimeRange(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_LifetimeRange {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var lifetimeRange DictionaryLifetimeRange
	diags.Append(c.As(ctx, &lifetimeRange, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_LifetimeRange{
		LifetimeRange: &clickhouseConfig.ClickhouseConfig_ExternalDictionary_Range{
			Min: lifetimeRange.Min.ValueInt64(),
			Max: lifetimeRange.Max.ValueInt64(),
		},
	}
}

type DictionaryHttpSource struct {
	Url     types.String `tfsdk:"url"`
	Format  types.String `tfsdk:"format"`
	Headers types.List   `tfsdk:"headers"`
}

var DictionaryHttpSourceAttrTypes = map[string]attr.Type{
	"url":     types.StringType,
	"format":  types.StringType,
	"headers": types.ListType{ElemType: types.ObjectType{AttrTypes: DictionaryHttpHeaderAttrTypes}},
}

func flattenDictionaryHttpSource(ctx context.Context, source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource, diags *diag.Diagnostics) types.Object {
	if source == nil {
		return types.ObjectNull(DictionaryHttpSourceAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryHttpSourceAttrTypes, DictionaryHttpSource{
			Url:     types.StringValue(source.Url),
			Format:  types.StringValue(source.Format),
			Headers: flattenListDictionaryHttpHeader(ctx, source.Headers, diags),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryHttpSource(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var source DictionaryHttpSource
	diags.Append(c.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource{
		Url:     source.Url.ValueString(),
		Format:  source.Format.ValueString(),
		Headers: expandDictionaryHttpHeaders(ctx, source.Headers, diags),
	}
}

type DictionaryHttpHeader struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

var DictionaryHttpHeaderAttrTypes = map[string]attr.Type{
	"name":  types.StringType,
	"value": types.StringType,
}

func flattenDictionaryHttpHeader(ctx context.Context, header *clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_Header, diags *diag.Diagnostics) types.Object {
	if header == nil {
		return types.ObjectNull(DictionaryHttpHeaderAttrTypes)
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryHttpHeaderAttrTypes, DictionaryHttpHeader{
			Name:  types.StringValue(header.Name),
			Value: types.StringValue(header.Value),
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListDictionaryHttpHeader(ctx context.Context, headers []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_Header, diags *diag.Diagnostics) types.List {
	if len(headers) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: DictionaryHttpHeaderAttrTypes})
	}

	tfHttpHeaders := make([]types.Object, len(headers))
	for i, header := range headers {
		tfHttpHeaders[i] = flattenDictionaryHttpHeader(ctx, header, diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DictionaryHttpHeaderAttrTypes}, tfHttpHeaders)
	diags.Append(d...)

	return list
}

func expandDictionaryHttpHeader(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_Header {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var header DictionaryHttpHeader
	diags.Append(c.As(ctx, &header, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_Header{
		Name:  header.Name.ValueString(),
		Value: header.Value.ValueString(),
	}
}

func expandDictionaryHttpHeaders(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_Header {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	elems := c.Elements()
	result := make([]*clickhouseConfig.ClickhouseConfig_ExternalDictionary_HttpSource_Header, 0, len(elems))
	for _, e := range elems {
		result = append(result, expandDictionaryHttpHeader(ctx, e.(types.Object), diags))
		if diags.HasError() {
			return nil
		}
	}

	return result
}

type DictionaryMysqlSource struct {
	Db              types.String `tfsdk:"db"`
	Table           types.String `tfsdk:"table"`
	Port            types.Int64  `tfsdk:"port"`
	User            types.String `tfsdk:"user"`
	Password        types.String `tfsdk:"password"`
	Replicas        types.List   `tfsdk:"replicas"`
	Where           types.String `tfsdk:"where"`
	InvalidateQuery types.String `tfsdk:"invalidate_query"`
	CloseConnection types.Bool   `tfsdk:"close_connection"`
	ShareConnection types.Bool   `tfsdk:"share_connection"`
}

var DictionaryMysqlSourceAttrTypes = map[string]attr.Type{
	"db":               types.StringType,
	"table":            types.StringType,
	"port":             types.Int64Type,
	"user":             types.StringType,
	"password":         types.StringType,
	"replicas":         types.ListType{ElemType: types.ObjectType{AttrTypes: DictionaryMysqlReplicaAttrTypes}},
	"where":            types.StringType,
	"invalidate_query": types.StringType,
	"close_connection": types.BoolType,
	"share_connection": types.BoolType,
}

func flattenDictionaryMySqlSource(ctx context.Context, source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource, prevObj types.Object, diags *diag.Diagnostics) types.Object {
	if source == nil {
		return types.ObjectNull(DictionaryMysqlSourceAttrTypes)
	}

	var prevMysql DictionaryMysqlSource
	if !prevObj.IsNull() && !prevObj.IsUnknown() {
		diags.Append(prevObj.As(ctx, &prevMysql, datasize.DefaultOpts)...)
	}

	password := types.StringValue(source.Password)
	if shouldRestorePassword(password, prevMysql.Password) {
		password = prevMysql.Password
	}

	prevReplicaPasswordByHost := buildMysqlReplicaPasswordByHost(ctx, prevMysql.Replicas, diags)

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryMysqlSourceAttrTypes, DictionaryMysqlSource{
			Db:              types.StringValue(source.Db),
			Table:           types.StringValue(source.Table),
			Port:            types.Int64Value(source.Port),
			User:            types.StringValue(source.User),
			Password:        password,
			Replicas:        flattenListDictionaryMySqlReplica(ctx, source.Replicas, prevReplicaPasswordByHost, diags),
			Where:           types.StringValue(source.Where),
			InvalidateQuery: types.StringValue(source.InvalidateQuery),
			CloseConnection: flattenBoolWrapperOrFalse(source.CloseConnection),
			ShareConnection: flattenBoolWrapperOrFalse(source.ShareConnection),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryMySqlSource(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var source DictionaryMysqlSource
	diags.Append(c.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource{
		Db:              source.Db.ValueString(),
		Table:           source.Table.ValueString(),
		Port:            source.Port.ValueInt64(),
		Password:        source.Password.ValueString(),
		User:            source.User.ValueString(),
		Replicas:        expandDictionaryMySqlReplicas(ctx, source.Replicas, diags),
		Where:           source.Where.ValueString(),
		InvalidateQuery: source.InvalidateQuery.ValueString(),
		CloseConnection: mdbcommon.ExpandBoolWrapper(ctx, source.CloseConnection, diags),
		ShareConnection: mdbcommon.ExpandBoolWrapper(ctx, source.ShareConnection, diags),
	}
}

type DictionaryMysqlReplica struct {
	Host     types.String `tfsdk:"host"`
	Priority types.Int64  `tfsdk:"priority"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
}

var DictionaryMysqlReplicaAttrTypes = map[string]attr.Type{
	"host":     types.StringType,
	"priority": types.Int64Type,
	"port":     types.Int64Type,
	"user":     types.StringType,
	"password": types.StringType,
}

func flattenDictionaryMySqlReplica(ctx context.Context, replica *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_Replica, prevPwd types.String, diags *diag.Diagnostics) types.Object {
	if replica == nil {
		return types.ObjectNull(DictionaryMysqlReplicaAttrTypes)
	}

	password := types.StringValue(replica.Password)
	if shouldRestorePassword(password, prevPwd) {
		password = prevPwd
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryMysqlReplicaAttrTypes, DictionaryMysqlReplica{
			Host:     types.StringValue(replica.Host),
			Priority: types.Int64Value(replica.Priority),
			Port:     types.Int64Value(replica.Port),
			User:     types.StringValue(replica.User),
			Password: password,
		},
	)
	diags.Append(d...)

	return obj
}

func flattenListDictionaryMySqlReplica(ctx context.Context, replicas []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_Replica, prevPasswordByHost map[string]types.String, diags *diag.Diagnostics) types.List {
	if len(replicas) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: DictionaryMysqlReplicaAttrTypes})
	}

	tfReplicas := make([]types.Object, len(replicas))
	for i, r := range replicas {
		tfReplicas[i] = flattenDictionaryMySqlReplica(ctx, r, prevPasswordByHost[r.Host], diags)
	}

	list, d := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: DictionaryMysqlReplicaAttrTypes}, tfReplicas)
	diags.Append(d...)

	return list
}

func buildMysqlReplicaPasswordByHost(ctx context.Context, prevReplicas types.List, diags *diag.Diagnostics) map[string]types.String {
	if prevReplicas.IsNull() || prevReplicas.IsUnknown() {
		return nil
	}
	elems := prevReplicas.Elements()
	if len(elems) == 0 {
		return nil
	}
	result := make(map[string]types.String, len(elems))
	for _, e := range elems {
		var r DictionaryMysqlReplica
		diags.Append(e.(types.Object).As(ctx, &r, datasize.DefaultOpts)...)
		result[r.Host.ValueString()] = r.Password
	}
	return result
}

func expandDictionaryMySqlReplica(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_Replica {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var replica DictionaryMysqlReplica
	diags.Append(c.As(ctx, &replica, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_Replica{
		Host:     replica.Host.ValueString(),
		Priority: replica.Priority.ValueInt64(),
		Port:     replica.Port.ValueInt64(),
		User:     replica.User.ValueString(),
		Password: replica.Password.ValueString(),
	}
}

func expandDictionaryMySqlReplicas(ctx context.Context, c types.List, diags *diag.Diagnostics) []*clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_Replica {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	elems := c.Elements()
	result := make([]*clickhouseConfig.ClickhouseConfig_ExternalDictionary_MysqlSource_Replica, 0, len(elems))
	for _, e := range elems {
		result = append(result, expandDictionaryMySqlReplica(ctx, e.(types.Object), diags))
		if diags.HasError() {
			return nil
		}
	}

	return result
}

type DictionaryClickhouseSource struct {
	Db       types.String `tfsdk:"db"`
	Table    types.String `tfsdk:"table"`
	Host     types.String `tfsdk:"host"`
	Port     types.Int64  `tfsdk:"port"`
	User     types.String `tfsdk:"user"`
	Password types.String `tfsdk:"password"`
	Where    types.String `tfsdk:"where"`
	Secure   types.Bool   `tfsdk:"secure"`
}

var DictionaryClickhouseSourceAttrTypes = map[string]attr.Type{
	"db":       types.StringType,
	"table":    types.StringType,
	"host":     types.StringType,
	"port":     types.Int64Type,
	"user":     types.StringType,
	"password": types.StringType,
	"where":    types.StringType,
	"secure":   types.BoolType,
}

func flattenDictionaryClickHouseSource(ctx context.Context, source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_ClickhouseSource, prevObj types.Object, diags *diag.Diagnostics) types.Object {
	if source == nil {
		return types.ObjectNull(DictionaryClickhouseSourceAttrTypes)
	}

	var prevSrc DictionaryClickhouseSource
	if !prevObj.IsNull() && !prevObj.IsUnknown() {
		diags.Append(prevObj.As(ctx, &prevSrc, datasize.DefaultOpts)...)
	}

	password := types.StringValue(source.Password)
	if shouldRestorePassword(password, prevSrc.Password) {
		password = prevSrc.Password
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryClickhouseSourceAttrTypes, DictionaryClickhouseSource{
			Db:       types.StringValue(source.Db),
			Table:    types.StringValue(source.Table),
			Host:     types.StringValue(source.Host),
			Port:     types.Int64Value(source.Port),
			User:     types.StringValue(source.User),
			Password: password,
			Where:    types.StringValue(source.Where),
			Secure:   flattenBoolWrapperOrFalse(source.Secure),
		},
	)
	diags.Append(d...)

	return obj
}

// flattenBoolWrapperOrFalse returns BoolValue(false) when proto wrapper is nil.
// Used for Computed bool fields where API returns nil instead of false, to avoid perpetual diffs.
func flattenBoolWrapperOrFalse(wb *wrapperspb.BoolValue) types.Bool {
	if wb == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(wb.GetValue())
}

func expandDictionaryClickHouseSource(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_ClickhouseSource {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var source DictionaryClickhouseSource
	diags.Append(c.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_ClickhouseSource{
		Db:       source.Db.ValueString(),
		Table:    source.Table.ValueString(),
		Host:     source.Host.ValueString(),
		Port:     source.Port.ValueInt64(),
		User:     source.User.ValueString(),
		Password: source.Password.ValueString(),
		Where:    source.Where.ValueString(),
		Secure:   mdbcommon.ExpandBoolWrapper(ctx, source.Secure, diags),
	}
}

type DictionaryMongodbSource struct {
	Db         types.String `tfsdk:"db"`
	Collection types.String `tfsdk:"collection"`
	Host       types.String `tfsdk:"host"`
	Port       types.Int64  `tfsdk:"port"`
	User       types.String `tfsdk:"user"`
	Password   types.String `tfsdk:"password"`
	Options    types.String `tfsdk:"options"`
}

var DictionaryMongodbSourceAttrTypes = map[string]attr.Type{
	"db":         types.StringType,
	"collection": types.StringType,
	"host":       types.StringType,
	"port":       types.Int64Type,
	"user":       types.StringType,
	"password":   types.StringType,
	"options":    types.StringType,
}

func flattenDictionaryMongoDbSource(ctx context.Context, source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MongodbSource, prevObj types.Object, diags *diag.Diagnostics) types.Object {
	if source == nil {
		return types.ObjectNull(DictionaryMongodbSourceAttrTypes)
	}

	var prevSrc DictionaryMongodbSource
	if !prevObj.IsNull() && !prevObj.IsUnknown() {
		diags.Append(prevObj.As(ctx, &prevSrc, datasize.DefaultOpts)...)
	}

	password := types.StringValue(source.Password)
	if shouldRestorePassword(password, prevSrc.Password) {
		password = prevSrc.Password
	}

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryMongodbSourceAttrTypes, DictionaryMongodbSource{
			Db:         types.StringValue(source.Db),
			Collection: types.StringValue(source.Collection),
			Host:       types.StringValue(source.Host),
			Port:       types.Int64Value(source.Port),
			User:       types.StringValue(source.User),
			Password:   password,
			Options:    types.StringValue(source.Options),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryMongoDbSource(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_MongodbSource {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var source DictionaryMongodbSource
	diags.Append(c.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_MongodbSource{
		Db:         source.Db.ValueString(),
		Collection: source.Collection.ValueString(),
		Host:       source.Host.ValueString(),
		Port:       source.Port.ValueInt64(),
		User:       source.User.ValueString(),
		Password:   source.Password.ValueString(),
		Options:    source.Options.ValueString(),
	}
}

type DictionaryPostgresqlSource struct {
	Db              types.String `tfsdk:"db"`
	Table           types.String `tfsdk:"table"`
	Hosts           types.List   `tfsdk:"hosts"`
	Port            types.Int64  `tfsdk:"port"`
	User            types.String `tfsdk:"user"`
	Password        types.String `tfsdk:"password"`
	InvalidateQuery types.String `tfsdk:"invalidate_query"`
	SslMode         types.String `tfsdk:"ssl_mode"`
}

var DictionaryPostgresqlSourceAttrTypes = map[string]attr.Type{
	"db":               types.StringType,
	"table":            types.StringType,
	"hosts":            types.ListType{ElemType: types.StringType},
	"port":             types.Int64Type,
	"user":             types.StringType,
	"password":         types.StringType,
	"invalidate_query": types.StringType,
	"ssl_mode":         types.StringType,
}

func flattenDictionaryPostgreSqlSource(ctx context.Context, source *clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSource, prevObj types.Object, diags *diag.Diagnostics) types.Object {
	if source == nil {
		return types.ObjectNull(DictionaryPostgresqlSourceAttrTypes)
	}

	var prevSrc DictionaryPostgresqlSource
	if !prevObj.IsNull() && !prevObj.IsUnknown() {
		diags.Append(prevObj.As(ctx, &prevSrc, datasize.DefaultOpts)...)
	}

	password := types.StringValue(source.Password)
	if shouldRestorePassword(password, prevSrc.Password) {
		password = prevSrc.Password
	}

	hosts, d := types.ListValueFrom(ctx, types.StringType, source.Hosts)
	diags.Append(d...)

	obj, d := types.ObjectValueFrom(
		ctx, DictionaryPostgresqlSourceAttrTypes, DictionaryPostgresqlSource{
			Db:              types.StringValue(source.Db),
			Table:           types.StringValue(source.Table),
			Hosts:           hosts,
			Port:            types.Int64Value(source.Port),
			User:            types.StringValue(source.User),
			Password:        password,
			InvalidateQuery: types.StringValue(source.InvalidateQuery),
			SslMode:         types.StringValue(source.SslMode.Enum().String()),
		},
	)
	diags.Append(d...)

	return obj
}

func expandDictionaryPostgreSqlSource(ctx context.Context, c types.Object, diags *diag.Diagnostics) *clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSource {
	if c.IsNull() || c.IsUnknown() {
		return nil
	}

	var source DictionaryPostgresqlSource
	diags.Append(c.As(ctx, &source, datasize.DefaultOpts)...)
	if diags.HasError() {
		return nil
	}

	hosts := make([]string, len(source.Hosts.Elements()))
	diags.Append(source.Hosts.ElementsAs(ctx, &hosts, false)...)
	if diags.HasError() {
		return nil
	}

	sslMode := utils.ExpandEnum("ssl_mode", source.SslMode.ValueString(), clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSource_SslMode_value, diags)
	if diags.HasError() {
		return nil
	}

	return &clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSource{
		Db:              source.Db.ValueString(),
		Table:           source.Table.ValueString(),
		Hosts:           hosts,
		Port:            source.Port.ValueInt64(),
		User:            source.User.ValueString(),
		Password:        source.Password.ValueString(),
		InvalidateQuery: source.InvalidateQuery.ValueString(),
		SslMode:         clickhouseConfig.ClickhouseConfig_ExternalDictionary_PostgresqlSource_SslMode(*sslMode),
	}
}

func shouldRestorePassword(curr, prev types.String) bool {
	return curr.ValueString() == "" && !prev.IsNull() && !prev.IsUnknown() && prev.ValueString() != ""
}
