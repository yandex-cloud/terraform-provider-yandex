package trino_catalog

import "time"

const (
	YandexTrinoCatalogCreateTimeout = 10 * time.Minute
	YandexTrinoCatalogDeleteTimeout = 5 * time.Minute
	YandexTrinoCatalogUpdateTimeout = 10 * time.Minute

	PasswordStubOnImport = "<real value unknown because resource was imported>"
)
