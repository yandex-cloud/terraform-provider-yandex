package topicwriter

type writeOptions struct{}

type WriteOption interface {
	applyWriteOption(options *writeOptions)
}
