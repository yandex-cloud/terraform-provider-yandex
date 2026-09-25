package xtable

import (
	"database/sql/driver"
	"errors"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xsql/badconn"
)

var (
	ErrUnsupported     = driver.ErrSkip
	errConnClosedEarly = badconn.New("conn closed early")
	errNotReadyConn    = badconn.New("conn not ready")
	ErrWrongQueryMode  = errors.New("wrong query mode")
)
