package credentials

type SourceInfoOption string

func (sourceInfo SourceInfoOption) ApplyStaticCredentialsOption(h *Static) {
	h.sourceInfo = string(sourceInfo)
}

func (sourceInfo SourceInfoOption) ApplyAnonymousCredentialsOption(h *Anonymous) {
	h.sourceInfo = string(sourceInfo)
}

func (sourceInfo SourceInfoOption) ApplyAccessTokenCredentialsOption(h *AccessToken) {
	h.sourceInfo = string(sourceInfo)
}

func (sourceInfo SourceInfoOption) ApplyOauth2CredentialsOption(h *oauth2TokenExchange) error {
	h.sourceInfo = string(sourceInfo)

	return nil
}

// WithSourceInfo option append to credentials object the source info for reporting source info details on error case
func WithSourceInfo(sourceInfo string) SourceInfoOption {
	return SourceInfoOption(sourceInfo)
}
