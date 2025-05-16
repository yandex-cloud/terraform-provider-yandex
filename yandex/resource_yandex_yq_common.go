package yandex

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
)

func flattenYandexYQCommonMeta(
	d *schema.ResourceData,
	meta *Ydb_FederatedQuery.CommonMeta,
) error {
	d.SetId(meta.GetId())

	return nil
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

func ParseColumnType(t string, notNull bool) (*Ydb.Type, error) {
	c, err := baseParseColumnType(strings.ToLower(t))
	if err != nil {
		return nil, err
	}

	if notNull {
		return c, err
	}
	return wrapWithOptionalIfNeeded(c), nil
}

func FormatTypeString(t *Ydb.Type) (string, error) {
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
