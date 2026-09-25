package options

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/query/result"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stats"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/types"
	"github.com/ydb-platform/ydb-go-sdk/v3/trace"
)

type (
	FetchScriptResultsRequest struct {
		Ydb_Query.FetchScriptResultsRequest

		Trace *trace.Query
	}
	FetchScriptOption      func(request *FetchScriptResultsRequest)
	ExecuteScriptOperation struct {
		ID            string
		ConsumedUnits float64
		Metadata      *MetadataExecuteQuery
	}
	FetchScriptResult struct {
		ResultSetIndex int64
		ResultSet      result.Set
		NextToken      string
	}
	MetadataExecuteQuery struct {
		ID     string
		Script struct {
			Syntax Syntax
			Query  string
		}
		Mode           ExecMode
		Stats          stats.QueryStats
		ResultSetsMeta []struct {
			Columns []struct {
				Name string
				Type types.Type
			}
		}
	}
)

func WithFetchToken(fetchToken string) FetchScriptOption {
	return func(request *FetchScriptResultsRequest) {
		request.FetchToken = fetchToken
	}
}

func WithResultSetIndex(resultSetIndex int64) FetchScriptOption {
	return func(request *FetchScriptResultsRequest) {
		request.ResultSetIndex = resultSetIndex
	}
}

func WithRowsLimit(rowsLimit int64) FetchScriptOption {
	return func(request *FetchScriptResultsRequest) {
		request.RowsLimit = rowsLimit
	}
}

func ToMetadataExecuteQuery(metadata *anypb.Any) *MetadataExecuteQuery {
	var pb Ydb_Query.ExecuteScriptMetadata
	if err := metadata.UnmarshalTo(&pb); err != nil {
		panic(err)
	}

	return &MetadataExecuteQuery{
		ID: pb.GetExecutionId(),
		Script: struct {
			Syntax Syntax
			Query  string
		}{
			Syntax: Syntax(pb.GetScriptContent().GetSyntax()),
			Query:  pb.GetScriptContent().GetText(),
		},
		Mode:  ExecMode(pb.GetExecMode()),
		Stats: stats.FromQueryStats(pb.GetExecStats()),
		ResultSetsMeta: func() (
			resultSetsMeta []struct {
				Columns []struct {
					Name string
					Type types.Type
				}
			},
		) {
			for _, rs := range pb.GetResultSetsMeta() {
				resultSetsMeta = append(resultSetsMeta, struct {
					Columns []struct {
						Name string
						Type types.Type
					}
				}{
					Columns: func() (
						columns []struct {
							Name string
							Type types.Type
						},
					) {
						for _, c := range rs.GetColumns() {
							columns = append(columns, struct {
								Name string
								Type types.Type
							}{
								Name: c.GetName(),
								Type: types.TypeFromYDB(c.GetType()),
							})
						}

						return columns
					}(),
				})
			}

			return resultSetsMeta
		}(),
	}
}
