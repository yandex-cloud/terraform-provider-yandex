package table

import (
	"errors"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	errNilClient = xerrors.Wrap(errors.New("table client is not initialized"))

	// errClosedClient returned by a Client instance to indicate
	// that Client is closed early and not able to complete requested operation.
	errClosedClient = xerrors.Wrap(errors.New("table client closed early"))

	// errParamsRequired returned by a Client instance to indicate that required params is not defined
	errParamsRequired = xerrors.Wrap(errors.New("params required"))
)
