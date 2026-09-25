package options

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry"
	"github.com/ydb-platform/ydb-go-sdk/v3/retry/budget"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var (
	_ DoOption = RetryOptionsOption(nil)
	_ DoOption = TraceOption{}

	_ DoTxOption = RetryOptionsOption(nil)
	_ DoTxOption = TraceOption{}
	_ DoTxOption = doTxSettingsOption{}
)

type (
	DoOption interface {
		applyDoOption(s *doSettings)
	}

	doSettings struct {
		retryOpts []retry.Option
		trace     *trace.Query
		label     string
	}

	DoTxOption interface {
		applyDoTxOption(o *doTxSettings)
	}

	doTxSettings struct {
		doSettings
		txSettings tx.Settings
	}

	RetryOptionsOption []retry.Option
	TraceOption        struct {
		t *trace.Query
	}
	doTxSettingsOption struct {
		txSettings tx.Settings
	}
)

type LabelOption string

func (label LabelOption) applyDoOption(s *doSettings) {
	s.label = string(label)
	RetryOptionsOption{retry.WithLabel(s.label)}.applyDoOption(s)
}

func (label LabelOption) applyDoTxOption(s *doTxSettings) {
	s.label = string(label)
	RetryOptionsOption{retry.WithLabel(s.label)}.applyDoTxOption(s)
}

func (label LabelOption) applyExecuteOption(s *executeSettings) {
	s.label = string(label)
}

func (opts RetryOptionsOption) applyExecuteOption(s *executeSettings) {
	s.retryOptions = append(s.retryOptions, opts...)
}

func (s *doSettings) Trace() *trace.Query {
	return s.trace
}

func (s *doSettings) RetryOpts() []retry.Option {
	return s.retryOpts
}

func (s *doSettings) Label() string {
	return s.label
}

func (s *doTxSettings) TxSettings() tx.Settings {
	return s.txSettings
}

func (opt TraceOption) applyDoOption(s *doSettings) {
	s.trace = s.trace.Compose(opt.t)
}

func (opt TraceOption) applyDoTxOption(s *doTxSettings) {
	opt.applyDoOption(&s.doSettings)
}

func (opts RetryOptionsOption) applyDoOption(s *doSettings) {
	s.retryOpts = append(s.retryOpts, opts...)
}

func (opts RetryOptionsOption) applyDoTxOption(s *doTxSettings) {
	opts.applyDoOption(&s.doSettings)
}

func (opt doTxSettingsOption) applyDoTxOption(opts *doTxSettings) {
	opts.txSettings = opt.txSettings
}

func WithTxSettings(txSettings tx.Settings) doTxSettingsOption {
	return doTxSettingsOption{txSettings: txSettings}
}

func WithIdempotent() RetryOptionsOption {
	return []retry.Option{retry.WithIdempotent(true)}
}

func WithLabel(lbl string) LabelOption {
	return LabelOption(lbl)
}

func WithTrace(t *trace.Query) TraceOption {
	return TraceOption{t: t}
}

func WithRetryBudget(b budget.Budget) RetryOptionsOption {
	return []retry.Option{retry.WithBudget(b)}
}

func ParseDoOpts(t *trace.Query, opts ...DoOption) (s *doSettings) {
	s = &doSettings{
		trace: t,
		label: "undefined",
	}

	for _, opt := range opts {
		if opt != nil {
			opt.applyDoOption(s)
		}
	}

	return s
}

func ParseDoTxOpts(t *trace.Query, opts ...DoTxOption) (s *doTxSettings) {
	s = &doTxSettings{
		txSettings: tx.NewSettings(tx.WithDefaultTxMode()),
		doSettings: doSettings{
			trace: t,
		},
	}

	for _, opt := range opts {
		if opt != nil {
			opt.applyDoTxOption(s)
		}
	}

	return s
}
