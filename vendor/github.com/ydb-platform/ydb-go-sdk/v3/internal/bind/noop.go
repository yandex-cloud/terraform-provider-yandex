package bind

type noopBind struct{}

func (m noopBind) ToYdb(sql string, args ...any) (yql string, newArgs []any, _ error) {
	return sql, args, nil
}

func (m noopBind) blockID() blockID {
	return blockCastArgs
}
