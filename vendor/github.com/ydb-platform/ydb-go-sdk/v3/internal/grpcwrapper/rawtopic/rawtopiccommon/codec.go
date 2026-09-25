package rawtopiccommon

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"
)

// Codec any int value, for example for custom codec
type Codec int32

const (
	CodecUNSPECIFIED Codec = iota
	CodecRaw               = Codec(Ydb_Topic.Codec_CODEC_RAW)
	CodecGzip              = Codec(Ydb_Topic.Codec_CODEC_GZIP)
	CodecLzop              = Codec(Ydb_Topic.Codec_CODEC_LZOP)
	CodecZstd              = Codec(Ydb_Topic.Codec_CODEC_ZSTD)
)

const (
	CodecCustomerFirst = 10000
	CodecCustomerEnd   = 20000 // last allowed custom codec id is 19999
)

func (c Codec) IsCustomerCodec() bool {
	return c >= CodecCustomerFirst && c <= CodecCustomerEnd
}

func (c *Codec) MustFromProto(codec Ydb_Topic.Codec) {
	*c = Codec(codec)
}

func (c Codec) ToProto() Ydb_Topic.Codec {
	return Ydb_Topic.Codec(c)
}

func (c Codec) ToInt32() int32 {
	return int32(c)
}

type SupportedCodecs []Codec

func (c *SupportedCodecs) AllowedByCodecsList(need Codec) bool {
	// empty list allow any codec
	if len(*c) == 0 {
		return true
	}

	return c.Contains(need)
}

func (c *SupportedCodecs) Contains(need Codec) bool {
	for _, v := range *c {
		if v == need {
			return true
		}
	}

	return false
}

func (c *SupportedCodecs) Clone() SupportedCodecs {
	res := make(SupportedCodecs, len(*c))
	copy(res, *c)

	return res
}

func (c *SupportedCodecs) IsEqualsTo(other SupportedCodecs) bool {
	if len(*c) != len(other) {
		return false
	}

	codecs := make(map[Codec]struct{})
	for _, v := range *c {
		codecs[v] = struct{}{}
	}

	for _, v := range other {
		if _, ok := codecs[v]; !ok {
			return false
		}
	}

	return true
}

func (c *SupportedCodecs) ToProto() *Ydb_Topic.SupportedCodecs {
	codecs := *c
	proto := &Ydb_Topic.SupportedCodecs{
		Codecs: make([]int32, len(codecs)),
	}
	for i := range codecs {
		proto.Codecs[i] = int32(codecs[i].ToProto().Number())
	}

	return proto
}

func (c *SupportedCodecs) MustFromProto(proto *Ydb_Topic.SupportedCodecs) {
	res := make([]Codec, len(proto.GetCodecs()))
	for i := range proto.GetCodecs() {
		res[i].MustFromProto(Ydb_Topic.Codec(proto.GetCodecs()[i]))
	}
	*c = res
}
