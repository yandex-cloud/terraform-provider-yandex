package stats

import (
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stats"
)

// QueryPhase holds query execution phase statistics.
type QueryPhase = stats.QueryPhase

// QueryStats holds query execution statistics.
type QueryStats = stats.QueryStats

// CompilationStats holds query compilation statistics.
type CompilationStats = stats.CompilationStats

// TableAccess contains query execution phase's table access statistics.
type TableAccess = stats.TableAccess

type OperationStats = stats.OperationStats
