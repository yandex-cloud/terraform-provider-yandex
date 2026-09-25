package helpers

import "time"

type ResourceDataProxy interface {
	Get(key string) interface{}
	GetOk(key string) (interface{}, bool)

	// GetOkExists and methods below are bypassed (i.e. call schema.ResourceData directly)
	// Deprecated: calls a deprecated method
	GetOkExists(key string) (interface{}, bool)

	Id() string
	SetId(id string)
	Set(key string, value interface{}) error
	HasChange(key string) bool
	GetChange(key string) (interface{}, interface{})
	Partial(on bool)
	Timeout(s string) time.Duration
}
