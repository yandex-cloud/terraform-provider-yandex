package datalens_connection

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
)

type connectionClient struct {
	client *datalens.Client
}

func (c *connectionClient) CreateConnection(ctx context.Context, orgID string, body map[string]any) (string, error) {
	var resp struct {
		ID string `json:"id"`
	}
	if err := c.client.Do(ctx, "/rpc/createConnection", orgID, body, &resp); err != nil {
		return "", fmt.Errorf("create connection: %w", err)
	}
	return resp.ID, nil
}

func (c *connectionClient) GetConnection(ctx context.Context, orgID, connectionID string) (map[string]any, error) {
	body := map[string]any{"connectionId": connectionID}
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/getConnection", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("get connection: %w", err)
	}
	return resp, nil
}

func (c *connectionClient) UpdateConnection(ctx context.Context, orgID, connectionID string, data map[string]any) (map[string]any, error) {
	body := map[string]any{"connectionId": connectionID, "data": data}
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/updateConnection", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("update connection: %w", err)
	}
	return resp, nil
}

func (c *connectionClient) DeleteConnection(ctx context.Context, orgID, connectionID string) error {
	body := map[string]any{"connectionId": connectionID}
	if err := c.client.Do(ctx, "/rpc/deleteConnection", orgID, body, nil); err != nil {
		return fmt.Errorf("delete connection: %w", err)
	}
	return nil
}
