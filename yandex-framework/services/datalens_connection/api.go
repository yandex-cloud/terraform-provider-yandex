package datalens_connection

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
)

type connectionClient struct {
	client *datalens.Client
}

type createConnectionRequest map[string]interface{}

type createConnectionResponse struct {
	ID string `json:"id"`
}

func (c *connectionClient) CreateConnection(ctx context.Context, orgID string, body createConnectionRequest) (string, error) {
	var resp createConnectionResponse
	if err := c.client.Do(ctx, "/rpc/createConnection", orgID, body, &resp); err != nil {
		return "", fmt.Errorf("create connection: %w", err)
	}
	return resp.ID, nil
}

type getConnectionRequest struct {
	ConnectionID string `json:"connectionId"`
}

func (c *connectionClient) GetConnection(ctx context.Context, orgID, connectionID string) (map[string]interface{}, error) {
	req := getConnectionRequest{ConnectionID: connectionID}
	var resp map[string]interface{}
	if err := c.client.Do(ctx, "/rpc/getConnection", orgID, req, &resp); err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	return resp, nil
}

type updateConnectionRequest struct {
	ConnectionID string                 `json:"connectionId"`
	Data         map[string]interface{} `json:"data"`
}

func (c *connectionClient) UpdateConnection(ctx context.Context, orgID, connectionID string, data map[string]interface{}) (map[string]interface{}, error) {
	req := updateConnectionRequest{ConnectionID: connectionID, Data: data}
	var resp map[string]interface{}
	if err := c.client.Do(ctx, "/rpc/updateConnection", orgID, req, &resp); err != nil {
		return nil, fmt.Errorf("update connection: %w", err)
	}
	return resp, nil
}

type deleteConnectionRequest struct {
	ConnectionID string `json:"connectionId"`
}

func (c *connectionClient) DeleteConnection(ctx context.Context, orgID, connectionID string) error {
	req := deleteConnectionRequest{ConnectionID: connectionID}
	if err := c.client.Do(ctx, "/rpc/deleteConnection", orgID, req, nil); err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	return nil
}
