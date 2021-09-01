package yandex

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/greenplum/v1"
)

func parseGreenplumEnv(e string) (greenplum.Cluster_Environment, error) {
	v, ok := greenplum.Cluster_Environment_value[e]
	if !ok {
		return 0, fmt.Errorf("value for 'environment' must be one of %s, not `%s`",
			getJoinedKeys(getEnumValueMapKeys(greenplum.Cluster_Environment_value)), e)
	}
	return greenplum.Cluster_Environment(v), nil
}
