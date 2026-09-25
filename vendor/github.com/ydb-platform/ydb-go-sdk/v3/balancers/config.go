package balancers

import (
	"encoding/json"
	"fmt"

	balancerConfig "github.com/ydb-platform/ydb-go-sdk/v3/internal/balancer/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type balancerType string

const (
	typeRoundRobin   = balancerType("round_robin")
	typeRandomChoice = balancerType("random_choice")
	typeSingle       = balancerType("single")
	typeDisable      = balancerType("disable")
)

type preferType string

const (
	preferTypeNearestDC = preferType("nearest_dc")
	preferTypeLocations = preferType("locations")

	// Deprecated
	// Will be removed after March 2025.
	// Read about versioning policy: https://github.com/ydb-platform/ydb-go-sdk/blob/master/VERSIONING.md#deprecated
	preferTypeLocalDC = preferType("local_dc")
)

type balancersConfig struct {
	Type      balancerType `json:"type"`
	Prefer    preferType   `json:"prefer,omitempty"`
	Fallback  bool         `json:"fallback,omitempty"`
	Locations []string     `json:"locations,omitempty"`
}

type fromConfigOptionsHolder struct {
	fallbackBalancer *balancerConfig.Config
	errorHandler     func(error)
}

type fromConfigOption func(h *fromConfigOptionsHolder)

func WithParseErrorFallbackBalancer(b *balancerConfig.Config) fromConfigOption {
	return func(h *fromConfigOptionsHolder) {
		h.fallbackBalancer = b
	}
}

func WithParseErrorHandler(errorHandler func(error)) fromConfigOption {
	return func(h *fromConfigOptionsHolder) {
		h.errorHandler = errorHandler
	}
}

func createByType(t balancerType) (*balancerConfig.Config, error) {
	switch t {
	case typeDisable:
		return SingleConn(), nil
	case typeSingle:
		return SingleConn(), nil
	case typeRandomChoice:
		return RandomChoice(), nil
	case typeRoundRobin:
		return RoundRobin(), nil
	default:
		return nil, xerrors.WithStackTrace(fmt.Errorf("unknown type of balancer: %s", t))
	}
}

func CreateFromConfig(s string) (*balancerConfig.Config, error) {
	// try to parse s as identifier of balancer
	if c, err := createByType(balancerType(s)); err == nil {
		return c, nil
	}

	var (
		b   *balancerConfig.Config
		err error
		c   balancersConfig
	)

	// try to parse s as json
	if err = json.Unmarshal([]byte(s), &c); err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	b, err = createByType(c.Type)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	switch c.Prefer {
	case preferTypeLocalDC:
		if c.Fallback {
			return PreferNearestDCWithFallBack(b), nil
		}

		return PreferNearestDC(b), nil
	case preferTypeNearestDC:
		if c.Fallback {
			return PreferNearestDCWithFallBack(b), nil
		}

		return PreferNearestDC(b), nil
	case preferTypeLocations:
		if len(c.Locations) == 0 {
			return nil, xerrors.WithStackTrace(fmt.Errorf("empty locations list in balancer '%s' config", c.Type))
		}
		if c.Fallback {
			return PreferLocationsWithFallback(b, c.Locations...), nil
		}

		return PreferLocations(b, c.Locations...), nil
	default:
		return b, nil
	}
}

func FromConfig(config string, opts ...fromConfigOption) *balancerConfig.Config {
	var (
		h = fromConfigOptionsHolder{
			fallbackBalancer: Default(),
		}
		b   *balancerConfig.Config
		err error
	)
	for _, opt := range opts {
		if opt != nil {
			opt(&h)
		}
	}

	b, err = CreateFromConfig(config)
	if err != nil {
		if h.errorHandler != nil {
			h.errorHandler(err)
		}

		return h.fallbackBalancer
	}

	return b
}
