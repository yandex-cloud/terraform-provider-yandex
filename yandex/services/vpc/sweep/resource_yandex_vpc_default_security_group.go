package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func registerResourceYandexVpcDefaultSecurityGroup(add func(string, *resource.Sweeper)) {
	add("yandex_vpc_default_security_group", &resource.Sweeper{
		Name:         "yandex_vpc_default_security_group",
		F:            sweepAllVPCSecurityGroups,
		Dependencies: getYandexVPCSecurityGroupSweeperDeps(),
	})
}
