package params

type (
	Builder struct {
		params Params
	}
)

func (b Builder) Build() *Params {
	return &b.params
}

func (b Builder) build() *Params {
	return &b.params
}

func (b Builder) Param(name string) *Parameter {
	return &Parameter{
		parent: b,
		name:   name,
		value:  nil,
	}
}
