package scheme

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Scheme"
)

func permissions(p Permissions) *Ydb_Scheme.Permissions {
	var y Ydb_Scheme.Permissions
	p.To(&y)

	return &y
}

type permissionsDesc interface {
	SetClear(clear bool)
	AppendAction(action *Ydb_Scheme.PermissionsAction)
}

type PermissionsOption func(permissionsDesc)

func WithClearPermissions() PermissionsOption {
	return func(p permissionsDesc) {
		p.SetClear(true)
	}
}

func WithGrantPermissions(p Permissions) PermissionsOption {
	return func(d permissionsDesc) {
		d.AppendAction(&Ydb_Scheme.PermissionsAction{
			Action: &Ydb_Scheme.PermissionsAction_Grant{
				Grant: permissions(p),
			},
		})
	}
}

func WithRevokePermissions(p Permissions) PermissionsOption {
	return func(d permissionsDesc) {
		d.AppendAction(&Ydb_Scheme.PermissionsAction{
			Action: &Ydb_Scheme.PermissionsAction_Revoke{
				Revoke: permissions(p),
			},
		})
	}
}

func WithSetPermissions(p Permissions) PermissionsOption {
	return func(d permissionsDesc) {
		d.AppendAction(&Ydb_Scheme.PermissionsAction{
			Action: &Ydb_Scheme.PermissionsAction_Set{
				Set: permissions(p),
			},
		})
	}
}

func WithChangeOwner(owner string) PermissionsOption {
	return func(d permissionsDesc) {
		d.AppendAction(&Ydb_Scheme.PermissionsAction{
			Action: &Ydb_Scheme.PermissionsAction_ChangeOwner{
				ChangeOwner: owner,
			},
		})
	}
}
