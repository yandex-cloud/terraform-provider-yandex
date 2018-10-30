// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkerrors

import (
	"google.golang.org/grpc/status"
)

type operationError struct {
	err         error
	operationID string
}

func (e *operationError) Error() string {
	return e.err.Error() + "; operation-id: " + e.operationID
}

func (e *operationError) Cause() error {
	return e.err
}

func (e *operationError) GRPCStatus() *status.Status {
	s, ok := status.FromError(e.err)
	if ok {
		return s
	}
	return nil
}

func WithOperationID(x error, operationID string) error {
	if x == nil {
		return nil
	}
	return &operationError{x, operationID}
}

type Causer interface {
	// github.com/pkg/errors
	Cause() error
}

func Cause(err error) error {
	return Visit(err, func(error) bool {
		return true
	})
}

func Visit(err error, visitor func(err error) bool) error {
	for {
		if !visitor(err) {
			break
		}
		causer, ok := err.(Causer)
		if !ok {
			break
		}
		err = causer.Cause()
	}
	return err
}

func OperationID(err error) string {
	id := ""
	Visit(err, func(x error) bool {
		withID, ok := x.(*operationError)
		if !ok {
			return true
		}
		id = withID.operationID
		return false
	})
	return id
}
