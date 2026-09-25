package named

type Type uint8

const (
	TypeUnknown Type = iota
	TypeRequired
	TypeOptional
	TypeOptionalWithUseDefault
)

type Value struct {
	Name  string
	Value interface{}
	Type  Type
}

// Optional makes an object with destination address for column value with name columnName
//
// # If column value is NULL, then ScanNamed will write a nil into destination
//
// Warning: value must double-pointed data destination
func Optional(columnName string, destination interface{}) Value {
	if columnName == "" {
		panic("columnName must be not empty")
	}

	return Value{
		Name:  columnName,
		Value: destination,
		Type:  TypeOptional,
	}
}

// Required makes an object with destination address for column value with name columnName
//
// Warning: value must single-pointed data destination
func Required(columnName string, destinationValueReference interface{}) Value {
	if columnName == "" {
		panic("columnName must be not empty")
	}

	return Value{
		Name:  columnName,
		Value: destinationValueReference,
		Type:  TypeRequired,
	}
}

// OptionalWithDefault makes an object with destination address for column value with name columnName
//
// If scanned YDB value is NULL - default type value will be applied to value destination
// Warning: value must single-pointed data destination
func OptionalWithDefault(columnName string, destinationValueReference interface{}) Value {
	if columnName == "" {
		panic("columnName must be not empty")
	}

	return Value{
		Name:  columnName,
		Value: destinationValueReference,
		Type:  TypeOptionalWithUseDefault,
	}
}
