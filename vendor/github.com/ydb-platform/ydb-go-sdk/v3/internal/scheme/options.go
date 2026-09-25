package scheme

import "github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Scheme"

type permissionsDesc struct {
	clear   bool
	actions []*Ydb_Scheme.PermissionsAction
}

func (p *permissionsDesc) SetClear(clear bool) {
	p.clear = clear
}

func (p *permissionsDesc) AppendAction(action *Ydb_Scheme.PermissionsAction) {
	p.actions = append(p.actions, action)
}
