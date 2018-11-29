// Copyright (c) 2018 Yandex LLC. All rights reserved.
// Author: Vladimir Skipor <skipor@yandex-team.ru>

package ycsdk

import (
	"time"

	iampb "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
)

const (
	// DefaultIAMTokenRefreshInterval is recommended expiration time for created IAM tokens.
	DefaultIAMTokenRefreshInterval = time.Hour

	// OAuthTokenGetIAMTokenExpiration is expiration time of IAM token received for given OAuth token.
	// See https://cloud.yandex.ru/docs/iam/concepts/authorization/iam-token for details.
	IAMTokenFromOAuthExpiration = 12 * time.Hour
)

// Credentials is an abstraction of API authorization credentials.
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/authorization for details.
// Note, that functions that return Credentials may return different Credentials implementation
// in next SDK version, and this is not considered breaking change.
type Credentials interface {
	// YandexCloudAPICredentials is a marker method. All compatible Credentials implementations have it
	YandexCloudAPICredentials()
}

// ExchangeableCredentials can be exchanged for IAM Token in IAM Token Service, that can be used
// to authorize API calls.
// For now, this is the only option to authorize API calls, but this may be changed in future.
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/iam-token for details.
type ExchangeableCredentials interface {
	Credentials
	// IAMTokenRequest returns request for fresh IAM token and token expiration (time to live) or error.
	// SDK will refresh IAM token after returned expiration duration.
	// In case of zero expiration duration IAM tokens will not be cached.
	IAMTokenRequest() (iamTokenReq *iampb.CreateIamTokenRequest, iamTokenExpiration time.Duration, err error)
}

// OAuthToken returns API credentials for user Yandex Passport OAuth token, that can be received
// on page https://oauth.yandex.ru/authorize?response_type=token&client_id=1a6990aa636648e9b2ef855fa7bec2fb
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/oauth-token for details.
func OAuthToken(token string) Credentials {
	return OAuthTokenWithCustomRefreshInterval(token, DefaultIAMTokenRefreshInterval)
}

// OAuthTokenWithCustomRefreshInterval is like OAuthToken but with custom refresh interval.
func OAuthTokenWithCustomRefreshInterval(token string, refreshInterval time.Duration) Credentials {
	if refreshInterval > IAMTokenFromOAuthExpiration {
		refreshInterval = IAMTokenFromOAuthExpiration
	}
	return exchangeableCredentialsFunc(func() (*iampb.CreateIamTokenRequest, time.Duration, error) {
		return &iampb.CreateIamTokenRequest{
			Identity: &iampb.CreateIamTokenRequest_YandexPassportOauthToken{
				YandexPassportOauthToken: token,
			},
		}, refreshInterval, nil
	})
}

type exchangeableCredentialsFunc func() (iamTokenReq *iampb.CreateIamTokenRequest, iamTokenExpiration time.Duration, err error)

func (exchangeableCredentialsFunc) YandexCloudAPICredentials() {}

func (f exchangeableCredentialsFunc) IAMTokenRequest() (iamTokenReq *iampb.CreateIamTokenRequest, iamTokenExpiration time.Duration, err error) {
	return f()
}
