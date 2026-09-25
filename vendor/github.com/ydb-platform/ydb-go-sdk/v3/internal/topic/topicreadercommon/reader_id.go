package topicreadercommon

import "sync/atomic"

var globalReaderCounter int64

func NextReaderID() int64 {
	return atomic.AddInt64(&globalReaderCounter, 1)
}
