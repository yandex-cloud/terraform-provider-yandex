package topic

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/grpcwrapper/rawydb"
)

func OperationParamsFromConfig(operationParams *rawydb.OperationParams, cfg *config.Common) {
	operationParams.OperationMode = rawydb.OperationParamsModeSync

	operationParams.OperationTimeout.Value = cfg.OperationTimeout()
	operationParams.OperationTimeout.HasValue = operationParams.OperationTimeout.Value != 0

	operationParams.CancelAfter.Value = cfg.OperationCancelAfter()
	operationParams.CancelAfter.HasValue = operationParams.CancelAfter.Value != 0
}
