package certificates

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"sync"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	// fileCache stores certificates by file name
	fileCache sync.Map
	// pemCache stores certificates by pem cache
	pemCache sync.Map
)

type (
	fromFileOptions struct {
		onHit   func()
		onMiss  func()
		noCache bool
	}
	FromFileOption func(opts *fromFileOptions)
)

func FromFileOnMiss(onMiss func()) FromFileOption {
	return func(opts *fromFileOptions) {
		opts.onMiss = onMiss
	}
}

func FromFileOnHit(onHit func()) FromFileOption {
	return func(opts *fromFileOptions) {
		opts.onHit = onHit
	}
}

func loadFromFileCache(key string) (_ []*x509.Certificate, exists bool) {
	value, exists := fileCache.Load(key)
	if !exists {
		return nil, false
	}
	certs, ok := value.([]*x509.Certificate)
	if !ok {
		panic(fmt.Sprintf("unexpected value type '%T'", value))
	}

	return certs, true
}

// FromFile reads and parses pem-encoded certificate(s) from file.
func FromFile(file string, opts ...FromFileOption) ([]*x509.Certificate, error) {
	options := fromFileOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	if !options.noCache {
		certs, exists := loadFromFileCache(file)
		if exists {
			if options.onHit != nil {
				options.onHit()
			}

			return certs, nil
		}
	}

	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	certs, err := FromPem(bytes,
		FromPemNoCache(true), // no use pem cache - certs stored in file cache only
	)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	if !options.noCache {
		fileCache.Store(file, certs)
		if options.onMiss != nil {
			options.onMiss()
		}
	}

	return certs, nil
}

func loadFromPemCache(key string) (_ *x509.Certificate, exists bool) {
	value, exists := pemCache.Load(key)
	if !exists {
		return nil, false
	}
	cert, ok := value.(*x509.Certificate)
	if !ok {
		panic(fmt.Sprintf("unexpected value type '%T'", value))
	}

	return cert, true
}

// parseCertificate is a cached version of x509.ParseCertificate. Cache key is string(der)
func parseCertificate(der []byte, opts ...FromPemOption) (*x509.Certificate, error) {
	options := fromPemOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	key := string(der)

	if !options.noCache {
		cert, exists := loadFromPemCache(key)
		if exists {
			if options.onHit != nil {
				options.onHit()
			}

			return cert, nil
		}
	}

	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, xerrors.WithStackTrace(err)
	}

	if !options.noCache {
		pemCache.Store(key, cert)
		if options.onMiss != nil {
			options.onMiss()
		}
	}

	return cert, nil
}

type (
	fromPemOptions struct {
		onHit   func()
		onMiss  func()
		noCache bool
	}
	FromPemOption func(opts *fromPemOptions)
)

func FromPemMiss(onMiss func()) FromPemOption {
	return func(opts *fromPemOptions) {
		opts.onMiss = onMiss
	}
}

func FromPemOnHit(onHit func()) FromPemOption {
	return func(opts *fromPemOptions) {
		opts.onHit = onHit
	}
}

func FromPemNoCache(noCache bool) FromPemOption {
	return func(opts *fromPemOptions) {
		opts.noCache = noCache
	}
}

// FromPem parses one or more certificate from pem blocks in bytes.
// It returns nil error if at least one certificate was successfully parsed.
// This function uses cached parseCertificate.
func FromPem(bytes []byte, opts ...FromPemOption) (certs []*x509.Certificate, err error) {
	var block *pem.Block

	for len(bytes) > 0 {
		block, bytes = pem.Decode(bytes)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		var cert *x509.Certificate
		cert, err = parseCertificate(block.Bytes, opts...)
		if err != nil {
			continue
		}
		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, xerrors.WithStackTrace(err)
	}

	return certs, nil
}
