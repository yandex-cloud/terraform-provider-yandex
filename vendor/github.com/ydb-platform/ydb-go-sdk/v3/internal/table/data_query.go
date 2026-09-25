package table

import (
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Table"
)

type (
	Query interface {
		String() string
		ID() string
		YQL() string

		toYDB() *Ydb_Table.Query
	}
	textQuery     string
	preparedQuery struct {
		id  string
		sql string
	}
)

func (q textQuery) String() string {
	return string(q)
}

func (q textQuery) ID() string {
	return ""
}

func (q textQuery) YQL() string {
	return string(q)
}

func (q textQuery) toYDB() *Ydb_Table.Query {
	return &Ydb_Table.Query{
		Query: &Ydb_Table.Query_YqlText{
			YqlText: string(q),
		},
	}
}

func (q preparedQuery) String() string {
	return q.sql
}

func (q preparedQuery) ID() string {
	return q.id
}

func (q preparedQuery) YQL() string {
	return q.sql
}

func (q preparedQuery) toYDB() *Ydb_Table.Query {
	return &Ydb_Table.Query{
		Query: &Ydb_Table.Query_YqlText{
			YqlText: q.sql,
		},
	}
}

func queryFromText(s string) Query {
	return textQuery(s)
}

func queryPrepared(id, sql string) Query {
	return preparedQuery{
		id:  id,
		sql: sql,
	}
}
