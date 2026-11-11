package yandex_cic_private_connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestCICDataSources(t *testing.T) {
	for _, entity := range []string{"partner", "point_of_presence", "trunk_connection", "public_connection"} {
		t.Run(entity, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: privateConnectionTestFactories(t, &privateConnectionAPIStub{}),
				Steps: []resource.TestStep{{
					Config: fmt.Sprintf(`data "yandex_cic_%s" "test" { %s_id = "test-id" }`, entity, entity),
					Check:  resource.TestCheckResourceAttr("data.yandex_cic_"+entity+".test", "id", "test-id"),
				}},
			})
		})
	}
}

func TestCICPrivateConnectionDataSource(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: privateConnectionTestFactories(t, &privateConnectionAPIStub{}),
		Steps: []resource.TestStep{
			{
				Config: privateConnectionTestConfig("connection", privateConnectionTestPeering) + `
data "yandex_cic_private_connection" "test" {
  private_connection_id = yandex_cic_private_connection.test.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.yandex_cic_private_connection.test", "id", privateConnectionResourceName, "id"),
					resource.TestCheckResourceAttr("data.yandex_cic_private_connection.test", "ipv4_peering.peer_bgp_asn", "65001"),
					resource.TestCheckNoResourceAttr("data.yandex_cic_private_connection.test", "ipv4_peering.peer_bgp_md5_key"),
				),
			},
		},
	})
}
