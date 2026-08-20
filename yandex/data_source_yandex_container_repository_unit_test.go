package yandex

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	containerregistry "github.com/yandex-cloud/go-genproto/yandex/cloud/containerregistry/v1"
	"google.golang.org/grpc"
)

type repositoryByNameClientStub struct {
	request *containerregistry.GetRepositoryByNameRequest
	result  *containerregistry.Repository
	err     error
}

func (c *repositoryByNameClientStub) GetByName(
	_ context.Context,
	req *containerregistry.GetRepositoryByNameRequest,
	_ ...grpc.CallOption,
) (*containerregistry.Repository, error) {
	c.request = req
	return c.result, c.err
}

func TestGetContainerRepositoryByNameUsesDedicatedRPC(t *testing.T) {
	client := &repositoryByNameClientStub{
		result: &containerregistry.Repository{Id: "repository-id", Name: "registry-id/image"},
	}

	result, err := getContainerRepositoryIDByName(context.Background(), client, "registry-id/image")

	require.NoError(t, err)
	require.Equal(t, "registry-id/image", client.request.GetRepositoryName())
	require.Equal(t, "repository-id", result)
}
