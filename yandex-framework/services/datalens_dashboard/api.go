package datalens_dashboard

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
)

type dashboardClient struct {
	client *datalens.Client
}

func (c *dashboardClient) CreateDashboard(ctx context.Context, orgID string, body map[string]any) (map[string]any, error) {
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/createDashboard", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("create dashboard: %w", err)
	}
	return resp, nil
}

func (c *dashboardClient) GetDashboard(ctx context.Context, orgID, dashboardID string) (map[string]any, error) {
	body := map[string]any{"dashboardId": dashboardID}
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/getDashboard", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("get dashboard: %w", err)
	}
	return resp, nil
}

func (c *dashboardClient) UpdateDashboard(ctx context.Context, orgID string, body map[string]any) error {
	if err := c.client.Do(ctx, "/rpc/updateDashboard", orgID, body, nil); err != nil {
		return fmt.Errorf("update dashboard: %w", err)
	}
	return nil
}

func (c *dashboardClient) DeleteDashboard(ctx context.Context, orgID, dashboardID string) error {
	body := map[string]any{"dashboardId": dashboardID}
	if err := c.client.Do(ctx, "/rpc/deleteDashboard", orgID, body, nil); err != nil {
		return fmt.Errorf("delete dashboard: %w", err)
	}
	return nil
}
