package topicreader

import "github.com/ydb-platform/ydb-go-sdk/v3/internal/topic/topicreaderinternal"

// WithBatchMaxCount max messages within batch
type WithBatchMaxCount int

var _ ReadBatchOption = WithBatchMaxCount(0)

// Apply implements ReadBatchOption interface
func (count WithBatchMaxCount) Apply(
	options topicreaderinternal.ReadMessageBatchOptions,
) topicreaderinternal.ReadMessageBatchOptions {
	options.MaxCount = int(count)

	return options
}

// WithBatchPreferMinCount set prefer min count for batch size. Sometime result batch can be less then count
// for example if internal buffer full and can't receive more messages or server stop send messages in partition
//
// count must be 1 or greater
// it will panic if count < 1
//
// Deprecated: was experimental and not actual now.
// The option will be removed for simplify code internals.
// Will be removed after Oct 2024.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
type WithBatchPreferMinCount int

// Apply implements ReadBatchOption interface
func (count WithBatchPreferMinCount) Apply(
	options topicreaderinternal.ReadMessageBatchOptions,
) topicreaderinternal.ReadMessageBatchOptions {
	if count < 1 {
		panic("ydb: min batch size must be 1 or greater")
	}
	options.MinCount = int(count)

	return options
}
