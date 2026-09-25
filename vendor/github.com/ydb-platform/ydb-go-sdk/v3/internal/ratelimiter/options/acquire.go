package options

import "time"

const (
	DefaultDecrease = 100 * time.Millisecond
)

type AcquireType uint8

const (
	AcquireTypeAcquire = AcquireType(iota)
	AcquireTypeReport

	AcquireTypeDefault = AcquireTypeAcquire
)

type Acquire interface {
	// Type defines type of acquire request
	Type() AcquireType

	// OperationTimeout defines operation Timeout for acquire request
	OperationTimeout() time.Duration

	// OperationCancelAfter defines operation CancelAfter for acquire request
	OperationCancelAfter() time.Duration
}

type acquireOptionsHolder struct {
	acquireType          AcquireType
	operationTimeout     time.Duration
	operationCancelAfter time.Duration
}

func (h *acquireOptionsHolder) OperationTimeout() time.Duration {
	return h.operationTimeout
}

func (h *acquireOptionsHolder) OperationCancelAfter() time.Duration {
	return h.operationTimeout
}

func (h *acquireOptionsHolder) Type() AcquireType {
	return h.acquireType
}

type AcquireOption func(h *acquireOptionsHolder)

func WithAcquire() AcquireOption {
	return func(h *acquireOptionsHolder) {
		h.acquireType = AcquireTypeAcquire
	}
}

func WithReport() AcquireOption {
	return func(h *acquireOptionsHolder) {
		h.acquireType = AcquireTypeReport
	}
}

func WithOperationTimeout(operationTimeout time.Duration) AcquireOption {
	return func(h *acquireOptionsHolder) {
		h.operationTimeout = operationTimeout
	}
}

func WithOperationCancelAfter(operationCancelAfter time.Duration) AcquireOption {
	return func(h *acquireOptionsHolder) {
		h.operationCancelAfter = operationCancelAfter
	}
}

func NewAcquire(opts ...AcquireOption) Acquire {
	h := &acquireOptionsHolder{
		acquireType: AcquireTypeDefault,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(h)
		}
	}

	return h
}
