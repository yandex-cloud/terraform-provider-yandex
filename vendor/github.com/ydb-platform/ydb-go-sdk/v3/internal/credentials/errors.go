package credentials

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	grpcCodes "google.golang.org/grpc/codes"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
	"github.com/ydb-platform/ydb-go-sdk/v3/pkg/xstring"
)

var errNilExpiresAt = errors.New("nil claims.ExpiresAt field")

type authErrorOption interface {
	applyAuthErrorOption(w io.Writer)
}

var (
	_ authErrorOption = nodeIDAuthErrorOption(0)
	_ authErrorOption = addressAuthErrorOption("")
	_ authErrorOption = endpointAuthErrorOption("")
	_ authErrorOption = databaseAuthErrorOption("")
)

type addressAuthErrorOption string

func (address addressAuthErrorOption) applyAuthErrorOption(w io.Writer) {
	fmt.Fprint(w, "address:")
	fmt.Fprintf(w, "%q", address)
}

func WithAddress(address string) addressAuthErrorOption {
	return addressAuthErrorOption(address)
}

type endpointAuthErrorOption string

func (endpoint endpointAuthErrorOption) applyAuthErrorOption(w io.Writer) {
	fmt.Fprint(w, "endpoint:")
	fmt.Fprintf(w, "%q", endpoint)
}

func WithEndpoint(endpoint string) endpointAuthErrorOption {
	return endpointAuthErrorOption(endpoint)
}

type databaseAuthErrorOption string

func (address databaseAuthErrorOption) applyAuthErrorOption(w io.Writer) {
	fmt.Fprint(w, "database:")
	fmt.Fprintf(w, "%q", address)
}

func WithDatabase(database string) databaseAuthErrorOption {
	return databaseAuthErrorOption(database)
}

type nodeIDAuthErrorOption uint32

func (id nodeIDAuthErrorOption) applyAuthErrorOption(w io.Writer) {
	fmt.Fprint(w, "nodeID:")
	fmt.Fprint(w, strconv.FormatUint(uint64(id), 10))
}

func WithNodeID(id uint32) authErrorOption {
	return nodeIDAuthErrorOption(id)
}

type credentialsUnauthenticatedErrorOption struct {
	credentials Credentials
}

func (opt credentialsUnauthenticatedErrorOption) applyAuthErrorOption(w io.Writer) {
	fmt.Fprint(w, "credentials:")
	if stringer, has := opt.credentials.(fmt.Stringer); has {
		fmt.Fprintf(w, "%q", stringer.String())
	} else {
		t := reflect.TypeOf(opt.credentials)
		fmt.Fprintf(w, "%q", t.PkgPath()+"."+t.Name())
	}
}

func WithCredentials(credentials Credentials) credentialsUnauthenticatedErrorOption {
	return credentialsUnauthenticatedErrorOption{
		credentials: credentials,
	}
}

func AccessError(msg string, err error, opts ...authErrorOption) error {
	buffer := xstring.Buffer()
	defer buffer.Free()
	buffer.WriteString(msg)
	buffer.WriteString(" (")
	for i, opt := range opts {
		if opt != nil {
			if i != 0 {
				buffer.WriteString(",")
			}
			opt.applyAuthErrorOption(buffer)
		}
	}
	buffer.WriteString("): %w")

	return xerrors.WithStackTrace(fmt.Errorf(buffer.String(), err), xerrors.WithSkipDepth(1))
}

func IsAccessError(err error) bool {
	if xerrors.IsTransportError(err,
		grpcCodes.Unauthenticated,
		grpcCodes.PermissionDenied,
	) {
		return true
	}
	if xerrors.IsOperationError(err,
		Ydb.StatusIds_UNAUTHORIZED,
	) {
		return true
	}

	return false
}
