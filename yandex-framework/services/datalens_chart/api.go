package datalens_chart

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/terraform-provider-yandex/pkg/datalens"
)

type chartClient struct {
	client *datalens.Client
}

// chartRPCSuffix maps the user-facing chart type onto the DataLens RPC method
// suffix. wizard -> "Wizard", ql -> "QL".
func chartRPCSuffix(chartType string) (string, error) {
	switch chartType {
	case "wizard":
		return "Wizard", nil
	case "ql":
		return "QL", nil
	default:
		return "", fmt.Errorf("unsupported chart type: %q (expected wizard or ql)", chartType)
	}
}

func (c *chartClient) CreateChart(ctx context.Context, orgID, chartType string, body map[string]any) (map[string]any, error) {
	suffix, err := chartRPCSuffix(chartType)
	if err != nil {
		return nil, err
	}
	injectChartConstants(body, chartType)
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/create"+suffix+"Chart", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("create %s chart: %w", chartType, err)
	}
	return resp, nil
}

// injectChartConstants adds the two API-required fields the model doesn't
// carry: the constant `template = "datalens"` at the top level and the
// `data.type` discriminator (the API expects it inside `data` even though
// the same value also dispatches the RPC URL).
func injectChartConstants(body map[string]any, chartType string) {
	body["template"] = "datalens"
	if data, ok := body["data"].(map[string]any); ok {
		data["type"] = chartType
	}
}

func (c *chartClient) GetChart(ctx context.Context, orgID, chartType, chartID string) (map[string]any, error) {
	suffix, err := chartRPCSuffix(chartType)
	if err != nil {
		return nil, err
	}
	body := map[string]any{"chartId": chartID}
	var resp map[string]any
	if err := c.client.Do(ctx, "/rpc/get"+suffix+"Chart", orgID, body, &resp); err != nil {
		return nil, fmt.Errorf("get %s chart: %w", chartType, err)
	}
	return resp, nil
}

func (c *chartClient) UpdateChart(ctx context.Context, orgID, chartType string, body map[string]any) error {
	suffix, err := chartRPCSuffix(chartType)
	if err != nil {
		return err
	}
	injectChartConstants(body, chartType)
	if err := c.client.Do(ctx, "/rpc/update"+suffix+"Chart", orgID, body, nil); err != nil {
		return fmt.Errorf("update %s chart: %w", chartType, err)
	}
	return nil
}

func (c *chartClient) DeleteChart(ctx context.Context, orgID, chartType, chartID string) error {
	suffix, err := chartRPCSuffix(chartType)
	if err != nil {
		return err
	}
	body := map[string]any{"chartId": chartID}
	if err := c.client.Do(ctx, "/rpc/delete"+suffix+"Chart", orgID, body, nil); err != nil {
		return fmt.Errorf("delete %s chart: %w", chartType, err)
	}
	return nil
}
