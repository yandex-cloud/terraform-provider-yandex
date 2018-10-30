// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkerrors

import (
	"fmt"

	"google.golang.org/grpc/status"
)

type withMessage struct {
	err     error
	message string
}

func (e *withMessage) Error() string {
	return e.message + ": " + e.err.Error()
}

func (e *withMessage) Cause() error {
	return e.err
}

func (e *withMessage) GRPCStatus() *status.Status {
	s, ok := status.FromError(e.err)
	if ok {
		return s
	}
	return nil
}

func WithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &withMessage{err, message}
}

func WithMessagef(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return &withMessage{err, fmt.Sprintf(format, args...)}
}
