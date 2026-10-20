package yandex_cic_public_connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	ciclogging "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cic/logging"
)

type publicConnectionDataSourceWithMaskedLogging struct {
	*yandexCicPublicConnectionDataSource
}

func NewDataSourceWithMaskedLogging() datasource.DataSource {
	return &publicConnectionDataSourceWithMaskedLogging{&yandexCicPublicConnectionDataSource{}}
}

func (r *publicConnectionDataSourceWithMaskedLogging) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	r.yandexCicPublicConnectionDataSource.Read(ciclogging.MaskContext(ctx), req, resp)
}
