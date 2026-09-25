package xquery

import (
	"database/sql/driver"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
)

var (
	ErrUnsupported     = driver.ErrSkip
	errDeprecated      = driver.ErrSkip
	errConnClosedEarly = badconn.New("conn closed early")
	errNotReadyConn    = badconn.New("conn not ready")
)
