package yandex

import (
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_FederatedQuery"
)

const (
	attributeVisibilityPrivateValue = "PRIVATE"
	attributeVisibilityScopeValue   = "SCOPE"

	attributeServiceAccountIDNoneValue = ""
)

func parseServiceIDToIAMAuth(serviceAccountID string) *Ydb_FederatedQuery.IamAuth {
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

func iAMAuthToString(auth *Ydb_FederatedQuery.IamAuth) (string, error) {
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

func parseVisibilityToAclVisibility(visibility string) (Ydb_FederatedQuery.Acl_Visibility, error) {
	switch visibility {
	case attributeVisibilityPrivateValue:
		return Ydb_FederatedQuery.Acl_PRIVATE, nil
	case attributeVisibilityScopeValue:
		return Ydb_FederatedQuery.Acl_SCOPE, nil
	default:
		return Ydb_FederatedQuery.Acl_PRIVATE, nil
	}
}

func visibilityToString(visibility Ydb_FederatedQuery.Acl_Visibility) (string, error) {
	switch visibility {
	case Ydb_FederatedQuery.Acl_PRIVATE:
		return attributeVisibilityPrivateValue, nil
	case Ydb_FederatedQuery.Acl_SCOPE:
		return attributeVisibilityScopeValue, nil
	case Ydb_FederatedQuery.Acl_VISIBILITY_UNSPECIFIED:
		return "", fmt.Errorf("visivility is \"unspecified\", unsupported type")
	default:
		return "", fmt.Errorf("unsupported visibility type: %d", visibility)
	}
}
