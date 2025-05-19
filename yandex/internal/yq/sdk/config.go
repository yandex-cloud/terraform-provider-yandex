package sdk

import (
	"context"
	"crypto/tls"
)

type authTokenProviderType func(context.Context) (string, error)

type Config struct {
	AuthTokenProvider authTokenProviderType
	FolderID          string
	TLSConfig         *tls.Config
	Endpoint          string
	Plaintext         bool
}
