package balancer

import (
	"context"

	balancerConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/balancer/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/conn"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/endpoint"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xrand"
)

type connectionsState struct {
	connByNodeID map[uint32]conn.Conn

	prefer   []conn.Conn
	fallback []conn.Conn
	all      []conn.Conn

	rand xrand.Rand
}

func newConnectionsState(
	conns []conn.Conn,
	filter balancerConfig.Filter,
	info balancerConfig.Info,
	allowFallback bool,
) *connectionsState {
	res := &connectionsState{
		connByNodeID: connsToNodeIDMap(conns),
		rand:         xrand.New(xrand.WithLock()),
	}

	res.prefer, res.fallback = sortPreferConnections(conns, filter, info, allowFallback)
	if allowFallback {
		res.all = conns
	} else {
		res.all = res.prefer
	}

	return res
}

func (s *connectionsState) PreferredCount() int {
	return len(s.prefer)
}

func (s *connectionsState) All() (all []endpoint.Endpoint) {
	if s == nil {
		return nil
	}

	all = make([]endpoint.Endpoint, len(s.all))
	for i, c := range s.all {
		all[i] = c.Endpoint()
	}

	return all
}

func (s *connectionsState) GetConnection(ctx context.Context) (_ conn.Conn, failedCount int) {
	if err := ctx.Err(); err != nil {
		return nil, 0
	}

	if c := s.preferConnection(ctx); c != nil {
		return c, 0
	}

	try := func(conns []conn.Conn) conn.Conn {
		c, tryFailed := s.selectRandomConnection(conns, false)
		failedCount += tryFailed

		return c
	}

	if c := try(s.prefer); c != nil {
		return c, failedCount
	}

	if c := try(s.fallback); c != nil {
		return c, failedCount
	}

	c, _ := s.selectRandomConnection(s.all, true)

	return c, failedCount
}

func (s *connectionsState) preferConnection(ctx context.Context) conn.Conn {
	if nodeID, hasPreferEndpoint := endpoint.ContextNodeID(ctx); hasPreferEndpoint {
		c := s.connByNodeID[nodeID]
		if c != nil && isOkConnection(c, false) {
			return c
		}
	}

	return nil
}

func (s *connectionsState) selectRandomConnection(conns []conn.Conn, allowBanned bool) (c conn.Conn, failedConns int) {
	connCount := len(conns)
	if connCount == 0 {
		// return for empty list need for prevent panic in fast path
		return nil, 0
	}

	// fast path
	if c := conns[s.rand.Int(connCount)]; isOkConnection(c, allowBanned) {
		return c, 0
	}

	// shuffled indexes slices need for guarantee about every connection will check
	indexes := make([]int, connCount)
	for index := range indexes {
		indexes[index] = index
	}
	s.rand.Shuffle(connCount, func(i, j int) {
		indexes[i], indexes[j] = indexes[j], indexes[i]
	})

	for _, index := range indexes {
		c := conns[index]
		if isOkConnection(c, allowBanned) {
			return c, 0
		}
		failedConns++
	}

	return nil, failedConns
}

func connsToNodeIDMap(conns []conn.Conn) (nodes map[uint32]conn.Conn) {
	if len(conns) == 0 {
		return nil
	}
	nodes = make(map[uint32]conn.Conn, len(conns))
	for _, c := range conns {
		nodes[c.Endpoint().NodeID()] = c
	}

	return nodes
}

func sortPreferConnections(
	conns []conn.Conn,
	filter balancerConfig.Filter,
	info balancerConfig.Info,
	allowFallback bool,
) (prefer, fallback []conn.Conn) {
	if filter == nil {
		return conns, nil
	}

	prefer = make([]conn.Conn, 0, len(conns))
	if allowFallback {
		fallback = make([]conn.Conn, 0, len(conns))
	}

	for _, c := range conns {
		if filter.Allow(info, c.Endpoint()) {
			prefer = append(prefer, c)
		} else if allowFallback {
			fallback = append(fallback, c)
		}
	}

	return prefer, fallback
}

func isOkConnection(c conn.Conn, bannedIsOk bool) bool {
	switch c.GetState() {
	case conn.Online, conn.Created, conn.Offline:
		return true
	case conn.Banned:
		return bannedIsOk
	default:
		return false
	}
}
