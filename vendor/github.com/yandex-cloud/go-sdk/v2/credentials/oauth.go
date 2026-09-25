package credentials

// OAuthToken returns API credentials for user Yandex Passport OAuth token, that can be received
// on page https://oauth.yandex.ru/authorize?response_type=token&client_id=1a6990aa636648e9b2ef855fa7bec2fb
// See https://cloud.yandex.ru/docs/iam/concepts/authorization/oauth-token for details.
//
// Deprecated: Please consider to use other credential provider. By the end of 2026, the use of oauth tokens in the Yandex cloud will be discontinued.
func OAuthToken(token string) ExchangeableCredentials {
	return exchangeableCredentialsFunc(func() (*CredentialsTokenRequest, error) {
		return &CredentialsTokenRequest{
			Identity: CredentialsIdentityYandexPassportOauthToken,
			Token:    token,
		}, nil
	})
}
