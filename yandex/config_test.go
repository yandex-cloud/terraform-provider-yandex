package yandex

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"

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
	go func() { _ = grpcServer.Serve(l) }()
	defer grpcServer.Stop()

	// instance of sdk
	config := Config{
		Endpoint:  l.Addr().String(),
		FolderID:  testConfigFolder,
		CloudID:   testConfigCloudID,
		Zone:      testConfigZone,
		Token:     testConfigToken,
		Insecure:  true,
		Plaintext: true,
	}

	err := config.initAndValidate(context.Background(), testTerraformVersion, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}

	// make a request to the mock server
	_, _ = config.sdk.ApiEndpoint().ApiEndpoint().List(context.Background(), &endpoint.ListApiEndpointsRequest{})

	// check user-agent value
	assert.Contains(t, mockServerImpl.userAgent, "terraform.io")
	assert.Contains(t, mockServerImpl.userAgent, "Terraform/")
}

type userAgentMockServerAPIEndpoint struct {
	userAgent string
	addr      string
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
	reqMd, _ := metadata.FromIncomingContext(ctx)
	userAgent := reqMd.Get("user-agent")
	if len(userAgent) > 0 {
		s.userAgent = userAgent[0]
	}

	return &endpoint.ListApiEndpointsResponse{
		Endpoints: []*endpoint.ApiEndpoint{
			{
				Id:      "endpoint",
				Address: s.addr,
			},
		},
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

func Test_iamKeyFromJSONContent(t *testing.T) {
	content, err := ioutil.ReadFile(fakeSAKeyFile)
	require.NoError(t, err, "fail on file read %s", fakeSAKeyFile)
	_, err = iamKeyFromJSONContent(string(content))
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
