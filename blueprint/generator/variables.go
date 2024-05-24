package generator

import (
	"fmt"
	"strings"
)

type variablesGenerator func(service, resource string, skipComments bool) any

var templateVariables = map[string]variablesGenerator{
	"resource-iam_member": resourceIamVars,
}

func variablesForTemplate(tplType, tplName, service, resource string, skipComments bool) any {
	return templateVariables[fmt.Sprintf("%s-%s", tplType, tplName)](service, resource, skipComments)
}

func resourceIamVars(service, resource string, skipComments bool) any {
	return struct {
		PackageName       string
		ServiceName       string
		PublicPackageName string
		SDKPath           string
		TipIncluded       bool
	}{
		PackageName:       resource,
		ServiceName:       service,
		PublicPackageName: toTitle(resource),
		SDKPath:           getSdkPath(service, resource),
		TipIncluded:       !skipComments,
	}
}

func getSdkPath(service, resource string) string {
	return fmt.Sprintf("SDK.%s().%s()", toTitle(service), toTitle(resource))
}

func toTitle(s string) string {
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
