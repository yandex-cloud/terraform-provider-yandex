package datalens

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestCachedTokenProvider_CachesToken(t *testing.T) {
	t.Parallel()

	var calls int64
	provider := func(ctx context.Context) (string, error) {
		atomic.AddInt64(&calls, 1)
		return "token-1", nil
	}

	cached := NewCachedTokenProvider(provider)

	tok, err := cached.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "token-1" {
		t.Fatalf("expected token-1, got %s", tok)
	}

	tok, err = cached.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "token-1" {
		t.Fatalf("expected token-1, got %s", tok)
	}

	if c := atomic.LoadInt64(&calls); c != 1 {
		t.Fatalf("expected provider to be called once, got %d", c)
	}
}

func TestCachedTokenProvider_RefreshesExpiredToken(t *testing.T) {
	t.Parallel()

	var calls int64
	provider := func(ctx context.Context) (string, error) {
		n := atomic.AddInt64(&calls, 1)
		if n == 1 {
			return "token-old", nil
		}
		return "token-new", nil
	}

	cached := NewCachedTokenProvider(provider)
	cached.ttl = time.Millisecond

	tok, err := cached.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "token-old" {
		t.Fatalf("expected token-old, got %s", tok)
	}

	time.Sleep(10 * time.Millisecond)

	tok, err = cached.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok != "token-new" {
		t.Fatalf("expected token-new, got %s", tok)
	}

	if c := atomic.LoadInt64(&calls); c != 2 {
		t.Fatalf("expected provider to be called twice, got %d", c)
	}
}

func TestCachedToken_IsValid(t *testing.T) {
	t.Parallel()

	var ct *cachedToken
	if ct.isValid() {
		t.Error("nil cachedToken should not be valid")
	}

	ct = &cachedToken{token: "", expiresAt: time.Now().Add(time.Hour)}
	if ct.isValid() {
		t.Error("empty token should not be valid")
	}

	ct = &cachedToken{token: "tok", expiresAt: time.Now().Add(-time.Second)}
	if ct.isValid() {
		t.Error("expired token should not be valid")
	}

	ct = &cachedToken{token: "tok", expiresAt: time.Now().Add(30 * time.Second)}
	if ct.isValid() {
		t.Error("token expiring within safety margin should not be valid")
	}

	ct = &cachedToken{token: "tok", expiresAt: time.Now().Add(2 * time.Minute)}
	if !ct.isValid() {
		t.Error("token with future expiry should be valid")
	}
}
