// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Maxim Kolganov <manykey@yandex-team.ru>

package sdk

import (
	"context"
	"net/url"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

const (
	iamTokenHeaderName = "Authorization"
)

type subjectTokenCreds struct {
	mutex sync.RWMutex // guards currentState, conn and excludes multiple simultaneous token updates

	getConn lazyConn
	conn    *grpc.ClientConn // initialized lazily from getConn

	conf     Config
	now      func() time.Time
	tokenTTL time.Duration

	currentState credsState
}

type credsState struct {
	token      string
	expiration time.Time
	version    int64
}

func creds(conf Config) *subjectTokenCreds {
	return &subjectTokenCreds{
		conf: conf,
	}
}

func (c *subjectTokenCreds) Init(lazyConn lazyConn, now func() time.Time, tokenTTL time.Duration) {
	c.getConn = lazyConn
	c.now = now
	c.tokenTTL = tokenTTL
}

func (c *subjectTokenCreds) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	audienceURL, err := url.Parse(uri[0])
	if err != nil {
		return nil, err
	}
	if audienceURL.Path == "/yandex.cloud.iam.v1.IamTokenService" ||
		audienceURL.Path == "/yandex.cloud.endpoint.ApiEndpointService" {
		return nil, nil
	}

	c.mutex.RLock()
	state := c.currentState
	c.mutex.RUnlock()

	token := state.token
	outdated := state.expiration.Before(c.now())
	if outdated {
		token, err = c.updateToken(ctx, state)
		if err != nil {
			return nil, err
		}
	}

	return map[string]string{
		iamTokenHeaderName: "Bearer " + token,
	}, nil
}

func (c *subjectTokenCreds) RequireTransportSecurity() bool {
	return !c.conf.Plaintext
}

func (c *subjectTokenCreds) updateToken(ctx context.Context, currentState credsState) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.currentState.version != currentState.version {
		// someone have already updated it
		return c.currentState.token, nil
	}

	if c.conn == nil {
		conn, err := c.getConn(ctx)
		if err != nil {
			return "", err
		}
		c.conn = conn
	}
	tokenClient := iam.NewIamTokenServiceClient(c.conn)
	resp, err := tokenClient.Create(ctx, &iam.CreateIamTokenRequest{
		Identity: &iam.CreateIamTokenRequest_YandexPassportOauthToken{
			YandexPassportOauthToken: c.conf.OAuthToken,
		},
	})
	if err != nil {
		return "", err
	}

	c.currentState = credsState{
		token:      resp.IamToken,
		expiration: c.now().Add(c.tokenTTL),
		version:    currentState.version + 1,
	}

	return c.currentState.token, nil
}
