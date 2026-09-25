package xstring

import (
	"unsafe"
)

func FromBytes(b []byte) string {
	if len(b) == 0 {
		return ""
	}

	return unsafe.String(&b[0], len(b))
}

func ToBytes(s string) (b []byte) {
	if s == "" {
		return nil
	}

	return unsafe.Slice(unsafe.StringData(s), len(s))
}
