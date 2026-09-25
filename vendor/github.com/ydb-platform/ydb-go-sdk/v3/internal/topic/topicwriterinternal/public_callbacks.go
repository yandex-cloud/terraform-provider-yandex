package topicwriterinternal

import "github.com/ydb-platform/ydb-go-sdk/v3/topic/topictypes"

type PublicWithOnWriterConnectedInfo struct {
	LastSeqNo        int64
	SessionID        string
	PartitionID      int64
	CodecsFromServer []topictypes.Codec
	AllowedCodecs    []topictypes.Codec // Intersection between codecs from server and codecs, supported by writer
}

type PublicOnWriterInitResponseCallback func(info PublicWithOnWriterConnectedInfo) error
