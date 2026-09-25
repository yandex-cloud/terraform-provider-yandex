package rawydb

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

type Operation struct {
	ID     string
	Ready  bool
	Status StatusCode
	Issues Issues
}

func (o *Operation) FromProto(proto *Ydb_Operations.Operation) error {
	o.ID = proto.GetId()
	o.Ready = proto.GetReady()
	if err := o.Status.FromProto(proto.GetStatus()); err != nil {
		return err
	}

	return o.Issues.FromProto(proto.GetIssues())
}

func (o *Operation) OperationStatusToError() error {
	if !o.Status.IsSuccess() {
		return xerrors.WithStackTrace(fmt.Errorf("ydb: create topic error [%v]: %v", o.Status, o.Issues))
	}

	return nil
}

func (o *Operation) FromProtoWithStatusCheck(proto *Ydb_Operations.Operation) error {
	if err := o.FromProto(proto); err != nil {
		return err
	}

	return o.OperationStatusToError()
}
