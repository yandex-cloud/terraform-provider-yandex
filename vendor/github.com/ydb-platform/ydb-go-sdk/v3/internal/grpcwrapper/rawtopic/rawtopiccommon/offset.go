package rawtopiccommon

type Offset int64

func NewOffset(v int64) Offset {
	return Offset(v)
}

func (offset *Offset) FromInt64(v int64) {
	*offset = Offset(v)
}

func (offset Offset) ToInt64() int64 {
	return int64(offset)
}
