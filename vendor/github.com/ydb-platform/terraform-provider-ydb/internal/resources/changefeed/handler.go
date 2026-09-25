package changefeed

import (
	"github.com/ydb-platform/terraform-provider-ydb/internal/resources"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
)

type handler struct {
	authCreds auth.YdbCredentials
}

func NewHandler(authCreds auth.YdbCredentials) resources.Handler {
	return &handler{
		authCreds: authCreds,
	}
}
