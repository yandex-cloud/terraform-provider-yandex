package yqcommon

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/yandex-cloud/terraform-provider-yandex/common"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

// descriptions
const (
	attributeServiceAccountIDNoneValue = ""
)

var (
	availableConnectionAttributes = map[string]schema.Attribute{
		AttributeBucket: schema.StringAttribute{
			MarkdownDescription: "The bucket name from ObjectStorage.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		AttributeCloudID: schema.StringAttribute{
			MarkdownDescription: "The cloud identifier.",
			Computed:            true,
		},
		AttributeFolderID: schema.StringAttribute{
			MarkdownDescription: "The folder identifier.",
			Optional:            true,
			Computed:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		AttributeDatabaseID: schema.StringAttribute{
			MarkdownDescription: "The database identifier.",
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		AttributeSharedReading: schema.BoolAttribute{
			MarkdownDescription: "Whether to enable shared reading by different queries from the same connection.",
			Optional:            true,
			Computed:            true,
			Default:             booldefault.StaticBool(false),
			PlanModifiers: []planmodifier.Bool{
				boolplanmodifier.UseStateForUnknown(),
			},
		},
	}
)

func ParseServiceIDToIAMAuth(serviceAccountID string) *Ydb_FederatedQuery.IamAuth {
	switch serviceAccountID {
	case attributeServiceAccountIDNoneValue:
		return &Ydb_FederatedQuery.IamAuth{
			Identity: &Ydb_FederatedQuery.IamAuth_None{},
		}
	default:
		return &Ydb_FederatedQuery.IamAuth{
			Identity: &Ydb_FederatedQuery.IamAuth_ServiceAccount{
				ServiceAccount: &Ydb_FederatedQuery.ServiceAccountAuth{
					Id: serviceAccountID,
				},
			},
		}
	}
}

func IAMAuthToString(auth *Ydb_FederatedQuery.IamAuth) (string, error) {
	switch auth.Identity.(type) {
	case *Ydb_FederatedQuery.IamAuth_None:
		return attributeServiceAccountIDNoneValue, nil
	case *Ydb_FederatedQuery.IamAuth_ServiceAccount:
		serviceAccountAuth := auth.Identity.(*Ydb_FederatedQuery.IamAuth_ServiceAccount)
		return serviceAccountAuth.ServiceAccount.GetId(), nil
	case *Ydb_FederatedQuery.IamAuth_CurrentIam:
		return "", fmt.Errorf("unsupported auth type: current IAM")
	default:
		return "", fmt.Errorf("unknown auth type")
	}
}

func NewConnectionResourceSchema(additionalAttributes ...string) map[string]schema.Attribute {
	result := map[string]schema.Attribute{
		AttributeID: schema.StringAttribute{
			MarkdownDescription: common.ResourceDescriptions["id"],
			Computed:            true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		AttributeName: schema.StringAttribute{
			MarkdownDescription: common.ResourceDescriptions["name"],
			Required:            true,
			Validators: []validator.String{
				stringvalidator.LengthAtLeast(1),
			},
		},
		AttributeServiceAccountID: schema.StringAttribute{
			MarkdownDescription: "The service account ID to access resources on behalf of.",
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		AttributeDescription: schema.StringAttribute{
			MarkdownDescription: common.ResourceDescriptions["description"],
			Optional:            true,
			Computed:            true,
			Default:             stringdefault.StaticString(""),
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
	}

	for _, a := range additionalAttributes {
		b := availableConnectionAttributes[a]
		if b == nil {
			panic(fmt.Sprintf("Additional attribute %v for connection not found", b))
		}
		result[a] = b
	}

	return result
}
