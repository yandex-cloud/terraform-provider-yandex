package yandex_cic_private_connection_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cic "github.com/yandex-cloud/go-genproto/yandex/cloud/cic/v1"
)

func TestCICPrivateConnection_maskedLogging(t *testing.T) {
	const initialSecret = "cic-secret-initial\"\\\n雪"
	const rotatedSecret = "cic-secret-rotated\\\"\t"
	const publicSecret = "cic-secret-public\"\\"
	api := &privateConnectionAPIStub{returnSecrets: true, publicConnection: &cic.PublicConnection{
		Id: "public-id", Ipv4Peering: &cic.Peering{PeerBgpMd5Key: publicSecret},
	}}
	logPath := filepath.Join(t.TempDir(), "%s.log")
	t.Setenv("TF_LOG", "TRACE")
	t.Setenv("TF_LOG_SDK_HELPER_RESOURCE", "WARN")
	t.Setenv("TF_LOG_PATH_MASK", logPath)
	t.Setenv("TF_ACC_LOG_PATH", "")
	factories := privateConnectionTestFactories(t, api)
	factory := factories["yandex"]
	factories["yandex"] = func() (tfprotov6.ProviderServer, error) {
		server, err := factory()
		return &cicLoggingTestServer{ProviderServer: server}, err
	}
	configuration := func(secret string) string {
		peering := strings.ReplaceAll(privateConnectionTestPeering, `"test-secret"`, fmt.Sprintf("%q", secret))
		return privateConnectionTestConfig("connection", peering) + `
data "yandex_cic_private_connection" "test" {
  private_connection_id = yandex_cic_private_connection.test.id
}
data "yandex_cic_public_connection" "test" {
  public_connection_id = "public-id"
}
`
	}
	checkAPIKey := func(expected string) resource.TestCheckFunc {
		return func(_ *terraform.State) error {
			api.mu.Lock()
			defer api.mu.Unlock()
			if api.connection.GetIpv4Peering().GetPeerBgpMd5Key() != expected {
				return fmt.Errorf("API received a changed BGP MD5 key")
			}
			return nil
		}
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: factories,
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{Config: configuration(initialSecret), Check: checkAPIKey(initialSecret)},
			{Config: configuration(rotatedSecret), Check: checkAPIKey(rotatedSecret)},
			{Config: configuration(rotatedSecret), PlanOnly: true},
			{ResourceName: privateConnectionResourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"ipv4_peering.peer_bgp_md5_key"}},
		},
	})
	logs, err := os.ReadFile(fmt.Sprintf(logPath, strings.ReplaceAll(t.Name(), "/", "__")))
	if err != nil {
		t.Fatal(err)
	}
	output := string(logs)
	messages := strings.Split(output, "\n")
	if strings.Contains(output, "cic-secret-") {
		t.Fatal("BGP MD5 key reached the provider log")
	}
	for _, expected := range []struct{ rpc, prefix string }{
		{"ApplyResourceChange", "Create private_connection request:"},
		{"ApplyResourceChange", "Update private_connection request:"},
		{"ReadResource", "Read private_connection response:"},
		{"ReadDataSource", "Read private_connection response:"},
		{"ReadDataSource", "Read public_connection response:"},
	} {
		found := false
		for _, message := range messages {
			if strings.Contains(message, "test_rpc="+expected.rpc) && strings.Contains(message, expected.prefix) && strings.Contains(message, "***") {
				found = true
			}
		}
		if !found {
			t.Errorf("missing masked log entry for %s: %s", expected.rpc, expected.prefix)
		}
	}
	if !strings.Contains(output, "private-connection-1") || !strings.Contains(output, "public-id") {
		t.Error("non-secret resource IDs were lost from logs")
	}
}

type cicLoggingTestServer struct {
	tfprotov6.ProviderServer
}

func (s *cicLoggingTestServer) ApplyResourceChange(ctx context.Context, req *tfprotov6.ApplyResourceChangeRequest) (*tfprotov6.ApplyResourceChangeResponse, error) {
	ctx = tflog.SetField(ctx, "test_rpc", "ApplyResourceChange")
	return s.ProviderServer.ApplyResourceChange(ctx, req)
}

func (s *cicLoggingTestServer) ReadResource(ctx context.Context, req *tfprotov6.ReadResourceRequest) (*tfprotov6.ReadResourceResponse, error) {
	ctx = tflog.SetField(ctx, "test_rpc", "ReadResource")
	return s.ProviderServer.ReadResource(ctx, req)
}

func (s *cicLoggingTestServer) ReadDataSource(ctx context.Context, req *tfprotov6.ReadDataSourceRequest) (*tfprotov6.ReadDataSourceResponse, error) {
	ctx = tflog.SetField(ctx, "test_rpc", "ReadDataSource")
	return s.ProviderServer.ReadDataSource(ctx, req)
}
