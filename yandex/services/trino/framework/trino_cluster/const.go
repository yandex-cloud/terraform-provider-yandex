package trino_cluster

import "time"

const (
	YandexTrinoClusterCreateTimeout = 30 * time.Minute
	YandexTrinoClusterDeleteTimeout = 15 * time.Minute
	YandexTrinoClusterUpdateTimeout = 60 * time.Minute
)
