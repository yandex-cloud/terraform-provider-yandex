package resourceid

import (
	"fmt"
	"strings"
)

func Construct(clusterID string, resourceName string) string {
	return fmt.Sprintf("%s:%s", clusterID, resourceName)
}

func Deconstruct(resourceID string) (string, string, error) {
	parts := strings.SplitN(resourceID, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Invalid resource id format: %q", resourceID)
	}

	clusterID := parts[0]
	resourceName := parts[1]
	return clusterID, resourceName, nil
}
