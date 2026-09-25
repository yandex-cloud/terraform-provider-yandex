package pool

import (
	"context"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/stack"
)

type (
	Trace struct {
		OnNew    func(ctx *context.Context, call stack.Caller) func(limit int)
		OnClose  func(ctx *context.Context, call stack.Caller) func(err error)
		OnTry    func(ctx *context.Context, call stack.Caller) func(err error)
		OnWith   func(ctx *context.Context, call stack.Caller) func(attempts int, err error)
		OnPut    func(ctx *context.Context, call stack.Caller, item any) func(err error)
		OnGet    func(ctx *context.Context, call stack.Caller) func(item any, attempts int, err error)
		onWait   func() func(item any, err error)
		OnChange func(Stats)
	}
)
