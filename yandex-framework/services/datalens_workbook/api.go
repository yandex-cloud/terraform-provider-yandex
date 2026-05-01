package datalens_workbook

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
)

type workbookClient struct {
	client *datalens.Client
}

func (c *workbookClient) CreateWorkbook(ctx context.Context, orgID string, body map[string]any) (map[string]any, error) {
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/createWorkbook", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("create workbook: %w", err)
	}
	return resp, nil
}

func (c *workbookClient) GetWorkbook(ctx context.Context, orgID, workbookID string) (map[string]any, error) {
	var resp map[string]any
	body := map[string]any{"workbookId": workbookID}
	if err := c.client.Do(ctx, "/rpc/getWorkbook", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("get workbook: %w", err)
	}
	return resp, nil
}

func (c *workbookClient) UpdateWorkbook(ctx context.Context, orgID string, body map[string]any) (map[string]any, error) {
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/updateWorkbook", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("update workbook: %w", err)
	}
	return resp, nil
}

func (c *workbookClient) DeleteWorkbook(ctx context.Context, orgID, workbookID string) error {
	body := map[string]any{"workbookId": workbookID}
	if err := c.client.Do(ctx, "/rpc/deleteWorkbook", orgID, body, nil); err != nil {
		return fmt.Errorf("delete workbook: %w", err)
	}
	return nil
}
