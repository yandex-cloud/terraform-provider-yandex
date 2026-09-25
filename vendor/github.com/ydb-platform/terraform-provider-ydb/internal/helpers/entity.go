package helpers

import (
	"fmt"
	"strings"
)

type YDBEntity struct {
	databaseEndpoint string
	database         string
	entityPath       string
	useTLS           bool
}

func (y *YDBEntity) PrepareFullYDBEndpoint() string {
	prefix := "grpc://"
	if y.useTLS {
		prefix = "grpcs://"
	}
	return prefix + y.databaseEndpoint + "/?database=" + y.database
}

func (y *YDBEntity) GetFullEntityPath() string {
	return y.database + "/" + y.entityPath
}

func (y *YDBEntity) GetEntityPath() string {
	return y.entityPath
}

func (y *YDBEntity) GetDatabasePath() string {
	return y.database
}

func (y *YDBEntity) ID() string {
	return y.PrepareFullYDBEndpoint() + "?path=" + y.entityPath
}

func ParseYDBEntityID(id string) (*YDBEntity, error) {
	if id == "" {
		return nil, fmt.Errorf("failed to parse ydb entity id: %s", "got empty id")
	}

	endpoint, database, useTLS, err := ParseYDBDatabaseEndpoint(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ydb entity id: %w", err)
	}

	slashCount := 0
	i := 0
	split := strings.Split(database, "?path=")
	if len(split) > 1 {
		return &YDBEntity{
			databaseEndpoint: endpoint,
			database:         split[0],
			entityPath:       split[1],
			useTLS:           useTLS,
		}, nil
	}

	for i = 0; i < len(database); i++ {
		if database[i] == '/' {
			slashCount++
		}
		// NOTE(shmel1k@): /pre-prod_ydb_public/abacaba/babacaba/
		if slashCount == 4 {
			break
		}
	}
	if i == len(database) || i == len(database)-1 || slashCount < 4 {
		return nil, fmt.Errorf("failed to parse ydb entity id: %s", "got empty entity path")
	}

	return &YDBEntity{
		databaseEndpoint: endpoint,
		database:         database[:i],
		entityPath:       database[i+1:],
		useTLS:           useTLS,
	}, nil
}
