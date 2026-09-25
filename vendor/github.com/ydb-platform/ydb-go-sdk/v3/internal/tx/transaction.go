package tx

import (
	"context"
	"fmt"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type Transaction interface {
	Identifier
	UnLazy(ctx context.Context) error
	SessionID() string

	// OnBeforeCommit add callback, which will be called before commit transaction
	// the method will be not call the method if some error happen and transaction will not be committed
	OnBeforeCommit(f OnTransactionBeforeCommit)
	OnCompleted(f OnTransactionCompletedFunc)
	Rollback(ctx context.Context) error
}

type (
	OnTransactionBeforeCommit  func(ctx context.Context) error
	OnTransactionCompletedFunc func(transactionResult error)
)

func AsTransaction(id Identifier) (Transaction, error) {
	if t, ok := id.(Transaction); ok {
		return t, nil
	}

	return nil, xerrors.WithStackTrace(xerrors.Wrap(fmt.Errorf(
		"waiting transaction object of type query.Transaction, got: %T",
		id,
	)))
}
