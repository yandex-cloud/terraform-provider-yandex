package xtable

import (
	"slices"
	"time"

	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
)

type Option func(*Conn)

func WithFakeTxModes(modes ...QueryMode) Option {
	return func(c *Conn) {
		c.fakeTxModes = append(c.fakeTxModes, modes...)

		slices.Sort(c.fakeTxModes)
	}
}

func WithDataOpts(dataOpts ...options.ExecuteDataQueryOption) Option {
	return func(c *Conn) {
		c.dataOpts = append(c.dataOpts, dataOpts...)
	}
}

func WithScanOpts(scanOpts ...options.ExecuteScanQueryOption) Option {
	return func(c *Conn) {
		c.scanOpts = scanOpts
	}
}

func WithDefaultTxControl(defaultTxControl *table.TransactionControl) Option {
	return func(c *Conn) {
		c.defaultTxControl = defaultTxControl
	}
}

func WithDefaultQueryMode(mode QueryMode) Option {
	return func(c *Conn) {
		c.defaultQueryMode = mode
	}
}

func WithIdleThreshold(idleThreshold time.Duration) Option {
	return func(c *Conn) {
		c.idleThreshold = idleThreshold
	}
}

func WithOnClose(onCLose func()) Option {
	return func(c *Conn) {
		c.onClose = append(c.onClose, onCLose)
	}
}
