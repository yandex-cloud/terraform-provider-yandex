package errors

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type EndpointNotFoundError struct {
	Method protoreflect.FullName
}

func (e *EndpointNotFoundError) Error() string {
	return fmt.Sprintf("endpoint not found for method '%s'", e.Method)
}
