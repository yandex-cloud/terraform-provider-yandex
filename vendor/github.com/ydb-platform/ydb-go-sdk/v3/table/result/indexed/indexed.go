package indexed

// Required is a type scan destination of non-optional ydb values
// Required must be a single pointer value destination
//
// This is a proxy type for preparing go1.18 type set constrains such as
//
//	type Required interface {
//	  *int8 | *int64 | *string | types.Scanner | json.Unmarshaler
//	}
type Required interface{}

// Optional is a type scan destination of optional ydb values
// Optional must be a double pointer value destination
//
// This is a proxy type for preparing go1.18 type set constrains such as
//
//	type Optional interface {
//	  **int8 | **int64 | **string | types.Scanner | json.Unmarshaler
//	}
//
// or alias such as
//
//	type Optional *Required
type Optional interface{}

// RequiredOrOptional is a type scan destination of ydb values
// This is a proxy type for preparing go1.18 type set constrains such as
//
//	type valueType interface {
//	  Required | Optional
//	}
type RequiredOrOptional interface{}
