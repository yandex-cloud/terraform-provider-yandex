package rawscheme

import (
	"errors"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Scheme"

	"github.com/ydb-platform/ydb-go-sdk/v3/internal/xerrors"
)

var (
	errUnexpectedNilForSchemePermissions = xerrors.Wrap(errors.New("ydb: unexpected nil for scheme permissions"))
	errUnexpectedNilForSchemeEntry       = xerrors.Wrap(errors.New("ydb: unexpected nil for scheme entry"))
)

type Entry struct {
	Name                 string
	Owner                string
	Type                 EntryType
	EffectivePermissions []Permissions
	Permissions          []Permissions
	SizeBytes            uint64
}

func (e *Entry) FromProto(proto *Ydb_Scheme.Entry) error {
	if proto == nil {
		return xerrors.WithStackTrace(errUnexpectedNilForSchemeEntry)
	}
	e.Name = proto.GetName()
	e.Owner = proto.GetOwner()
	e.Type = EntryType(proto.GetType())

	e.EffectivePermissions = make([]Permissions, len(proto.GetEffectivePermissions()))
	for i := range proto.GetEffectivePermissions() {
		if err := e.EffectivePermissions[i].FromProto(proto.GetEffectivePermissions()[i]); err != nil {
			return err
		}
	}

	e.Permissions = make([]Permissions, len(proto.GetPermissions()))
	for i := range proto.GetPermissions() {
		if err := e.Permissions[i].FromProto(proto.GetPermissions()[i]); err != nil {
			return err
		}
	}

	e.SizeBytes = proto.GetSizeBytes()

	return nil
}

type EntryType int

const (
	EntryTypeUnspecified      = EntryType(Ydb_Scheme.Entry_TYPE_UNSPECIFIED)
	EntryTypeDirectory        = EntryType(Ydb_Scheme.Entry_DIRECTORY)
	EntryTypeTable            = EntryType(Ydb_Scheme.Entry_TABLE)
	EntryTypePersQueueGroup   = EntryType(Ydb_Scheme.Entry_PERS_QUEUE_GROUP)
	EntryTypeDatabase         = EntryType(Ydb_Scheme.Entry_DATABASE)
	EntryTypeRtmrVolume       = EntryType(Ydb_Scheme.Entry_RTMR_VOLUME)
	EntryTypeBlockStoreVolume = EntryType(Ydb_Scheme.Entry_BLOCK_STORE_VOLUME)
	EntryTypeCoordinationNode = EntryType(Ydb_Scheme.Entry_COORDINATION_NODE)
	EntryTypeSequence         = EntryType(Ydb_Scheme.Entry_SEQUENCE)
	EntryTypeReplication      = EntryType(Ydb_Scheme.Entry_REPLICATION)
	EntryTypeTopic            = EntryType(Ydb_Scheme.Entry_TOPIC)
)

type Permissions struct {
	Subject         string
	PermissionNames []string
}

func (p *Permissions) FromProto(proto *Ydb_Scheme.Permissions) error {
	if proto == nil {
		return xerrors.WithStackTrace(errUnexpectedNilForSchemePermissions)
	}
	p.Subject = proto.GetSubject()
	p.PermissionNames = proto.GetPermissionNames()

	return nil
}
