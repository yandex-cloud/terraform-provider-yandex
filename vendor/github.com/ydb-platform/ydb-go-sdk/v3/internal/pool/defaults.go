package pool

import (
	"time"
)

const (
	DefaultLimit         = 50
	defaultCreateTimeout = 5 * time.Second
	defaultCloseTimeout  = time.Second
)
