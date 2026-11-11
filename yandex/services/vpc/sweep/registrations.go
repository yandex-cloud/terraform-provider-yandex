package sweep

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func Register(add func(string, *resource.Sweeper)) {
	registerResourceYandexVpcDefaultSecurityGroup(add)
	registerResourceYandexVpcGateway(add)
	registerResourceYandexVpcNetwork(add)
	registerResourceYandexVpcPrivateEndpoint(add)
	registerResourceYandexVpcRouteTable(add)
	registerResourceYandexVpcSecurityGroup(add)
	registerResourceYandexVpcSubnet(add)
	registerVpcSecurityGroupRuleResource(add)
}
