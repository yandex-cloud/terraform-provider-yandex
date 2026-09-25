package xcontext

import (
	"context"
	"sync"
)

type (
	CancelsGuard struct {
		mu      sync.Mutex
		cancels map[*context.CancelFunc]struct{}
	}
)

func NewCancelsGuard() *CancelsGuard {
	return &CancelsGuard{
		cancels: make(map[*context.CancelFunc]struct{}),
	}
}

func (g *CancelsGuard) WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	g.mu.Lock()
	defer g.mu.Unlock()
	ctx, cancel := WithCancel(ctx)
	g.cancels[&cancel] = struct{}{}

	return ctx, func() {
		cancel()
		g.mu.Lock()
		defer g.mu.Unlock()
		delete(g.cancels, &cancel)
	}
}

func (g *CancelsGuard) Cancel() {
	g.mu.Lock()
	defer g.mu.Unlock()
	for cancel := range g.cancels {
		(*cancel)()
	}
	g.cancels = make(map[*context.CancelFunc]struct{})
}
