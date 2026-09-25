package xstring

import (
	"bytes"
	"fmt"
	"sync"
)

type buffer struct {
	bytes.Buffer
}

var buffersPool = sync.Pool{New: func() interface{} {
	return &buffer{}
}}

func (b *buffer) Free() {
	b.Reset()
	buffersPool.Put(b)
}

func Buffer() *buffer {
	val, ok := buffersPool.Get().(*buffer)
	if !ok {
		panic(fmt.Sprintf("unsupported type conversion from %T to *buffer", val))
	}

	return val
}
