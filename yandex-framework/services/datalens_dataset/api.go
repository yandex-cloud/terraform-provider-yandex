package datalens_dataset

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
)

type datasetClient struct {
	client *datalens.Client
}

func (c *datasetClient) CreateDataset(ctx context.Context, orgID string, body map[string]any) (map[string]any, error) {
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/createDataset", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("create dataset: %w", err)
	}
	return resp, nil
}

func (c *datasetClient) GetDataset(ctx context.Context, orgID, datasetID string) (map[string]any, error) {
	body := map[string]any{"datasetId": datasetID}
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/getDataset", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("get dataset: %w", err)
	}
	return resp, nil
}

func (c *datasetClient) UpdateDataset(ctx context.Context, orgID, datasetID string, dataset map[string]any) error {
	body := map[string]any{"datasetId": datasetID, "dataset": dataset}
	if err := c.client.Do(ctx, "/rpc/updateDataset", orgID, body, nil); err != nil {
		return fmt.Errorf("update dataset: %w", err)
	}
	return nil
}

func (c *datasetClient) DeleteDataset(ctx context.Context, orgID, datasetID string) error {
	body := map[string]any{"datasetId": datasetID}
	if err := c.client.Do(ctx, "/rpc/deleteDataset", orgID, body, nil); err != nil {
		return fmt.Errorf("delete dataset: %w", err)
	}
	return nil
}
