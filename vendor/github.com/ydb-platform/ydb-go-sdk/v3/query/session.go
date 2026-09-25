package query

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/node"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stats"
)

type (
	SessionInfo interface {
		node.ID

		ID() string
		Status() string
	}
	Session interface {
		SessionInfo
		Executor

		Begin(ctx context.Context, txSettings TransactionSettings) (Transaction, error)
	}
	Stats = stats.QueryStats
)
