package xtable

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

// toYDB maps driver transaction options to ydb transaction Option or query transaction control.
// This caused by ydb logic that prevents start actual transaction with OnlineReadOnly mode and ReadCommitted
// and ReadUncommitted isolation levels should use tx_control in every query request.
// It returns error on unsupported options.
func toYDB(opts driver.TxOptions) (tx.SettingsOption, error) {
	level := sql.IsolationLevel(opts.Isolation)
	switch level {
	case sql.LevelDefault, sql.LevelSerializable:
		if !opts.ReadOnly {
			return tx.WithSerializableReadWrite(), nil
		}
	case sql.LevelSnapshot:
		if opts.ReadOnly {
			return tx.WithSnapshotReadOnly(), nil
		}
	}

	return nil, xerrors.WithStackTrace(fmt.Errorf(
		"unsupported transaction options: %+v", opts,
	))
}
