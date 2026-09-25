package authentication

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-sdk/v2/pkg/authentication"
)

// authSubject defines an interface for entities capable of creating IAM tokens and providing a unique hash identifier.
type authSubject interface {
	createIAMToken(ctx context.Context, a authentication.Authenticator) (authentication.IamToken, error)
	hash() string
}

// userAccount represents an entity for authenticated user actions, including generating IAM tokens and managing identity.
type userAccount struct {
	hash_ string
}

// serviceAccount represents a service account with an ID and a hash used for generating and managing IAM tokens.
type serviceAccount struct {
	id    string
	hash_ string
}

// createIAMToken generates an IAM token using the provided Authenticator and context, returning the token or an error.
func (s userAccount) createIAMToken(ctx context.Context, a authentication.Authenticator) (authentication.IamToken, error) {
	return a.CreateIAMToken(ctx)
}

// hash generates and returns a unique identifier string for the userAccount if it hasn't been generated already.
func (s userAccount) hash() string {
	if s.hash_ == "" {
		s.hash_ = uuid.New().String()
	}

	return s.hash_
}

// createIAMToken generates an IAM token for the current service account using the provided authenticator.
func (s serviceAccount) createIAMToken(ctx context.Context, a authentication.Authenticator) (authentication.IamToken, error) {
	return a.CreateIAMTokenForServiceAccount(ctx, s.id)
}

// hash generates and returns a unique identifier string for a service account if it hasn't already been set.
func (s serviceAccount) hash() string {
	if s.hash_ == "" {
		s.hash_ = uuid.New().String()
	}

	return s.hash_
}

// SAGetter is a function type that retrieves a service account ID from the given context and may return an error.
type SAGetter func(ctx context.Context) (string, error)

// withServiceAccountID is a gRPC call option containing logic to retrieve and use a service account ID for delegation.
type withServiceAccountID struct {
	grpc.EmptyCallOption
	serviceAccountIDGet SAGetter
}

// callAuthSubject resolves and returns an authSubject depending on the context and provided options.
func callAuthSubject(ctx context.Context, originalSubject bool, os []grpc.CallOption) (authSubject, error) {
	if originalSubject {
		return userAccount{}, nil
	}

	var saOpt *withServiceAccountID

	for _, o := range os {
		o, ok := o.(*withServiceAccountID)
		if ok {
			saOpt = o
		}
	}

	var subject authSubject = userAccount{}

	if saOpt != nil {
		sa, err := saOpt.serviceAccountIDGet(ctx)
		if err != nil {
			return nil, fmt.Errorf("error getting SA for delegation: %v", err)
		}

		subject = serviceAccount{
			id: sa,
		}
	}

	return subject, nil
}
