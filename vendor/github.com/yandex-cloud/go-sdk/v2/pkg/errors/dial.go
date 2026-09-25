package errors

import (
	"errors"
	"fmt"
)

// ErrConnContextClosed is returned when a client connection operation is attempted on a closed connection context.
var ErrConnContextClosed = errors.New("client connection context closed")

// DialError represents an error encountered while attempting to dial a specific endpoint.
// It includes both the address of the endpoint and the error that occurred.
type DialError struct {
	Err  error
	Addr string
}

// Error returns the error message formatted with the endpoint address and the underlying error details.
func (d *DialError) Error() string {
	return fmt.Sprintf("error dialing to endpoint '%s': %s", d.Addr, d.Err.Error())
}
