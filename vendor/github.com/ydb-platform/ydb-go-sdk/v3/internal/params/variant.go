package params

import "github.com/ydb-platform/ydb-go-sdk/v3/internal/value"

type (
	variant struct {
		parent Builder
		name   string
		value  value.Value
	}

	variantBuilder struct {
		variant *variant
	}
)

func (vb *variantBuilder) EndVariant() Builder {
	vb.variant.parent.params = append(vb.variant.parent.params, &Parameter{
		parent: vb.variant.parent,
		name:   vb.variant.name,
		value:  vb.variant.value,
	})

	return vb.variant.parent
}

func (v *variant) BeginTuple() *variantTuple {
	return &variantTuple{
		parent: v,
	}
}

func (v *variant) BeginStruct() *variantStruct {
	return &variantStruct{
		parent: v,
	}
}
