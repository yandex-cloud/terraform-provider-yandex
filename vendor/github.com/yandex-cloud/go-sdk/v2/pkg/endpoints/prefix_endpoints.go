package endpoints

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/yandex-cloud/go-sdk/v2/pkg/errors"
)

// PrefixToEndpoint maps a protobuf FullName prefix to its EndpointParams.
type PrefixToEndpoint map[protoreflect.FullName]*EndpointParams

// Get returns the EndpointParams for the longest matching prefix of p.
// It walks up the namespace hierarchy until it finds a match or returns nil.
func (p2e PrefixToEndpoint) Get(p protoreflect.FullName) *EndpointParams {
	for name := p; name != ""; name = name.Parent() {
		if ep, ok := p2e[name]; ok {
			return ep
		}
	}
	return nil
}

// Clone returns a shallow copy of the PrefixToEndpoint map.
func (p2e PrefixToEndpoint) Clone() PrefixToEndpoint {
	out := make(PrefixToEndpoint, len(p2e))
	for k, v := range p2e {
		out[k] = v
	}
	return out
}

// Merge returns a new PrefixToEndpoint containing:
// 1) All entries from the original map that are not overridden by src (including any parent overrides),
// 2) All entries from src (overriding or adding as needed).
func (p2e PrefixToEndpoint) Merge(src PrefixToEndpoint) PrefixToEndpoint {
	result := make(PrefixToEndpoint, len(p2e)+len(src))
	// Copy entries from the original that are not covered by src.
	for k, v := range p2e {
		if src.Get(k) == nil {
			result[k] = v
		}
	}
	// Add or override with all entries from src.
	for k, v := range src {
		result[k] = v
	}
	return result
}

type PrefixEndpointsResolver struct {
	base             PrefixToEndpoint
	prefixToEndpoint map[protoreflect.FullName]*Endpoint
	cache            sync.Map // map[protoreflect.FullName]*Endpoint
}

func NewPrefixEndpointsResolver(p2e PrefixToEndpoint) *PrefixEndpointsResolver {
	m := make(map[protoreflect.FullName]*Endpoint, len(p2e))

	for p, params := range p2e {
		m[p] = params.Build()
	}
	return &PrefixEndpointsResolver{prefixToEndpoint: m, base: p2e}
}

func (r *PrefixEndpointsResolver) Endpoint(_ context.Context, method protoreflect.FullName, _ ...grpc.CallOption) (*Endpoint, error) {
	if v, ok := r.cache.Load(method); ok {
		return v.(*Endpoint), nil
	}

	for name := method; name != ""; name = name.Parent() {
		if e, ok := r.prefixToEndpoint[name]; ok {
			r.cache.Store(method, e)
			return e, nil
		}
	}

	return nil, &errors.EndpointNotFoundError{Method: method}
}

func (r *PrefixEndpointsResolver) PrefixToEndpoint() PrefixToEndpoint {
	return r.base.Clone()
}
