package yandex_cic_private_connection_test

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	frameworkprovider "github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	cic "github.com/yandex-cloud/go-genproto/yandex/cloud/cic/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	ycsdk "github.com/yandex-cloud/go-sdk/v2"
	"github.com/yandex-cloud/go-sdk/v2/credentials"
	"github.com/yandex-cloud/go-sdk/v2/pkg/endpoints"
	"github.com/yandex-cloud/go-sdk/v2/pkg/options"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_partner"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_point_of_presence"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_private_connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_public_connection"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/gen/yandex/yandex_cic_trunk_connection"
	providerconfig "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const privateConnectionResourceName = "yandex_cic_private_connection.test"

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func TestCICPrivateConnection_lifecycle(t *testing.T) {
	api := &privateConnectionAPIStub{}
	initial := privateConnectionTestConfig("connection-initial", privateConnectionTestPeering+`
  description = "initial description"
  labels = { stage = "initial" }
  ipv4_static_routes = [{ prefix = "10.10.0.0/24" }, { prefix = "10.20.0.0/24" }]
`)
	updated := privateConnectionTestConfig("connection-updated", privateConnectionTestPeering+`
  description = ""
  labels = {}
  ipv4_static_routes = [{ prefix = "10.20.0.0/24" }, { prefix = "10.10.0.0/24" }]
`)
	cleared := privateConnectionTestConfig("connection-updated", privateConnectionTestPeering+`
  description = ""
  labels = {}
  ipv4_static_routes = []
`)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: privateConnectionTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: initial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(privateConnectionResourceName, "folder_id", "folder-id"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "vlan_id", "101"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_peering.cloud_bgp_asn", "200350"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_peering.peer_bgp_md5_key", "test-secret"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_static_routes.#", "2"),
					api.checkMutations(1, 0, 0),
				),
			},
			{ResourceName: privateConnectionResourceName, ImportState: true, ImportStateVerify: true, ImportStateVerifyIgnore: []string{"ipv4_peering.peer_bgp_md5_key"}},
			{Config: initial, PlanOnly: true},
			{
				Config: updated,
				Check: resource.ComposeTestCheckFunc(
					api.checkUpdateMask("description", "labels", "name"),
					api.checkMutations(1, 1, 0),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_static_routes.#", "2"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_peering.peer_bgp_md5_key", "test-secret"),
				),
			},
			{
				Config: cleared,
				Check: resource.ComposeTestCheckFunc(
					api.checkUpdateMask("ipv4_static_routes"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_static_routes.#", "0"),
				),
			},
			{
				Config: strings.ReplaceAll(cleared, "test-secret", "rotated-secret"),
				Check: resource.ComposeTestCheckFunc(
					api.checkUpdateMask("ipv4_peering.peer_bgp_md5_key"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "ipv4_peering.peer_bgp_md5_key", "rotated-secret"),
				),
			},
			{
				Config:   strings.ReplaceAll(cleared, "test-secret", "rotated-secret") + "\n",
				PlanOnly: true,
			},
			{
				Config: privateConnectionTestConfig("connection-updated", strings.ReplaceAll(privateConnectionTestPeering, "test-secret", "rotated-secret")+`
  folder_id = "other-folder"
  ipv4_static_routes = []
`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(privateConnectionResourceName, "id", "private-connection-1"),
					resource.TestCheckResourceAttr(privateConnectionResourceName, "folder_id", "other-folder"),
					api.checkMutations(1, 3, 1),
				),
			},
			{
				PreConfig: func() {
					api.mu.Lock()
					defer api.mu.Unlock()
					api.connection = nil
				},
				Config: cleared,
				Check:  api.checkMutations(2, 3, 1),
			},
		},
	})
}

func TestCICPrivateConnection_optionalPeering(t *testing.T) {
	api := &privateConnectionAPIStub{}
	config := privateConnectionTestConfig("connection-static", "ipv4_static_routes = []")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: privateConnectionTestFactories(t, api),
		CheckDestroy:             api.checkDestroyed,
		Steps: []resource.TestStep{
			{Config: config},
			{Config: config, PlanOnly: true},
			{ResourceName: privateConnectionResourceName, ImportState: true, ImportStateVerify: true},
			{Config: strings.ReplaceAll(config, "connection-static", "connection-renamed"), Check: api.checkUpdateMask("name")},
		},
	})
}

func TestCICPrivateConnection_updateErrors(t *testing.T) {
	for _, operationFailure := range []bool{false, true} {
		t.Run(fmt.Sprintf("operation=%t", operationFailure), func(t *testing.T) {
			api := &privateConnectionAPIStub{}
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: privateConnectionTestFactories(t, api),
				CheckDestroy:             api.checkDestroyed,
				Steps: []resource.TestStep{
					{Config: privateConnectionTestConfig("connection-initial", privateConnectionTestPeering)},
					{
						PreConfig: func() {
							api.mu.Lock()
							defer api.mu.Unlock()
							api.updateError = status.Error(codes.FailedPrecondition, "update rejected")
							api.operationFailure = operationFailure
						},
						Config:      privateConnectionTestConfig("connection-updated", privateConnectionTestPeering),
						ExpectError: regexp.MustCompile("(?s)(Failed to Update resource|Unable to Update Resource).*update rejected"),
					},
				},
			})
		})
	}
}

type privateConnectionTestProvider struct {
	config *providerconfig.Config
}

func (p *privateConnectionTestProvider) Metadata(_ context.Context, _ frameworkprovider.MetadataRequest, resp *frameworkprovider.MetadataResponse) {
	resp.TypeName = "yandex"
}

func (p *privateConnectionTestProvider) Schema(_ context.Context, _ frameworkprovider.SchemaRequest, resp *frameworkprovider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}

func (p *privateConnectionTestProvider) Configure(_ context.Context, _ frameworkprovider.ConfigureRequest, resp *frameworkprovider.ConfigureResponse) {
	resp.ResourceData = p.config
	resp.DataSourceData = p.config
}

func (p *privateConnectionTestProvider) Resources(_ context.Context) []func() frameworkresource.Resource {
	return []func() frameworkresource.Resource{yandex_cic_private_connection.NewResourceWithMaskedLogging}
}

func (p *privateConnectionTestProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		yandex_cic_private_connection.NewDataSourceWithMaskedLogging,
		yandex_cic_trunk_connection.NewDataSource,
		yandex_cic_public_connection.NewDataSourceWithMaskedLogging,
		yandex_cic_partner.NewDataSource,
		yandex_cic_point_of_presence.NewDataSource,
	}
}

func privateConnectionTestFactories(t *testing.T, api *privateConnectionAPIStub) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()
	sdk, err := ycsdk.Build(context.Background(),
		options.WithCredentials(credentials.NoAuthentication()),
		options.WithEndpointsResolver(endpoints.NewSingleEndpointResolver("unused:443")),
		options.WithPlaintext(),
		options.WithCustomDialOptions(grpc.WithUnaryInterceptor(api.invoke)),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			t.Error(err)
		}
	})
	return map[string]func() (tfprotov6.ProviderServer, error){
		"yandex": providerserver.NewProtocol6WithError(&privateConnectionTestProvider{
			config: &providerconfig.Config{
				SDKv2:         sdk,
				ProviderState: providerconfig.State{FolderID: types.StringValue("folder-id")},
			},
		}),
	}
}

type privateConnectionAPIStub struct {
	mu                             sync.Mutex
	connection                     *cic.PrivateConnection
	publicConnection               *cic.PublicConnection
	returnSecrets                  bool
	pending                        *operation.Operation
	lastUpdate                     *cic.UpdatePrivateConnectionRequest
	creates, updates, moves, polls int
	updateError                    error
	moveError                      error
	waitForUpdateTimeout           bool
	operationFailure               bool
}

func (s *privateConnectionAPIStub) invoke(ctx context.Context, method string, request, reply any, _ *grpc.ClientConn, _ grpc.UnaryInvoker, _ ...grpc.CallOption) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var response proto.Message
	switch method {
	case cic.PrivateConnectionService_Create_FullMethodName:
		req := request.(*cic.CreatePrivateConnectionRequest)
		s.creates++
		s.connection = &cic.PrivateConnection{
			Id: "private-connection-1", Name: req.Name, Description: req.Description,
			FolderId: req.FolderId, TrunkConnectionId: req.TrunkConnectionId, VlanId: req.VlanId,
			Labels: req.Labels, Ipv4StaticRoutes: req.Ipv4StaticRoutes,
			Status:    cic.PrivateConnection_ACTIVE,
			CreatedAt: timestamppb.New(time.Date(2026, 10, 6, 0, 0, 0, 0, time.UTC)),
		}
		if req.Ipv4Peering != nil {
			s.connection.Ipv4Peering = &cic.Peering{
				PeeringSubnet: req.Ipv4Peering.PeeringSubnet, CloudIp: req.Ipv4Peering.CloudIp,
				PeerIp: req.Ipv4Peering.PeerIp, PeerBgpAsn: req.Ipv4Peering.PeerBgpAsn,
				PeerBgpMd5Key: req.Ipv4Peering.PeerBgpMd5Key, CloudBgpAsn: 200350,
			}
		}
		response = s.operation(&cic.CreatePrivateConnectionMetadata{PrivateConnectionId: s.connection.Id}, s.connection)
	case cic.PrivateConnectionService_Get_FullMethodName:
		if s.connection == nil || request.(*cic.GetPrivateConnectionRequest).PrivateConnectionId != s.connection.Id {
			return status.Error(codes.NotFound, "private connection not found")
		}
		connection := proto.Clone(s.connection).(*cic.PrivateConnection)
		if connection.Ipv4Peering != nil && !s.returnSecrets {
			connection.Ipv4Peering.PeerBgpMd5Key = "***"
		}
		response = connection
	case cic.PrivateConnectionService_Update_FullMethodName:
		if s.updateError != nil && !s.operationFailure {
			return s.updateError
		}
		req := request.(*cic.UpdatePrivateConnectionRequest)
		s.lastUpdate = proto.Clone(req).(*cic.UpdatePrivateConnectionRequest)
		if s.updateError == nil {
			for _, field := range req.GetUpdateMask().GetPaths() {
				switch field {
				case "name":
					s.connection.Name = req.Name
				case "description":
					s.connection.Description = req.Description
				case "labels":
					s.connection.Labels = req.Labels
				case "ipv4_static_routes":
					s.connection.Ipv4StaticRoutes = req.Ipv4StaticRoutes
				case "trunk_connection_id":
					s.connection.TrunkConnectionId = req.TrunkConnectionId
				case "vlan_id":
					s.connection.VlanId = req.VlanId
				case "ipv4_peering.peer_bgp_md5_key":
					s.connection.Ipv4Peering.PeerBgpMd5Key = req.Ipv4Peering.PeerBgpMd5Key
				case "ipv4_peering.peer_bgp_asn":
					s.connection.Ipv4Peering.PeerBgpAsn = req.Ipv4Peering.PeerBgpAsn
				default:
					return status.Errorf(codes.InvalidArgument, "unexpected update path %s", field)
				}
			}
			s.updates++
		}
		response = s.operation(&cic.UpdatePrivateConnectionMetadata{PrivateConnectionId: s.connection.Id}, s.connection)
		if s.updateError != nil {
			s.pending.Result = &operation.Operation_Error{Error: status.Convert(s.updateError).Proto()}
		}
	case cic.PrivateConnectionService_Move_FullMethodName:
		req := request.(*cic.MovePrivateConnectionRequest)
		if s.moveError != nil && !s.operationFailure {
			return s.moveError
		}
		if s.moveError == nil {
			s.connection.FolderId = req.DestinationFolderId
			s.moves++
		}
		response = s.operation(&cic.MovePrivateConnectionMetadata{PrivateConnectionId: s.connection.Id}, s.connection)
		if s.moveError != nil {
			s.pending.Result = &operation.Operation_Error{Error: status.Convert(s.moveError).Proto()}
		}
	case cic.PrivateConnectionService_Delete_FullMethodName:
		req := request.(*cic.DeletePrivateConnectionRequest)
		s.connection = nil
		response = s.operation(&cic.DeletePrivateConnectionMetadata{PrivateConnectionId: req.PrivateConnectionId}, &emptypb.Empty{})
	case "/yandex.cloud.operation.OperationService/Get":
		if s.waitForUpdateTimeout && s.lastUpdate != nil && s.connection != nil {
			<-ctx.Done()
			return ctx.Err()
		}
		s.polls++
		response = s.pending
	case cic.PartnerService_Get_FullMethodName:
		response = &cic.Partner{Id: request.(*cic.GetPartnerRequest).PartnerId, Name: "partner"}
	case cic.PointOfPresenceService_Get_FullMethodName:
		response = &cic.PointOfPresence{Id: request.(*cic.GetPointOfPresenceRequest).PointOfPresenceId, Name: "point"}
	case cic.TrunkConnectionService_Get_FullMethodName:
		response = &cic.TrunkConnection{Id: request.(*cic.GetTrunkConnectionRequest).TrunkConnectionId, Name: "trunk"}
	case cic.PublicConnectionService_Get_FullMethodName:
		response = &cic.PublicConnection{Id: request.(*cic.GetPublicConnectionRequest).PublicConnectionId, Name: "public"}
		if s.publicConnection != nil {
			response = s.publicConnection
		}
	default:
		return status.Errorf(codes.Unimplemented, "unexpected RPC %s", method)
	}
	proto.Merge(reply.(proto.Message), response)
	return nil
}

func (s *privateConnectionAPIStub) operation(metadata, response proto.Message) *operation.Operation {
	metadataAny, _ := anypb.New(metadata)
	responseAny, _ := anypb.New(response)
	s.pending = &operation.Operation{
		Id: "operation-id", Done: true, Metadata: metadataAny,
		Result: &operation.Operation_Response{Response: responseAny},
	}
	return &operation.Operation{Id: s.pending.Id, Metadata: metadataAny}
}

func (s *privateConnectionAPIStub) checkUpdateMask(paths ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		got := slices.Clone(s.lastUpdate.GetUpdateMask().GetPaths())
		slices.Sort(got)
		slices.Sort(paths)
		if !reflect.DeepEqual(got, paths) {
			return fmt.Errorf("update mask = %v, want %v", got, paths)
		}
		return nil
	}
}

func (s *privateConnectionAPIStub) checkMutations(creates, updates, moves int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.creates != creates || s.updates != updates || s.moves != moves {
			return fmt.Errorf("mutations = (%d, %d, %d), want (%d, %d, %d)", s.creates, s.updates, s.moves, creates, updates, moves)
		}
		if s.polls < creates+updates+moves {
			return fmt.Errorf("operations were not polled: %d", s.polls)
		}
		return nil
	}
}

func (s *privateConnectionAPIStub) checkDestroyed(_ *terraform.State) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.connection != nil {
		return fmt.Errorf("private connection %s still exists", s.connection.Id)
	}
	return nil
}

func privateConnectionTestConfig(name, attributes string) string {
	return fmt.Sprintf(`
resource "yandex_cic_private_connection" "test" {
  name = %q
  trunk_connection_id = "trunk-id"
  %s
}
`, name, attributes)
}

const privateConnectionTestPeering = `
  vlan_id = 101
  ipv4_peering = {
    peering_subnet = "192.168.0.0/30"
    cloud_ip = "192.168.0.1"
    peer_ip = "192.168.0.2"
    peer_bgp_asn = 65001
    peer_bgp_md5_key = "test-secret"
  }
`
