package log

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/opensearch/v1"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/services/mdb_opensearch_cluster/model"
	"google.golang.org/protobuf/proto"
)

const AdminPasswordAttribute = "admin_password"
const RedactedPassword = "[REDACTED]"

func RedactModel(ctx context.Context, source *model.OpenSearch) *model.OpenSearch {
	if source == nil {
		return nil
	}

	result := *source
	result.Config = RedactAdminPassword(ctx, source.Config)
	return &result
}

func RedactAdminPassword(ctx context.Context, source types.Object) types.Object {
	if source.IsNull() || source.IsUnknown() {
		return source
	}

	attributeTypes := source.AttributeTypes(ctx)
	attributes := make(map[string]attr.Value, len(source.Attributes()))
	maps.Copy(attributes, source.Attributes())

	value, ok := attributes[AdminPasswordAttribute]
	if ok && !value.IsNull() && !value.IsUnknown() {
		attributes[AdminPasswordAttribute] = types.StringValue(RedactedPassword)
	}

	config, diags := types.ObjectValue(attributeTypes, attributes)
	if diags.HasError() {
		return types.ObjectNull(attributeTypes)
	}

	return config
}

func RedactAdminPasswordList(ctx context.Context, source types.List) types.List {
	if source.IsNull() || source.IsUnknown() {
		return source
	}

	elementType := source.ElementType(ctx)
	elements := make([]attr.Value, 0, len(source.Elements()))
	for _, element := range source.Elements() {
		config, ok := element.(types.Object)
		if !ok {
			return types.ListNull(elementType)
		}
		elements = append(elements, RedactAdminPassword(ctx, config))
	}

	result, diags := types.ListValue(elementType, elements)
	if diags.HasError() {
		return types.ListNull(elementType)
	}

	return result
}

func RedactCreateClusterRequest(source *opensearch.CreateClusterRequest) *opensearch.CreateClusterRequest {
	if source == nil {
		return nil
	}

	result := proto.Clone(source).(*opensearch.CreateClusterRequest)
	if result.ConfigSpec != nil && result.ConfigSpec.AdminPassword != "" {
		result.ConfigSpec.AdminPassword = RedactedPassword
	}
	return result
}

func RedactUpdateClusterRequest(source *opensearch.UpdateClusterRequest) *opensearch.UpdateClusterRequest {
	if source == nil {
		return nil
	}

	result := proto.Clone(source).(*opensearch.UpdateClusterRequest)
	if result.ConfigSpec != nil && result.ConfigSpec.AdminPassword != "" {
		result.ConfigSpec.AdminPassword = RedactedPassword
	}
	return result
}
