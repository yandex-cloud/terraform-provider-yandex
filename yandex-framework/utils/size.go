package utils

import (
	"github.com/c2h5oh/datasize"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var DefaultOpts = basetypes.ObjectAsOptions{UnhandledNullAsEmpty: false, UnhandledUnknownAsEmpty: false}

// Primary use to store value from API in state file as Gigabytes
func ToGigabytes(bytesCount int64) int64 {
	return int64((datasize.ByteSize(bytesCount) * datasize.B).GBytes())
}

// Primary use to send byte value to API
func ToBytes(gigabytesCount int64) int64 {
	return int64((datasize.ByteSize(gigabytesCount) * datasize.GB).Bytes())
}
