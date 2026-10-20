package yandex_cic_private_connection_test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_private_connection"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestCICPrivateConnection_schema(t *testing.T) {
	var resp frameworkresource.SchemaResponse
	yandex_cic_private_connection.NewResource().Schema(context.Background(), frameworkresource.SchemaRequest{}, &resp)
	attributes := resp.Schema.Attributes
	for _, name := range []string{"id", "private_connection_id", "status", "created_at"} {
		if !attributes[name].IsComputed() || attributes[name].IsOptional() {
			t.Errorf("%s must be computed only", name)
		}
	}
	for _, name := range []string{"rt", "rd", "routing_policy", "cloud_id", "acl"} {
		if _, exists := attributes[name]; exists {
			t.Errorf("internal attribute %s is exposed", name)
		}
	}
	peering := attributes["ipv4_peering"].(schema.SingleNestedAttribute)
	if !peering.Attributes["peer_bgp_md5_key"].IsSensitive() {
		t.Error("BGP MD5 key must be sensitive")
	}
	if !peering.Attributes["cloud_bgp_asn"].IsComputed() || peering.Attributes["cloud_bgp_asn"].IsOptional() {
		t.Error("cloud_bgp_asn must be computed only")
	}
	routes := attributes["ipv4_static_routes"].(schema.SetNestedAttribute)
	if _, exists := routes.NestedObject.Attributes["next_hop"]; exists {
		t.Error("internal next_hop is exposed")
	}
}

func TestCICPrivateConnection_unknownValues(t *testing.T) {
	config := `
resource "terraform_data" "input" {
  input = "connection"
}
resource "yandex_cic_private_connection" "test" {
  name = terraform_data.input.output
  trunk_connection_id = "trunk-id"
  ipv4_static_routes = []
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: privateConnectionTestFactories(t, &privateConnectionAPIStub{}),
		Steps: []resource.TestStep{
			{Config: config},
			{Config: config, PlanOnly: true},
			{Config: strings.ReplaceAll(config, `input = "connection"`, `input = "renamed"`)},
		},
	})
}

func TestCICPrivateConnection_moveErrors(t *testing.T) {
	for _, operationFailure := range []bool{false, true} {
		t.Run(fmt.Sprintf("operation=%t", operationFailure), func(t *testing.T) {
			api := &privateConnectionAPIStub{}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: privateConnectionTestFactories(t, api),
				CheckDestroy:             api.checkDestroyed,
				Steps: []resource.TestStep{
					{Config: privateConnectionTestConfig("connection", "")},
					{
						PreConfig: func() {
							api.mu.Lock()
							defer api.mu.Unlock()
							api.moveError = status.Error(codes.FailedPrecondition, "move rejected")
							api.operationFailure = operationFailure
						},
						Config:      privateConnectionTestConfig("connection", `folder_id = "other-folder"`),
						ExpectError: regexp.MustCompile("(?s)(Failed to Move resource|Unable to Move Resource).*move rejected"),
					},
				},
			})
		})
	}
}

func TestCICPrivateConnection_updateTimeout(t *testing.T) {
	api := &privateConnectionAPIStub{waitForUpdateTimeout: true}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: privateConnectionTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{Config: privateConnectionTestConfig("initial", "")},
			{
				Config:      privateConnectionTestConfig("updated", `timeouts = { update = "100ms" }`),
				ExpectError: regexp.MustCompile("(?s)Unable to Update Resource.*context deadline exceeded"),
			},
		},
	})
}

func TestCICPrivateConnection_omittedFields(t *testing.T) {
	api := &privateConnectionAPIStub{}
	initial := privateConnectionTestConfig("connection", privateConnectionTestPeering+`
  ipv4_static_routes = [{ prefix = "10.10.0.0/24" }]
`)
	metadata := privateConnectionTestConfig("connection", `description = "changed"`)
	peering := strings.ReplaceAll(privateConnectionTestPeering, "65001", "65002")
	peering = strings.ReplaceAll(peering, `peer_bgp_md5_key = "test-secret"`, "")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: privateConnectionTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{Config: initial},
			{
				Config: metadata,
				Check: resource.ComposeTestCheckFunc(
					api.checkUpdateMask("description"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_static_routes.#", "1"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "vlan_id", "101"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_peering.peer_bgp_md5_key", "test-secret"),
				),
			},
			{Config: metadata, PlanOnly: true},
			{
				Config: privateConnectionTestConfig("connection", peering+`description = "changed"`),
				Check: resource.ComposeTestCheckFunc(
					api.checkUpdateMask("ipv4_peering.peer_bgp_asn"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_peering.peer_bgp_md5_key", "test-secret"),
				),
			},
		},
	})
}
