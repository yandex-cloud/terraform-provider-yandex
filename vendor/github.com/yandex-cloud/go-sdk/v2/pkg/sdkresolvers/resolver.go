package sdkresolvers

import (
	"context"
)

type Resolver interface {
	ID() string
	Err() error

	Run(context.Context) error
}
