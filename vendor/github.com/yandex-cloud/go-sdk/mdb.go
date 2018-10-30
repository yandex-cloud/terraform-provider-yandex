// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdk

import (
	"github.com/yandex-cloud/go-sdk/mdb/clickhouse"
	"github.com/yandex-cloud/go-sdk/mdb/mongodb"
	"github.com/yandex-cloud/go-sdk/mdb/postgresql"
)

const (
	MDBMongoDBServiceID    Endpoint = "mdb-mongodb"
	MDBClickhouseServiceID Endpoint = "mdb-clickhouse"
	MDBPostgreSQLServiceID Endpoint = "mdb-postgresql"
)

type MDB struct {
	sdk *SDK
}

func (m *MDB) PostgreSQL() *postgresql.PostgreSQL {
	return postgresql.NewPostgreSQL(m.sdk.requestContext(MDBPostgreSQLServiceID).getConn)
}

func (m *MDB) MongoDB() *mongodb.MongoDB {
	return mongodb.NewMongoDB(m.sdk.requestContext(MDBMongoDBServiceID).getConn)
}

func (m *MDB) Clickhouse() *clickhouse.Clickhouse {
	return clickhouse.NewClickhouse(m.sdk.requestContext(MDBClickhouseServiceID).getConn)
}
