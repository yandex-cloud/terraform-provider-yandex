// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Dmitry Novikov <novikoff@yandex-team.ru>

package sdkresolvers

import (
	"fmt"
	"reflect"

	"github.com/yandex-cloud/go-sdk/pkg/sdkerrors"
)

const DefaultResolverPageSize = 100

func CreateResolverFilter(nameField string, value string) string {
	// TODO(novikoff): should we add escaping or value validation?
	return fmt.Sprintf(`%s = "%s"`, nameField, value)
}

type resolveOptions struct {
	out      *string
	folderID string
	cloudID  string
}

type ResolveOption func(*resolveOptions)

func Out(out *string) func(*resolveOptions) {
	return func(o *resolveOptions) {
		o.out = out
	}
}

// FolderID specifies folder id for resolvers that need it (most of the resolvers).
func FolderID(folderID string) ResolveOption {
	return func(o *resolveOptions) {
		o.folderID = folderID
	}
}

// CloudID specifies cloud id for resolvers that need it, e.g. FolderResolver
func CloudID(cloudID string) ResolveOption {
	return func(o *resolveOptions) {
		o.cloudID = cloudID
	}
}

func combineOpts(opts ...ResolveOption) *resolveOptions {
	o := &resolveOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

type BaseResolver struct {
	Name string

	id   string
	err  error
	opts *resolveOptions
}

func NewBaseResolver(name string, opts ...ResolveOption) BaseResolver {
	return BaseResolver{
		Name: name,
		opts: combineOpts(opts...),
	}
}

func (r *BaseResolver) ID() string {
	return r.id
}
func (r *BaseResolver) Err() error {
	return r.err
}

func (r *BaseResolver) SetErr(err error) error {
	if r.err != nil {
		panic(fmt.Sprintf("Trying to change error. Old: %v; New: %v", r.err, err))
	}
	r.err = err
	return r.err
}

func (r *BaseResolver) SetID(id string) {
	r.id = id
	r.writeOut()
}

func (r *BaseResolver) Set(entity Entity, err error) error {
	if err != nil {
		return r.SetErr(err)
	}
	r.SetID(entity.GetId())
	return nil
}

type Entity interface {
	GetId() string
}

func (r *BaseResolver) FolderID() string {
	return r.opts.folderID
}

func (r *BaseResolver) CloudID() string {
	return r.opts.cloudID
}

func (r *BaseResolver) writeOut() {
	if r.opts.out != nil {
		*r.opts.out = r.id
	}
}

func (r *BaseResolver) findName(caption string, slice interface{}, err error) error {
	return r.SetErr(r.findNameImpl(caption, slice, err))
}

type ErrNotFound struct {
	Caption string
	Name    string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%v with name \"%v\" not found", e.Caption, e.Name)
}

func (r *BaseResolver) findNameImpl(caption string, slice interface{}, err error) error {
	if err != nil {
		return sdkerrors.WithMessagef(err, "failed to find %v with name \"%v\"", caption, r.Name)
	}
	rv := reflect.ValueOf(slice)
	var found nameAndID
	for i := 0; i < rv.Len(); i++ {
		v := rv.Index(i).Interface().(nameAndID)
		if v.GetName() == r.Name {
			if found != nil {
				return fmt.Errorf("multiple %v items with name \"%v\" found", caption, r.Name)
			}
			found = v
		}
	}
	if found == nil {
		return &ErrNotFound{Caption: caption, Name: r.Name}
	}
	r.SetID(found.GetId())
	return nil
}

type nameAndID interface {
	GetId() string
	GetName() string
}
