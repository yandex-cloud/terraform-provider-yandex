package log

import (
	"go.uber.org/zap"
)

type LogInjector interface {
	// InjectLogger injects the provided zap.Logger instance to enhance logging capabilities within the implementing object.
	InjectLogger(logger *zap.Logger)
}
