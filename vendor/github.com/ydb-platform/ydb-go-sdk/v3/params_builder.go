package ydb

import "github.com/ydb-platform/ydb-go-sdk/v3/internal/params"

// Params is an interface of parameters
// which returns from ydb.ParamsBuilder().Build()
type Params = params.Parameters

// ParamsBuilder used for create query arguments instead of tons options.
func ParamsBuilder() params.Builder {
	return params.Builder{}
}
