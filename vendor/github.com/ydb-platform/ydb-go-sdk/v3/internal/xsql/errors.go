package xsql

import (
	"database/sql/driver"
	"errors"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
)

var (
	ErrUnsupported         = driver.ErrSkip
	errDeprecated          = driver.ErrSkip
	errWrongQueryProcessor = errors.New("wrong query processor")
	errNotReadyConn        = badconn.New("conn not ready")
)
