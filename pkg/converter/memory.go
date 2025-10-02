package converter

import "github.com/c2h5oh/datasize"

func ToGigabytesInFloat(bytesCount int64) float64 {
	return (datasize.ByteSize(bytesCount) * datasize.B).GBytes()
}

func ToBytesFromFloat(gigabytesCount float64) int64 {
	return int64(gigabytesCount * float64(datasize.GB))
}
