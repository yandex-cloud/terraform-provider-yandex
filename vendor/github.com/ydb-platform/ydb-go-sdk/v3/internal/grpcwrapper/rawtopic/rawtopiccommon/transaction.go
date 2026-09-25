package rawtopiccommon

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Topic"

type TransactionIdentity struct {
	ID      string
	Session string
}

func (t TransactionIdentity) ToProto() *Ydb_Topic.TransactionIdentity {
	if t.ID == "" && t.Session == "" {
		return nil
	}

	return &Ydb_Topic.TransactionIdentity{
		Id:      t.ID,
		Session: t.Session,
	}
}
