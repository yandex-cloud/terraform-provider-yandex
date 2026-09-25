package table

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/operation"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/params"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/table/scanner"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/tx"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/table"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

var (
	errTxAlreadyCommitted = xerrors.Wrap(fmt.Errorf("transaction already committed"))
	errTxRollbackedEarly  = xerrors.Wrap(fmt.Errorf("transaction rollbacked early"))
)

type txState struct {
	rawVal atomic.Uint32
}

func (s *txState) Load() txStateEnum {
	return txStateEnum(s.rawVal.Load())
}

func (s *txState) Store(val txStateEnum) {
	s.rawVal.Store(uint32(val))
}

type txStateEnum uint32

const (
	txStateInitialized txStateEnum = iota
	txStateCommitted
	txStateRollbacked
)

var _ tx.Identifier = (*transaction)(nil)

type transaction struct {
	tx.Identifier

	s       *Session
	control *table.TransactionControl
	state   txState
}

// Execute executes query represented by text within transaction tx.
func (tx *transaction) Execute(ctx context.Context, sql string, params *params.Params,
	opts ...options.ExecuteDataQueryOption,
) (r result.Result, err error) {
	onDone := trace.TableOnTxExecute(
		tx.s.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*transaction).Execute"),
		tx.s, tx, queryFromText(sql), params,
	)
	defer func() {
		onDone(r, err)
	}()

	switch tx.state.Load() {
	case txStateCommitted:
		return nil, xerrors.WithStackTrace(errTxAlreadyCommitted)
	case txStateRollbacked:
		return nil, xerrors.WithStackTrace(errTxRollbackedEarly)
	default:
		_, r, err = tx.s.Execute(ctx, tx.control, sql, params, opts...)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		if tx.control.Commit() {
			tx.state.Store(txStateCommitted)
		}

		return r, nil
	}
}

// ExecuteStatement executes prepared statement stmt within transaction tx.
func (tx *transaction) ExecuteStatement(
	ctx context.Context,
	stmt table.Statement, parameters *params.Params,
	opts ...options.ExecuteDataQueryOption,
) (r result.Result, err error) {
	val, ok := stmt.(*statement)
	if !ok {
		panic(fmt.Sprintf("unsupported type conversion from %T to *statement", val))
	}

	onDone := trace.TableOnTxExecuteStatement(
		tx.s.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*transaction).ExecuteStatement"),
		tx.s, tx, val.query, parameters,
	)
	defer func() {
		onDone(r, err)
	}()

	switch tx.state.Load() {
	case txStateCommitted:
		return nil, xerrors.WithStackTrace(errTxAlreadyCommitted)
	case txStateRollbacked:
		return nil, xerrors.WithStackTrace(errTxRollbackedEarly)
	default:
		_, r, err = stmt.Execute(ctx, tx.control, parameters, opts...)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		if tx.control.Commit() {
			tx.state.Store(txStateCommitted)
		}

		return r, nil
	}
}

// CommitTx commits specified active transaction.
func (tx *transaction) CommitTx(
	ctx context.Context,
	opts ...options.CommitTransactionOption,
) (r result.Result, err error) {
	onDone := trace.TableOnTxCommit(
		tx.s.config.Trace(), &ctx,
		stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*transaction).CommitTx"),
		tx.s, tx,
	)
	defer func() {
		onDone(err)
	}()

	switch tx.state.Load() {
	case txStateCommitted:
		return nil, xerrors.WithStackTrace(errTxAlreadyCommitted)
	case txStateRollbacked:
		return nil, xerrors.WithStackTrace(errTxRollbackedEarly)
	default:
		var (
			request = &Ydb_Table.CommitTransactionRequest{
				SessionId: tx.s.id,
				TxId:      tx.ID(),
				OperationParams: operation.Params(
					ctx,
					tx.s.config.OperationTimeout(),
					tx.s.config.OperationCancelAfter(),
					operation.ModeSync,
				),
			}
			response *Ydb_Table.CommitTransactionResponse
			result   = new(Ydb_Table.CommitTransactionResult)
		)

		for _, opt := range opts {
			if opt != nil {
				opt((*options.CommitTransactionDesc)(request))
			}
		}

		response, err = tx.s.client.CommitTransaction(ctx, request)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		err = response.GetOperation().GetResult().UnmarshalTo(result)
		if err != nil {
			return nil, xerrors.WithStackTrace(err)
		}

		tx.state.Store(txStateCommitted)

		return scanner.NewUnary(
			nil,
			result.GetQueryStats(),
			scanner.WithIgnoreTruncated(tx.s.config.IgnoreTruncated()),
		), nil
	}
}

// Rollback performs a rollback of the specified active transaction.
func (tx *transaction) Rollback(ctx context.Context) (err error) {
	switch tx.state.Load() {
	case txStateCommitted:
		return nil // nop for committed tx
	case txStateRollbacked:
		return xerrors.WithStackTrace(errTxRollbackedEarly)
	default:
		onDone := trace.TableOnTxRollback(
			tx.s.config.Trace(), &ctx,
			stack.FunctionID("github.com/ydb-platform/ydb-go-sdk/v3/internal/table.(*transaction).Rollback"),
			tx.s, tx,
		)
		defer func() {
			onDone(err)
		}()

		_, err = tx.s.client.RollbackTransaction(ctx,
			&Ydb_Table.RollbackTransactionRequest{
				SessionId: tx.s.id,
				TxId:      tx.ID(),
				OperationParams: operation.Params(
					ctx,
					tx.s.config.OperationTimeout(),
					tx.s.config.OperationCancelAfter(),
					operation.ModeSync,
				),
			},
		)
		if err != nil {
			return xerrors.WithStackTrace(err)
		}

		tx.state.Store(txStateRollbacked)

		return nil
	}
}
