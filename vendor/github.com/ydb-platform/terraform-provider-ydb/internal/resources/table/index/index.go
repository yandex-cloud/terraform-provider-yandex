package index

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ydb-platform/ydb-go-sdk/v3/table/options"

	"github.com/ydb-platform/terraform-provider-ydb/internal/helpers"
	"github.com/ydb-platform/terraform-provider-ydb/internal/resources"
	"github.com/ydb-platform/terraform-provider-ydb/sdk/terraform/auth"
)

type handler struct {
	authCreds auth.YdbCredentials
}

type resource struct {
	TablePath        string
	TableEntity      *helpers.YDBEntity
	ConnectionString string
	Name             string
	Type             string
	Columns          []string
	Cover            []string
	Entity           *helpers.YDBEntity
}

func indexResourceSchemaToIndexResource(d *schema.ResourceData) (*resource, error) {
	var entity *helpers.YDBEntity
	var err error
	if d.Id() != "" {
		entity, err = helpers.ParseYDBEntityID(d.Id())
		if err != nil {
			return nil, fmt.Errorf("failed to parse index entity: %w", err)
		}
	}

	var tableEntity *helpers.YDBEntity
	if tableID, ok := d.GetOk("table_id"); ok {
		en, err := helpers.ParseYDBEntityID(tableID.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to parse table_id: %w", err)
		}
		tableEntity = en
	}

	tablePath := d.Get("table_path").(string)
	connectionString := d.Get("connection_string").(string)
	name := d.Get("name").(string)
	if name == "" && entity != nil {
		name = parseIndexNameFromIndexEntity(entity.GetFullEntityPath())
	}
	typ := d.Get("type").(string)
	colsRaw := d.Get("columns").([]interface{})
	colsArr := make([]string, 0, len(colsRaw))
	for _, c := range colsRaw {
		colsArr = append(colsArr, c.(string))
	}

	var coverArr []string
	if cover, ok := d.GetOk("cover"); ok {
		for _, c := range cover.([]interface{}) {
			coverArr = append(coverArr, c.(string))
		}
	}

	return &resource{
		TablePath:        tablePath,
		TableEntity:      tableEntity,
		ConnectionString: connectionString,
		Name:             name,
		Type:             typ,
		Columns:          colsArr,
		Cover:            coverArr,
		Entity:           entity,
	}, nil
}

func (r *resource) getConnectionString() string {
	// NOTE(shmel1k@): ConnectionString is set only when no `table_id` is present.
	if r.ConnectionString != "" {
		return r.ConnectionString
	}
	if r.TableEntity != nil {
		return r.TableEntity.PrepareFullYDBEndpoint()
	}
	return r.Entity.PrepareFullYDBEndpoint()
}

func (r *resource) getTablePath() string {
	// NOTE(shmel1k@): TablePath is set only when no `table_id` is present.
	if r.TablePath != "" {
		return r.TablePath
	}
	if r.TableEntity != nil {
		return r.TableEntity.GetEntityPath()
	}
	return parseTablePathFromIndexEntity(r.Entity.GetEntityPath())
}

func NewHandler(authCreds auth.YdbCredentials) resources.Handler {
	return &handler{
		authCreds: authCreds,
	}
}

func flattenIndexDescription(
	d *schema.ResourceData,
	indexResource *resource,
	indexDescription options.IndexDescription,
) (err error) {
	err = d.Set("table_path", indexResource.getTablePath())
	if err != nil {
		return
	}
	err = d.Set("connection_string", indexResource.getConnectionString())
	if err != nil {
		return
	}
	err = d.Set("table_id", indexResource.getConnectionString()+"?path="+indexResource.getTablePath())
	if err != nil {
		return
	}
	err = d.Set("name", indexDescription.Name)
	if err != nil {
		return
	}
	err = d.Set("type", getIndexType(indexDescription.Type))
	if err != nil {
		return
	}
	cols := make([]interface{}, 0, len(indexDescription.IndexColumns))
	for _, c := range indexDescription.IndexColumns {
		cols = append(cols, c)
	}
	err = d.Set("columns", cols)
	if err != nil {
		return
	}
	covers := make([]interface{}, 0, len(indexDescription.DataColumns))
	for _, c := range indexDescription.DataColumns {
		covers = append(covers, c)
	}

	return d.Set("cover", covers)
}

func parseTablePathFromIndexEntity(entityPath string) string {
	split := strings.Split(entityPath, "/")
	return strings.Join(split[:len(split)-1], "/")
}

func parseIndexNameFromIndexEntity(entityPath string) string {
	split := strings.Split(entityPath, "/")
	return split[len(split)-1]
}

func getIndexType(index options.IndexType) string {
	switch index {
	case 0:
		return "global_sync"
	case 1:
		return "global_async"
	}
	return ""
}
