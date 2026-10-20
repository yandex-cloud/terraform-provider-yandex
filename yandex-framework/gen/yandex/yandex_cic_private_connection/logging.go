package yandex_cic_private_connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	ciclogging "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/cic/logging"
)

type privateConnectionResourceWithMaskedLogging struct {
	*yandexCicPrivateConnectionResource
}

func NewResourceWithMaskedLogging() resource.Resource {
	return &privateConnectionResourceWithMaskedLogging{&yandexCicPrivateConnectionResource{}}
}

func (r *privateConnectionResourceWithMaskedLogging) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	r.yandexCicPrivateConnectionResource.Create(ciclogging.MaskContext(ctx), req, resp)
}

func (r *privateConnectionResourceWithMaskedLogging) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	r.yandexCicPrivateConnectionResource.Read(ciclogging.MaskContext(ctx), req, resp)
}

func (r *privateConnectionResourceWithMaskedLogging) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.yandexCicPrivateConnectionResource.Update(ciclogging.MaskContext(ctx), req, resp)
}

func (r *privateConnectionResourceWithMaskedLogging) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	r.yandexCicPrivateConnectionResource.Delete(ciclogging.MaskContext(ctx), req, resp)
}

type privateConnectionDataSourceWithMaskedLogging struct {
	*yandexCicPrivateConnectionDataSource
}

func NewDataSourceWithMaskedLogging() datasource.DataSource {
	return &privateConnectionDataSourceWithMaskedLogging{&yandexCicPrivateConnectionDataSource{}}
}

func (r *privateConnectionDataSourceWithMaskedLogging) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	r.yandexCicPrivateConnectionDataSource.Read(ciclogging.MaskContext(ctx), req, resp)
}
