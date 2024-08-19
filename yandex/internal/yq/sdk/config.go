package sdk

import (
	"crypto/tls"
)

type Config struct {
	AuthToken string
	FolderID  string
	TLSConfig *tls.Config
	Endpoint  string
	Plaintext bool
}
