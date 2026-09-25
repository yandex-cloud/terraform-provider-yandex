package query

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
)

type (
	TxActor interface {
		tx.Identifier
		Executor
	}
	TransactionActor = TxActor
	Transaction      interface {
		TxActor

		CommitTx(ctx context.Context) (err error)
		Rollback(ctx context.Context) (err error)
	}
	TransactionControl  = tx.Control
	TransactionSettings = tx.Settings
	TransactionOption   = tx.SettingsOption
)

// BeginTx returns selector transaction control option
func BeginTx(opts ...TransactionOption) tx.ControlOption {
	return tx.BeginTx(opts...)
}

func WithTx(t tx.Identifier) tx.ControlOption {
	return tx.WithTx(t)
}

func WithTxID(txID string) tx.ControlOption {
	return tx.WithTxID(txID)
}

// CommitTx returns commit transaction control option
func CommitTx() tx.ControlOption {
	return tx.CommitTx()
}

// TxControl makes transaction control from given options
func TxControl(opts ...tx.ControlOption) *TransactionControl {
	return tx.NewControl(opts...)
}

// EmptyTxControl defines transaction control inference on server-side by query content
func EmptyTxControl() *TransactionControl {
	return nil
}

// NoTx defines nil transaction control
// This is wrong name for transaction control inference on server-side by query content
// Deprecated: Use EmptyTxControl instead.
// Will be removed after Oct 2025.
// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
func NoTx() *TransactionControl {
	return EmptyTxControl()
}

// DefaultTxControl returns default transaction control for use default tx control on server-side
func DefaultTxControl() *TransactionControl {
	return EmptyTxControl()
}

// SerializableReadWriteTxControl returns transaction control with serializable read-write isolation mode
func SerializableReadWriteTxControl(opts ...tx.ControlOption) *TransactionControl {
	return tx.SerializableReadWriteTxControl(opts...)
}

// OnlineReadOnlyTxControl returns online read-only transaction control
func OnlineReadOnlyTxControl(opts ...tx.OnlineReadOnlyOption) *TransactionControl {
	return TxControl(
		BeginTx(WithOnlineReadOnly(opts...)),
		CommitTx(), // open transactions not supported for OnlineReadOnly
	)
}

// StaleReadOnlyTxControl returns stale read-only transaction control
func StaleReadOnlyTxControl() *TransactionControl {
	return TxControl(
		BeginTx(WithStaleReadOnly()),
		CommitTx(), // open transactions not supported for StaleReadOnly
	)
}

// SnapshotReadOnlyTxControl returns snapshot read-only transaction control
func SnapshotReadOnlyTxControl() *TransactionControl {
	return TxControl(
		BeginTx(WithSnapshotReadOnly()),
		CommitTx(), // open transactions not supported for StaleReadOnly
	)
}

// TxSettings returns transaction settings
func TxSettings(opts ...tx.SettingsOption) TransactionSettings {
	return opts
}

func WithDefaultTxMode() TransactionOption {
	return tx.WithDefaultTxMode()
}

func WithSerializableReadWrite() TransactionOption {
	return tx.WithSerializableReadWrite()
}

func WithSnapshotReadOnly() TransactionOption {
	return tx.WithSnapshotReadOnly()
}

func WithStaleReadOnly() TransactionOption {
	return tx.WithStaleReadOnly()
}

func WithInconsistentReads() tx.OnlineReadOnlyOption {
	return tx.WithInconsistentReads()
}

func WithOnlineReadOnly(opts ...tx.OnlineReadOnlyOption) TransactionOption {
	return tx.WithOnlineReadOnly(opts...)
}
