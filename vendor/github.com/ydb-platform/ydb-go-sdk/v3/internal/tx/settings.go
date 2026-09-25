package tx

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"
)

var (
	querySerializableReadWrite = &Ydb_Query.TransactionSettings_SerializableReadWrite{
		SerializableReadWrite: &Ydb_Query.SerializableModeSettings{},
	}
	queryStaleReadOnly = &Ydb_Query.TransactionSettings_StaleReadOnly{
		StaleReadOnly: &Ydb_Query.StaleModeSettings{},
	}
	querySnapshotReadOnly = &Ydb_Query.TransactionSettings_SnapshotReadOnly{
		SnapshotReadOnly: &Ydb_Query.SnapshotModeSettings{},
	}
	queryOnlineReadOnlyAllowInconsistentReads = &Ydb_Query.TransactionSettings_OnlineReadOnly{
		OnlineReadOnly: &Ydb_Query.OnlineModeSettings{AllowInconsistentReads: true},
	}
	queryOnlineReadOnlyForbidInconsistentReads = &Ydb_Query.TransactionSettings_OnlineReadOnly{
		OnlineReadOnly: &Ydb_Query.OnlineModeSettings{AllowInconsistentReads: false},
	}
	tableSerializableReadWrite = &Ydb_Table.TransactionSettings_SerializableReadWrite{
		SerializableReadWrite: &Ydb_Table.SerializableModeSettings{},
	}
	tableStaleReadOnly = &Ydb_Table.TransactionSettings_StaleReadOnly{
		StaleReadOnly: &Ydb_Table.StaleModeSettings{},
	}
	tableSnapshotReadOnly = &Ydb_Table.TransactionSettings_SnapshotReadOnly{
		SnapshotReadOnly: &Ydb_Table.SnapshotModeSettings{},
	}
	tableOnlineReadOnlyAllowInconsistentReads = &Ydb_Table.TransactionSettings_OnlineReadOnly{
		OnlineReadOnly: &Ydb_Table.OnlineModeSettings{AllowInconsistentReads: true},
	}
	tableOnlineReadOnlyForbidInconsistentReads = &Ydb_Table.TransactionSettings_OnlineReadOnly{
		OnlineReadOnly: &Ydb_Table.OnlineModeSettings{AllowInconsistentReads: false},
	}
)

// Transaction settings options
type (
	SettingsOption interface {
		ApplyQueryTxSettingsOption(txSettings *Ydb_Query.TransactionSettings)
		ApplyTableTxSettingsOption(txSettings *Ydb_Table.TransactionSettings)
	}
	Settings []SettingsOption
)

func (opts Settings) applyTableTxSelector(txControl *Ydb_Table.TransactionControl) {
	beginTx := &Ydb_Table.TransactionControl_BeginTx{
		BeginTx: &Ydb_Table.TransactionSettings{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTableTxSettingsOption(beginTx.BeginTx)
		}
	}
	txControl.TxSelector = beginTx
}

func (opts Settings) applyQueryTxSelector(txControl *Ydb_Query.TransactionControl) {
	beginTx := &Ydb_Query.TransactionControl_BeginTx{
		BeginTx: &Ydb_Query.TransactionSettings{},
	}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyQueryTxSettingsOption(beginTx.BeginTx)
		}
	}
	txControl.TxSelector = beginTx
}

func (opts Settings) ToYdbQuerySettings() *Ydb_Query.TransactionSettings {
	txSettings := &Ydb_Query.TransactionSettings{}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyQueryTxSettingsOption(txSettings)
		}
	}

	return txSettings
}

func (opts Settings) ToYdbTableSettings() *Ydb_Table.TransactionSettings {
	txSettings := &Ydb_Table.TransactionSettings{}
	for _, opt := range opts {
		if opt != nil {
			opt.ApplyTableTxSettingsOption(txSettings)
		}
	}

	return txSettings
}

// NewSettings returns transaction settings
func NewSettings(opts ...SettingsOption) Settings {
	return opts
}

func WithDefaultTxMode() SettingsOption {
	return WithSerializableReadWrite()
}

var _ SettingsOption = serializableReadWriteTxSettingsOption{}

type serializableReadWriteTxSettingsOption struct{}

func (serializableReadWriteTxSettingsOption) ApplyTableTxSettingsOption(settings *Ydb_Table.TransactionSettings) {
	settings.TxMode = tableSerializableReadWrite
}

func (serializableReadWriteTxSettingsOption) ApplyQueryTxSettingsOption(settings *Ydb_Query.TransactionSettings) {
	settings.TxMode = querySerializableReadWrite
}

func WithSerializableReadWrite() SettingsOption {
	return serializableReadWriteTxSettingsOption{}
}

var _ SettingsOption = snapshotReadOnlyTxSettingsOption{}

type snapshotReadOnlyTxSettingsOption struct{}

func (snapshotReadOnlyTxSettingsOption) ApplyTableTxSettingsOption(settings *Ydb_Table.TransactionSettings) {
	settings.TxMode = tableSnapshotReadOnly
}

func (snapshotReadOnlyTxSettingsOption) ApplyQueryTxSettingsOption(settings *Ydb_Query.TransactionSettings) {
	settings.TxMode = querySnapshotReadOnly
}

func WithSnapshotReadOnly() SettingsOption {
	return snapshotReadOnlyTxSettingsOption{}
}

var _ SettingsOption = staleReadOnlySettingsOption{}

type staleReadOnlySettingsOption struct{}

func (staleReadOnlySettingsOption) ApplyTableTxSettingsOption(settings *Ydb_Table.TransactionSettings) {
	settings.TxMode = tableStaleReadOnly
}

func (staleReadOnlySettingsOption) ApplyQueryTxSettingsOption(settings *Ydb_Query.TransactionSettings) {
	settings.TxMode = queryStaleReadOnly
}

func WithStaleReadOnly() SettingsOption {
	return staleReadOnlySettingsOption{}
}

type (
	onlineReadOnly       bool
	OnlineReadOnlyOption interface {
		applyTxOnlineReadOnlyOption(opt *onlineReadOnly)
	}
)

var _ OnlineReadOnlyOption = inconsistentReadsTxOnlineReadOnlyOption{}

type inconsistentReadsTxOnlineReadOnlyOption struct{}

func (i inconsistentReadsTxOnlineReadOnlyOption) applyTxOnlineReadOnlyOption(b *onlineReadOnly) {
	*b = true
}

func WithInconsistentReads() OnlineReadOnlyOption {
	return inconsistentReadsTxOnlineReadOnlyOption{}
}

var _ SettingsOption = onlineReadOnlySettingsOption{}

type onlineReadOnlySettingsOption []OnlineReadOnlyOption

func (opts onlineReadOnlySettingsOption) ApplyQueryTxSettingsOption(settings *Ydb_Query.TransactionSettings) {
	var ro onlineReadOnly
	for _, opt := range opts {
		if opt != nil {
			opt.applyTxOnlineReadOnlyOption(&ro)
		}
	}
	if ro {
		settings.TxMode = queryOnlineReadOnlyAllowInconsistentReads
	} else {
		settings.TxMode = queryOnlineReadOnlyForbidInconsistentReads
	}
}

func (opts onlineReadOnlySettingsOption) ApplyTableTxSettingsOption(settings *Ydb_Table.TransactionSettings) {
	var ro onlineReadOnly
	for _, opt := range opts {
		if opt != nil {
			opt.applyTxOnlineReadOnlyOption(&ro)
		}
	}
	if ro {
		settings.TxMode = tableOnlineReadOnlyAllowInconsistentReads
	} else {
		settings.TxMode = tableOnlineReadOnlyForbidInconsistentReads
	}
}

func WithOnlineReadOnly(opts ...OnlineReadOnlyOption) onlineReadOnlySettingsOption {
	return opts
}
