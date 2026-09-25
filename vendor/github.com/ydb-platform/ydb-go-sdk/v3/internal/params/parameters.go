package params

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/value"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xiter"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/types"
)

type (
	NamedValue interface {
		Name() string
		Value() value.Value
	}
	Parameter struct {
		parent Builder
		name   string
		value  value.Value
	}
	Parameters interface {
		fmt.Stringer

		ToYDB() (map[string]*Ydb.TypedValue, error)
		Range() xiter.Seq2[string, value.Value]
	}
	Params []*Parameter
)

func (p *Params) Range() xiter.Seq2[string, value.Value] {
	return func(yield func(name string, v value.Value) bool) {
		for _, param := range *p {
			cont := yield(param.name, param.value)
			if !cont {
				return
			}
		}
	}
}

var _ Parameters = (*Params)(nil)

func Named(name string, value value.Value) *Parameter {
	return &Parameter{
		name:  name,
		value: value,
	}
}

func (p *Parameter) Name() string {
	return p.name
}

func (p *Parameter) Value() value.Value {
	return p.value
}

func (p *Params) String() string {
	buffer := xstring.Buffer()
	defer buffer.Free()

	buffer.WriteByte('{')
	if p != nil {
		for i, param := range *p {
			if i != 0 {
				buffer.WriteByte(',')
			}
			buffer.WriteByte('"')
			buffer.WriteString(param.name)
			buffer.WriteString("\":")
			buffer.WriteString(param.value.Yql())
		}
	}
	buffer.WriteByte('}')

	return buffer.String()
}

func (p *Params) ToYDB() (map[string]*Ydb.TypedValue, error) {
	if p == nil {
		return nil, nil //nolint:nilnil
	}

	parameters := make(map[string]*Ydb.TypedValue, len(*p))
	for _, param := range *p {
		parameters[param.name] = value.ToYDB(param.value)
	}

	return parameters, nil
}

func (p *Params) toYDB() map[string]*Ydb.TypedValue {
	if p == nil {
		return nil
	}
	parameters := make(map[string]*Ydb.TypedValue, len(*p))
	for _, param := range *p {
		parameters[param.name] = value.ToYDB(param.value)
	}

	return parameters
}

func (p *Params) Each(it func(name string, v value.Value)) {
	if p == nil {
		return
	}
	for _, p := range *p {
		it(p.name, p.value)
	}
}

func (p *Params) Count() int {
	if p == nil {
		return 0
	}

	return len(*p)
}

func (p *Params) Add(params ...NamedValue) {
	for _, param := range params {
		*p = append(*p, Named(param.Name(), param.Value()))
	}
}

func (p *Parameter) BeginOptional() *optional {
	return &optional{
		parent: p.parent,
		name:   p.name,
	}
}

func (p *Parameter) BeginList() *list {
	return &list{
		parent: p.parent,
		name:   p.name,
	}
}

func (p *Parameter) Pg() pgParam {
	return pgParam{p}
}

func (p *Parameter) BeginSet() *set {
	return &set{
		parent: p.parent,
		name:   p.name,
	}
}

func (p *Parameter) BeginDict() *dict {
	return &dict{
		parent: p.parent,
		name:   p.name,
	}
}

func (p *Parameter) BeginTuple() *tuple {
	return &tuple{
		parent: p.parent,
		name:   p.name,
	}
}

func (p *Parameter) BeginStruct() *structure {
	return &structure{
		parent: p.parent,
		name:   p.name,
	}
}

func (p *Parameter) BeginVariant() *variant {
	return &variant{
		parent: p.parent,
		name:   p.name,
	}
}

func (p *Parameter) Text(v string) Builder {
	p.value = value.TextValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Bytes(v []byte) Builder {
	p.value = value.BytesValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Bool(v bool) Builder {
	p.value = value.BoolValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Uint64(v uint64) Builder {
	p.value = value.Uint64Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Int64(v int64) Builder {
	p.value = value.Int64Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Uint32(v uint32) Builder {
	p.value = value.Uint32Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Int32(v int32) Builder {
	p.value = value.Int32Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Uint16(v uint16) Builder {
	p.value = value.Uint16Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Int16(v int16) Builder {
	p.value = value.Int16Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Uint8(v uint8) Builder {
	p.value = value.Uint8Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Int8(v int8) Builder {
	p.value = value.Int8Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Float(v float32) Builder {
	p.value = value.FloatValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Double(v float64) Builder {
	p.value = value.DoubleValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Decimal(v [16]byte, precision, scale uint32) Builder {
	p.value = value.DecimalValue(v, precision, scale)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Timestamp(v time.Time) Builder {
	p.value = value.TimestampValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Timestamp64(v time.Time) Builder {
	p.value = value.Timestamp64ValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Date(v time.Time) Builder {
	p.value = value.DateValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Date32(v time.Time) Builder {
	p.value = value.Date32ValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Datetime(v time.Time) Builder {
	p.value = value.DatetimeValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Datetime64(v time.Time) Builder {
	p.value = value.Datetime64ValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Interval(v time.Duration) Builder {
	p.value = value.IntervalValueFromDuration(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Interval64(v time.Duration) Builder {
	p.value = value.Interval64ValueFromDuration(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) JSON(v string) Builder {
	p.value = value.JSONValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) JSONDocument(v string) Builder {
	p.value = value.JSONDocumentValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) YSON(v []byte) Builder {
	p.value = value.YSONValue(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

//// removed for https://github.com/ydb-platform/ydb-go-sdk/issues/1501
////func (p *Parameter) UUID(v [16]byte) Builder {
////	return p.UUIDWithIssue1501Value(v)
////}

// UUIDWithIssue1501Value is field serializer for save data with format bug.
// For any new code use Uuid
// https://github.com/ydb-platform/ydb-go-sdk/issues/1501
func (p *Parameter) UUIDWithIssue1501Value(v [16]byte) Builder {
	p.value = value.UUIDWithIssue1501Value(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Uuid(val uuid.UUID) Builder { //nolint:revive
	p.value = value.Uuid(val)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Any(v types.Value) Builder {
	p.value = v

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) TzDate(v time.Time) Builder {
	p.value = value.TzDateValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) TzTimestamp(v time.Time) Builder {
	p.value = value.TzTimestampValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) TzDatetime(v time.Time) Builder {
	p.value = value.TzDatetimeValueFromTime(v)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func (p *Parameter) Raw(pb *Ydb.TypedValue) Builder {
	p.value = value.FromProtobuf(pb)

	return Builder{
		params: append(p.parent.params, p),
	}
}

func Declare(p *Parameter) string {
	return fmt.Sprintf(
		"DECLARE %s AS %s",
		p.name,
		p.value.Type().Yql(),
	)
}
