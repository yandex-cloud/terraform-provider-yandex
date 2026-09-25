package stats

import (
	"time"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_TableStats"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xiter"
)

type (
	// QueryStats holds query execution statistics.
	QueryStats interface {
		ProcessCPUTime() time.Duration
		Compilation() (c *CompilationStats)
		QueryPlan() string
		QueryAST() string
		TotalCPUTime() time.Duration
		TotalDuration() time.Duration

		// NextPhase returns next execution phase within query.
		// If ok flag is false, then there are no more phases and p is invalid.
		NextPhase() (p QueryPhase, ok bool)

		// QueryPhases is a range iterator over query phases.
		QueryPhases() xiter.Seq[QueryPhase]
	}
	// QueryPhase holds query execution phase statistics.
	QueryPhase interface {
		// NextTableAccess returns next accessed table within query execution phase.
		// If ok flag is false, then there are no more accessed tables and t is invalid.
		NextTableAccess() (t *TableAccess, ok bool)
		// TableAccess is a range iterator over query execution phase's accessed tables.
		TableAccess() xiter.Seq[*TableAccess]
		Duration() time.Duration
		CPUTime() time.Duration
		AffectedShards() uint64
		IsLiteralPhase() bool
	}
	OperationStats struct {
		Rows  uint64
		Bytes uint64
	}
	Phase struct {
		Duration       time.Duration
		TableAccess    []TableAccess
		CPUTime        time.Duration
		AffectedShards uint64
		LiteralPhase   bool
	}
	// TableAccess contains query execution phase's table access statistics.
	TableAccess struct {
		Name            string
		Reads           OperationStats
		Updates         OperationStats
		Deletes         OperationStats
		PartitionsCount uint64
	}
	// CompilationStats holds query compilation statistics.
	CompilationStats struct {
		FromCache bool
		Duration  time.Duration
		CPUTime   time.Duration
	}
	// queryStats holds query execution statistics.
	queryStats struct {
		pb  *Ydb_TableStats.QueryStats
		pos int
	}
	// queryPhase holds query execution phase statistics.
	queryPhase struct {
		pb  *Ydb_TableStats.QueryPhaseStats
		pos int
	}
)

func fromUs(us uint64) time.Duration {
	return time.Duration(us) * time.Microsecond
}

func fromCompilationStats(pb *Ydb_TableStats.CompilationStats) *CompilationStats {
	return &CompilationStats{
		FromCache: pb.GetFromCache(),
		Duration:  fromUs(pb.GetDurationUs()),
		CPUTime:   fromUs(pb.GetCpuTimeUs()),
	}
}

func fromOperationStats(pb *Ydb_TableStats.OperationStats) OperationStats {
	return OperationStats{
		Rows:  pb.GetRows(),
		Bytes: pb.GetBytes(),
	}
}

func (stats *queryStats) ProcessCPUTime() time.Duration {
	return fromUs(stats.pb.GetProcessCpuTimeUs())
}

func (stats *queryStats) Compilation() (c *CompilationStats) {
	return fromCompilationStats(stats.pb.GetCompilation())
}

func (stats *queryStats) QueryPlan() string {
	return stats.pb.GetQueryPlan()
}

func (stats *queryStats) QueryAST() string {
	return stats.pb.GetQueryAst()
}

func (stats *queryStats) TotalCPUTime() time.Duration {
	return fromUs(stats.pb.GetTotalCpuTimeUs())
}

func (stats *queryStats) TotalDuration() time.Duration {
	return fromUs(stats.pb.GetTotalDurationUs())
}

// NextPhase returns next execution phase within query.
// If ok flag is false, then there are no more phases and p is invalid.
func (stats *queryStats) NextPhase() (p QueryPhase, ok bool) {
	if stats.pos >= len(stats.pb.GetQueryPhases()) {
		return
	}
	pb := stats.pb.GetQueryPhases()[stats.pos]
	if pb == nil {
		return
	}
	stats.pos++

	return &queryPhase{
		pb: pb,
	}, true
}

func (stats *queryStats) QueryPhases() xiter.Seq[QueryPhase] {
	return func(yield func(p QueryPhase) bool) {
		for _, pb := range stats.pb.GetQueryPhases() {
			cont := yield(&queryPhase{
				pb: pb,
			})
			if !cont {
				return
			}
		}
	}
}

// NextTableAccess returns next accessed table within query execution phase.
//
// If ok flag is false, then there are no more accessed tables and t is
// invalid.
func (phase *queryPhase) NextTableAccess() (t *TableAccess, ok bool) {
	if phase.pos >= len(phase.pb.GetTableAccess()) {
		return
	}
	pb := phase.pb.GetTableAccess()[phase.pos]
	phase.pos++

	return &TableAccess{
		Name:            pb.GetName(),
		Reads:           fromOperationStats(pb.GetReads()),
		Updates:         fromOperationStats(pb.GetUpdates()),
		Deletes:         fromOperationStats(pb.GetDeletes()),
		PartitionsCount: pb.GetPartitionsCount(),
	}, true
}

func (phase *queryPhase) TableAccess() xiter.Seq[*TableAccess] {
	return func(yield func(access *TableAccess) bool) {
		for _, pb := range phase.pb.GetTableAccess() {
			cont := yield(&TableAccess{
				Name:            pb.GetName(),
				Reads:           fromOperationStats(pb.GetReads()),
				Updates:         fromOperationStats(pb.GetUpdates()),
				Deletes:         fromOperationStats(pb.GetDeletes()),
				PartitionsCount: pb.GetPartitionsCount(),
			})
			if !cont {
				return
			}
		}
	}
}

func (phase *queryPhase) Duration() time.Duration {
	return fromUs(phase.pb.GetDurationUs())
}

func (phase *queryPhase) CPUTime() time.Duration {
	return fromUs(phase.pb.GetCpuTimeUs())
}

func (phase *queryPhase) AffectedShards() uint64 {
	return phase.pb.GetAffectedShards()
}

func (phase *queryPhase) IsLiteralPhase() bool {
	return phase.pb.GetLiteralPhase()
}

func FromQueryStats(pb *Ydb_TableStats.QueryStats) QueryStats {
	if pb == nil {
		return nil
	}

	return &queryStats{
		pb: pb,
	}
}
