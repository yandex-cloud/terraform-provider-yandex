package empty

import "sync"

type (
	Chan         = chan struct{}
	ChanReadonly = <-chan struct{}
	Struct       = struct{}
)

// DoNotCopy can be embedded in a struct to help prevent shallow copies.
// This does not rely on a Go language feature, but rather a special case
// within the vet checker.
//
// See https://golang.org/issues/8005.
type DoNotCopy [0]sync.Mutex
