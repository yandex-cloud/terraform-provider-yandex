package api

import "time"

const (
	YandexAirflowClusterCreateTimeout = 30 * time.Minute
	YandexAirflowClusterDeleteTimeout = 15 * time.Minute
	YandexAirflowClusterUpdateTimeout = 60 * time.Minute

	AdminPasswordStubOnImport = "<real value unknown because resource was imported>"
)
