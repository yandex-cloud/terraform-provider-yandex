package categories

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

type CategoryMapping struct {
	Mapping map[string]string `yaml:"CategoryMapping"`
}

func (c *CategoryMapping) LoadCategoriesMapping(categoriesFile string) error {
	data, err := os.ReadFile(categoriesFile)
	if err != nil {
		return fmt.Errorf("failed to load categories.yaml: %w", err)
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal categories.yaml: %w", err)
	}
	return nil

}

func (c *CategoryMapping) GetSubCategoryByPath(path string) (string, error) {
	for key, value := range c.Mapping {
		if strings.Contains(path, "/"+key+"/") {
			return value, nil
		}
	}
	return "", fmt.Errorf("failed to find subcategory for path: %s", path)
}

func (c *CategoryMapping) GetCategoryFromResource(resourceName string) (string, error) {
	for key, _ := range c.Mapping {
		if strings.HasPrefix(resourceName, strings.ReplaceAll(key, "/", "_")+"_") || key == resourceName {
			return key, nil
		}
	}
	return "", fmt.Errorf("failed to find category for resource: %s", resourceName)
}
