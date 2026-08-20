package yandex

import (
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	iampb "github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdkv2 "github.com/yandex-cloud/go-sdk/v2"
	endpointsdk "github.com/yandex-cloud/go-sdk/v2/services/endpoint"

	"github.com/yandex-cloud/terraform-provider-yandex/common"
)

const testConfigToken = "some_special_secured_token"
const testConfigEndpoint = "endpoint.secure.me"
const testConfigCloudID = "test-cloud-id"
const testConfigFolder = "test-folder-id"
const testConfigZone = "ru-central1-a"
const testTerraformVersion = "test-terraform"

const fakeSAKeyFile = "test-fixtures/fake_service_account_key.json"

func TestConfigInitAndValidate(t *testing.T) {
	t.Parallel()

	config := Config{
		Endpoint:  testConfigEndpoint,
		FolderID:  testConfigFolder,
		CloudID:   testConfigCloudID,
		Zone:      testConfigZone,
		Token:     testConfigToken,
		Plaintext: false,
		Insecure:  false,
	}

	err := config.initAndValidate(context.Background(), testTerraformVersion, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestConfigSDKV2ConnectionOptions(t *testing.T) {
	t.Parallel()

	for _, maxRetries := range []int{0, 1, 5} {
		t.Run(fmt.Sprintf("max_retries_%d", maxRetries), func(t *testing.T) {
			config := Config{
				Endpoint:   testConfigEndpoint,
				Token:      testConfigToken,
				MaxRetries: maxRetries,
			}
			require.NoError(t, config.initAndValidate(context.Background(), testTerraformVersion, false))
		})
	}
}

func TestConfigIAMTokenDoesNotRequireEndpointDiscovery(t *testing.T) {
	t.Parallel()

	config := Config{
		Endpoint: "unreachable.invalid:443",
		Token:    "t1.a.b",
	}
	require.NoError(t, config.initAndValidate(context.Background(), testTerraformVersion, false))
	token, err := config.SDK.CreateIAMToken(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "t1.a.b", token.GetIamToken())
}

func TestConfigInitByServiceAccountKey(t *testing.T) {
	t.Parallel()

	config := Config{
		Endpoint:                       testConfigEndpoint,
		FolderID:                       testConfigFolder,
		CloudID:                        testConfigCloudID,
		Zone:                           testConfigZone,
		ServiceAccountKeyFileOrContent: fakeSAKeyFile,
		Plaintext:                      false,
		Insecure:                       false,
	}

	err := config.initAndValidate(context.Background(), testTerraformVersion, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestConfigUserAgent(t *testing.T) {
	t.Parallel()

	// make mock grpc server with ApiEndpoint service
	grpcServer := grpc.NewServer()
	mockServerImpl := &userAgentMockServerAPIEndpoint{}

	endpoint.RegisterApiEndpointServiceServer(grpcServer, mockServerImpl)

	l := localListener(t)
	mockServerImpl.endpoints = []*endpoint.ApiEndpoint{
		{Id: "endpoint", Address: l.Addr().String()},
		{Id: "iam", Address: l.Addr().String()},
	}
	go func() { _ = grpcServer.Serve(l) }()
	defer grpcServer.Stop()

	// instance of sdk
	config := Config{
		Endpoint:   l.Addr().String(),
		FolderID:   testConfigFolder,
		CloudID:    testConfigCloudID,
		Zone:       testConfigZone,
		Token:      "t1.a.b",
		Insecure:   true,
		Plaintext:  true,
		MaxRetries: 4,
	}

	err := config.initAndValidate(context.Background(), testTerraformVersion, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	_, err = config.SDK.GetEndpoint(ycsdkv2.IamTokenCreateEndpoint)
	require.NoError(t, err)
	_, err = endpointsdk.NewApiEndpointClient(config.SDK).List(context.Background(), &endpoint.ListApiEndpointsRequest{})
	require.NoError(t, err)

	// check user-agent value
	assert.Contains(t, mockServerImpl.userAgent, "terraform.io")
	assert.Contains(t, mockServerImpl.userAgent, "Terraform/")
}

func TestConfigUsesDiscoveredServiceEndpoint(t *testing.T) {
	t.Parallel()

	grpcServer := grpc.NewServer()
	mockServerImpl := &userAgentMockServerAPIEndpoint{}
	endpoint.RegisterApiEndpointServiceServer(grpcServer, mockServerImpl)

	l := localListener(t)
	mockServerImpl.endpoints = []*endpoint.ApiEndpoint{
		{Id: "endpoint", Address: l.Addr().String()},
		{Id: "iam", Address: "iam.example.test:443"},
		{Id: "compute", Address: "compute.example.test:443"},
	}
	go func() { _ = grpcServer.Serve(l) }()
	defer grpcServer.Stop()

	config := Config{
		Endpoint:  l.Addr().String(),
		Token:     "t1.a.b",
		Plaintext: true,
	}
	require.NoError(t, config.initAndValidate(context.Background(), testTerraformVersion, false))

	computeEndpoint, err := config.SDK.GetEndpoint(protoreflect.FullName("yandex.cloud.compute.v1.DiskService.Get"))
	require.NoError(t, err)
	assert.Equal(t, "compute.example.test:443", computeEndpoint.Addr)
}

func TestConfigAppliesRequestOptionsToIAMTokenExchange(t *testing.T) {
	t.Parallel()

	grpcServer := grpc.NewServer()
	endpointServer := &userAgentMockServerAPIEndpoint{}
	iamServer := &iamTokenMockServer{}
	endpoint.RegisterApiEndpointServiceServer(grpcServer, endpointServer)
	iampb.RegisterIamTokenServiceServer(grpcServer, iamServer)

	l := localListener(t)
	endpointServer.endpoints = []*endpoint.ApiEndpoint{
		{Id: "endpoint", Address: l.Addr().String()},
		{Id: "iam", Address: l.Addr().String()},
	}
	go func() { _ = grpcServer.Serve(l) }()
	defer grpcServer.Stop()

	config := Config{
		Endpoint:  l.Addr().String(),
		Token:     "oauth-token",
		Plaintext: true,
	}
	require.NoError(t, config.initAndValidate(context.Background(), testTerraformVersion, false))
	_, err := config.SDK.CreateIAMToken(context.Background())
	require.NoError(t, err)

	userAgent := iamServer.metadata.Get("user-agent")
	require.NotEmpty(t, userAgent)
	assert.Contains(t, userAgent[0], "terraform.io")
	assert.NotEmpty(t, iamServer.metadata.Get("idempotency-key"))
	assert.NotEmpty(t, iamServer.metadata.Get("x-request-id"))
	assert.NotEmpty(t, iamServer.metadata.Get("x-client-trace-id"))
}

func TestConfigPreservesMaxRetriesSemantics(t *testing.T) {
	t.Parallel()

	grpcServer := grpc.NewServer()
	endpointServer := &userAgentMockServerAPIEndpoint{listError: status.Error(codes.Unavailable, "try again")}
	endpoint.RegisterApiEndpointServiceServer(grpcServer, endpointServer)

	l := localListener(t)
	go func() { _ = grpcServer.Serve(l) }()
	defer grpcServer.Stop()

	config := Config{
		Endpoint:   l.Addr().String(),
		Token:      "t1.a.b",
		Plaintext:  true,
		MaxRetries: 3,
	}
	require.NoError(t, config.initAndValidate(context.Background(), testTerraformVersion, false))
	_, err := config.SDK.GetEndpoint(ycsdkv2.IamTokenCreateEndpoint)
	require.Error(t, err)
	assert.Equal(t, 3, endpointServer.listCalls)
}

type userAgentMockServerAPIEndpoint struct {
	userAgent string
	endpoints []*endpoint.ApiEndpoint
	listError error
	listCalls int
}

func (s *userAgentMockServerAPIEndpoint) Get(
	context.Context,
	*endpoint.GetApiEndpointRequest,
) (*endpoint.ApiEndpoint, error) {
	return &endpoint.ApiEndpoint{}, nil
}

func (s *userAgentMockServerAPIEndpoint) List(
	ctx context.Context,
	r *endpoint.ListApiEndpointsRequest,
) (*endpoint.ListApiEndpointsResponse, error) {
	s.listCalls++
	reqMd, _ := metadata.FromIncomingContext(ctx)
	userAgent := reqMd.Get("user-agent")
	if len(userAgent) > 0 {
		s.userAgent = userAgent[0]
	}

	if s.listError != nil {
		return nil, s.listError
	}

	return &endpoint.ListApiEndpointsResponse{Endpoints: s.endpoints}, nil
}

type iamTokenMockServer struct {
	iampb.UnimplementedIamTokenServiceServer
	metadata metadata.MD
}

func (s *iamTokenMockServer) Create(
	ctx context.Context,
	_ *iampb.CreateIamTokenRequest,
) (*iampb.CreateIamTokenResponse, error) {
	s.metadata, _ = metadata.FromIncomingContext(ctx)
	return &iampb.CreateIamTokenResponse{
		IamToken:  "iam-token",
		ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
	}, nil
}

func localListener(t *testing.T) net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		l, err = net.Listen("tcp6", "[::1]:0")
	}
	require.NoError(t, err, "failed to listen on any port")
	return l
}

func Test_iamKeyV2FromJSONContent(t *testing.T) {
	content, err := ioutil.ReadFile(fakeSAKeyFile)
	require.NoError(t, err, "fail on file read %s", fakeSAKeyFile)
	_, err = iamKeyV2FromJSONContent(string(content))
	require.NoError(t, err)
}

func TestConfigInitDefaultS3ClientFromSharedCredentials(t *testing.T) {
	t.Parallel()

	config := Config{
		Endpoint:              testConfigEndpoint,
		FolderID:              testConfigFolder,
		CloudID:               testConfigCloudID,
		Zone:                  testConfigZone,
		Token:                 testConfigToken,
		StorageEndpoint:       common.DefaultStorageEndpoint,
		SharedCredentialsFile: "test-fixtures/shared-credentials-file",
		Profile:               "prod-profile",
	}

	err := config.initAndValidate(context.Background(), testTerraformVersion, false)

	if err != nil {
		t.Fatalf("failed to initAndValidate config: \"%v\"", err.Error())
	}
	require.NotNilf(t, config.defaultS3Client, "expected defaultS3Client to be initialized")
	credentials, err := config.defaultS3Client.S3().Config.Credentials.Get()
	require.NoError(t, err)
	assert.Equal(t, "YCAJEv2kbbNCegBdWneshv6Fa", credentials.AccessKeyID)
	assert.Equal(t, "YCMw-QhGTK40ulcCnr1v0EsTOKZwdNv0EsTOKZwdN", credentials.SecretAccessKey)
}

func TestConfigInitDefaultS3Client_PreferAccessKeysFromConfig(t *testing.T) {
	t.Parallel()

	config := Config{
		Endpoint:              testConfigEndpoint,
		FolderID:              testConfigFolder,
		CloudID:               testConfigCloudID,
		Zone:                  testConfigZone,
		Token:                 testConfigToken,
		StorageEndpoint:       common.DefaultStorageEndpoint,
		StorageAccessKey:      "access-key",
		StorageSecretKey:      "secret-key",
		SharedCredentialsFile: "test-fixtures/shared-credentials-file",
		Profile:               "prod-profile",
	}

	err := config.initAndValidate(context.Background(), testTerraformVersion, false)

	if err != nil {
		t.Fatalf("failed to initAndValidate config: \"%v\"", err.Error())
	}
	require.NotNilf(t, config.defaultS3Client, "expected defaultS3Client to be initialized")
	credentials, err := config.defaultS3Client.S3().Config.Credentials.Get()
	require.NoError(t, err)
	assert.Equal(t, "access-key", credentials.AccessKeyID)
	assert.Equal(t, "secret-key", credentials.SecretAccessKey)
}
