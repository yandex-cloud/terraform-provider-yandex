package yandex

import (
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
)

// connections
const (
	AttributeBucket           = "bucket"
	AttributeCluster          = "cluster"
	AttributeDatabaseID       = "database_id"
	AttributeDescription      = "description"
	AttributeName             = "name"
	AttributeProject          = "project"
	AttributeServiceAccountID = "service_account_id"
	AttributeSharedReading    = "shared_reading"
)

// bindings
const (
	AttributeCompression   = "compression"
	AttributeConnectionID  = "connection_id"
	AttributeFormat        = "format"
	AttributeFormatSetting = "format_setting"
	AttributePartitionedBy = "partitioned_by"
	AttributePathPattern   = "path_pattern"
	AttributeProjection    = "projection"
	AttributeStream        = "stream"

	// the same names as for ydb table
	AttributeColumn        = "column"
	AttributeColumnName    = "name"
	AttributeColumnNotNull = "not_null"
	AttributeColumnType    = "type"
)

func flattenYandexYQAuth(d *schema.ResourceData,
	auth *Ydb_FederatedQuery.IamAuth,
) error {
	serviceAccountID, err := iAMAuthToString(auth)
	if err != nil {
		return err
	}

	d.Set(AttributeServiceAccountID, serviceAccountID)

	return nil
}

func flattenYandexYQCommonMeta(
	d *schema.ResourceData,
	meta *Ydb_FederatedQuery.CommonMeta,
) error {
	d.SetId(meta.GetId())
	return nil
}

func flattenColumn(column *Ydb.Column) (map[string]any, error) {
	result := make(map[string]interface{})
	result[AttributeColumnName] = column.Name
	result[AttributeColumnNotNull] = column.Type.GetOptionalType() == nil
	columnType, err := formatTypeString(unwrapOptional(column.Type))
	if err != nil {
		return nil, err
	}
	result[AttributeColumnType] = columnType
	return result, nil
}

func flattenSchema(schema *Ydb_FederatedQuery.Schema) ([]any, error) {
	result := make([]any, 0, len(schema.Column))
	for _, column := range schema.Column {
		c, err := flattenColumn(column)
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

func parseColumns(d *schema.ResourceData) ([]*Ydb.Column, error) {
	columnsRaw := d.Get(AttributeColumn)
	if columnsRaw == nil {
		return nil, nil
	}

	raw := columnsRaw.([]any)
	columns := make([]*Ydb.Column, 0, len(raw))
	for _, rw := range raw {
		r := rw.(map[string]interface{})
		name := r[AttributeColumnName].(string)
		t := r[AttributeColumnType].(string)
		notNull := r[AttributeColumnNotNull].(bool)
		t2, err := parseColumnType(t, notNull)
		if err != nil {
			return nil, err
		}

		columns = append(columns, &Ydb.Column{
			Name: name,
			Type: t2,
		})
	}

	return columns, nil

}

func makePrimitiveType(typeId Ydb.Type_PrimitiveTypeId) *Ydb.Type {
	return &Ydb.Type{
		Type: &Ydb.Type_TypeId{
			TypeId: typeId,
		},
	}
}

func baseParseColumnType(t string) (*Ydb.Type, error) {
	if strings.HasSuffix(t, "?") {
		t2, err := baseParseColumnType(t[:len(t)-1])
		if err != nil {
			return nil, err
		}

		return wrapWithOptional(t2), nil
	}

	switch t {
	case "string":
		return makePrimitiveType(Ydb.Type_STRING), nil
	case "bool":
		return makePrimitiveType(Ydb.Type_BOOL), nil
	case "int":
		fallthrough
	case "int32":
		return makePrimitiveType(Ydb.Type_INT32), nil
	case "uint32":
		return makePrimitiveType(Ydb.Type_UINT32), nil
	case "int64":
		return makePrimitiveType(Ydb.Type_INT64), nil
	case "uint64":
		return makePrimitiveType(Ydb.Type_UINT64), nil
	case "float":
		return makePrimitiveType(Ydb.Type_FLOAT), nil
	case "double":
		return makePrimitiveType(Ydb.Type_DOUBLE), nil
	case "yson":
		return makePrimitiveType(Ydb.Type_YSON), nil
	case "utf8":
		fallthrough
	case "text":
		return makePrimitiveType(Ydb.Type_UTF8), nil
	case "json":
		return makePrimitiveType(Ydb.Type_JSON), nil
	case "date":
		return makePrimitiveType(Ydb.Type_DATE), nil
	case "datetime":
		return makePrimitiveType(Ydb.Type_DATETIME), nil
	case "timestamp":
		return makePrimitiveType(Ydb.Type_TIMESTAMP), nil
	case "interval":
		return makePrimitiveType(Ydb.Type_INTERVAL), nil
	case "int8":
		return makePrimitiveType(Ydb.Type_INT8), nil
	case "uint8":
		return makePrimitiveType(Ydb.Type_UINT8), nil
	case "int16":
		return makePrimitiveType(Ydb.Type_INT16), nil
	case "uint16":
		return makePrimitiveType(Ydb.Type_UINT16), nil
	case "tzdate":
		return makePrimitiveType(Ydb.Type_TZ_DATE), nil
	case "tzdatetime":
		return makePrimitiveType(Ydb.Type_TZ_DATETIME), nil
	case "tztimestamp":
		return makePrimitiveType(Ydb.Type_TZ_TIMESTAMP), nil
	case "uuid":
		return makePrimitiveType(Ydb.Type_UUID), nil
	case "date32":
		return makePrimitiveType(Ydb.Type_DATE32), nil
	case "datetime64":
		return makePrimitiveType(Ydb.Type_DATETIME64), nil
	case "timestamp64":
		return makePrimitiveType(Ydb.Type_TIMESTAMP64), nil
	case "interval64":
		return makePrimitiveType(Ydb.Type_INTERVAL64), nil
	}
	return nil, fmt.Errorf("unsupported type %v", t)
}

func wrapWithOptional(t *Ydb.Type) *Ydb.Type {
	if t == nil {
		return nil
	}

	return &Ydb.Type{
		Type: &Ydb.Type_OptionalType{
			OptionalType: &Ydb.OptionalType{
				Item: t,
			},
		},
	}
}

func unwrapOptional(t *Ydb.Type) *Ydb.Type {
	for t.GetOptionalType() != nil {
		t = t.GetOptionalType().GetItem()
	}
	return t
}

func wrapWithOptionalIfNeeded(t *Ydb.Type) *Ydb.Type {
	if t.GetOptionalType() != nil {
		return t
	}

	return wrapWithOptional(t)
}

func parseColumnType(t string, notNull bool) (*Ydb.Type, error) {
	c, err := baseParseColumnType(strings.ToLower(t))
	if err != nil {
		return nil, err
	}

	if notNull {
		return c, err
	}
	return wrapWithOptionalIfNeeded(c), nil
}

func formatTypeString(t *Ydb.Type) (string, error) {
	typeId := t.GetTypeId()
	switch typeId {
	case Ydb.Type_STRING:
		return "String", nil
	case Ydb.Type_BOOL:
		return "Bool", nil
	case Ydb.Type_INT8:
		return "Int8", nil
	case Ydb.Type_UINT8:
		return "Uint8", nil
	case Ydb.Type_INT16:
		return "Int16", nil
	case Ydb.Type_UINT16:
		return "Uint16", nil
	case Ydb.Type_INT32:
		return "Int32", nil
	case Ydb.Type_UINT32:
		return "Uin32", nil
	case Ydb.Type_INT64:
		return "Int64", nil
	case Ydb.Type_UINT64:
		return "Uint64", nil
	case Ydb.Type_FLOAT:
		return "Float", nil
	case Ydb.Type_DOUBLE:
		return "Double", nil
	case Ydb.Type_DATE:
		return "Date", nil
	case Ydb.Type_DATETIME:
		return "Datetime", nil
	case Ydb.Type_TIMESTAMP:
		return "Timestamp", nil
	case Ydb.Type_INTERVAL:
		return "Interval", nil
	case Ydb.Type_TZ_DATE:
		return "TzDate", nil
	case Ydb.Type_TZ_DATETIME:
		return "TzDatetime", nil
	case Ydb.Type_TZ_TIMESTAMP:
		return "TzTimestamp", nil
	case Ydb.Type_DATE32:
		return "Date32", nil
	case Ydb.Type_DATETIME64:
		return "Datetime64", nil
	case Ydb.Type_TIMESTAMP64:
		return "Timestamp64", nil
	case Ydb.Type_INTERVAL64:
		return "Interval64", nil
	case Ydb.Type_UTF8:
		return "Text", nil
	case Ydb.Type_YSON:
		return "Yson", nil
	case Ydb.Type_JSON:
		return "Json", nil
	case Ydb.Type_UUID:
		return "Uuid", nil
	case Ydb.Type_JSON_DOCUMENT:
		return "JsonDocument", nil
	case Ydb.Type_DYNUMBER:
		return "DyNumber", nil
	}

	return "", fmt.Errorf("unsupported type")
}

var (
	availableFormats = []string{
		"csv_with_names",
		"json_as_string",
		"json_each_row",
		"json_list",
		"parquet",
		"raw",
		"tsv_with_names",
	}

	availableCompressions = []string{
		"brotli",
		"bzip2",
		"gzip",
		"lz4",
		"xz",
		"zstd",
	}
)

func shouldSuppressDiffForColumnType(k, old, new string, d *schema.ResourceData) bool {
	oldLower := strings.ToLower(old)
	newLower := strings.ToLower(new)
	if oldLower == newLower {
		return true
	}

	textTypes := []string{"utf8", "text"}
	if slices.Contains(textTypes, oldLower) && slices.Contains(textTypes, newLower) {
		return true
	}

	blobTypes := []string{"string", "bytes"}
	if slices.Contains(blobTypes, oldLower) && slices.Contains(blobTypes, newLower) {
		return true
	}
	return false
}

var (
	availableBindingAttributes = map[string]*schema.Schema{
		AttributeStream: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		AttributeFormat: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(availableFormats, true),
		},
		AttributeCompression: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(availableCompressions, true),
		},
		AttributePathPattern: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		AttributeProjection: {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		AttributePartitionedBy: {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		AttributeColumn: {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					AttributeColumnName: {
						Type:         schema.TypeString,
						Description:  "Column name.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					AttributeColumnType: {
						Type:             schema.TypeString,
						Description:      "Column data type. YQL data types are used.",
						Required:         true,
						ValidateFunc:     validation.NoZeroValues,
						DiffSuppressFunc: shouldSuppressDiffForColumnType,
					},
					AttributeColumnNotNull: {
						Type:        schema.TypeBool,
						Description: "A column cannot have the NULL data type. Default: `false`.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}
)

func newBindingResourceSchema(additionalAttributes ...string) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		AttributeName: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		AttributeConnectionID: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.NoZeroValues,
		},
		AttributeDescription: {
			Type:     schema.TypeString,
			Optional: true,
		},
		AttributeFormat: {
			Type:         schema.TypeString,
			Required:     true,
			ValidateFunc: validation.StringInSlice(availableFormats, true),
		},
		AttributeCompression: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice(availableCompressions, true),
		},
		AttributeColumn: {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					AttributeColumnName: {
						Type:         schema.TypeString,
						Description:  "Column name.",
						Required:     true,
						ValidateFunc: validation.NoZeroValues,
					},
					AttributeColumnType: {
						Type:             schema.TypeString,
						Description:      "Column data type. YQL data types are used.",
						Required:         true,
						ValidateFunc:     validation.NoZeroValues,
						DiffSuppressFunc: shouldSuppressDiffForColumnType,
					},
					AttributeColumnNotNull: {
						Type:        schema.TypeBool,
						Description: "A column cannot have the NULL data type. Default: `false`.",
						Optional:    true,
						Computed:    true,
					},
				},
			},
		},
	}

	for _, a := range additionalAttributes {
		result[a] = availableBindingAttributes[a]
	}

	return result
}

func newObjectStorageBindingResourceSchema() map[string]*schema.Schema {
	return newBindingResourceSchema(AttributePathPattern, AttributeProjection, AttributePartitionedBy)
}

func newYDSBindingResourceSchema() map[string]*schema.Schema {
	return newBindingResourceSchema(AttributeStream)
}

var (
	availableConnectionAttributes = map[string]*schema.Schema{
		AttributeBucket: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeProject: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeCluster: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeDatabaseID: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeSharedReading: {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
)

func newConnectionResourceSchema(additionalAttributes ...string) map[string]*schema.Schema {
	result := map[string]*schema.Schema{
		AttributeName: {
			Type:     schema.TypeString,
			Required: true,
		},
		AttributeServiceAccountID: {
			Type:     schema.TypeString,
			Optional: true,
		},
		AttributeDescription: {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	for _, a := range additionalAttributes {
		result[a] = availableConnectionAttributes[a]
	}

	return result
}

func newObjectStorageConnectionResourceSchema() map[string]*schema.Schema {
	return newConnectionResourceSchema(AttributeBucket)
}

func newYDSConnectionResourceSchema() map[string]*schema.Schema {
	return newConnectionResourceSchema(AttributeDatabaseID)
}

func newMonitoringConnectionResourceSchema() map[string]*schema.Schema {
	return newConnectionResourceSchema(AttributeProject, AttributeCluster)
}

func newYDBConnectionResourceSchema() map[string]*schema.Schema {
	return newConnectionResourceSchema(AttributeDatabaseID)
}
