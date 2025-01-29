package categories

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Categories struct {
	Categories []string `yaml:"DocsCategories"`
	LookUpMap  map[string]string
}

func (c *Categories) LoadCategoriesMapping(categoriesFile string) error {
	data, err := os.ReadFile(categoriesFile)
	if err != nil {
		return fmt.Errorf("failed to load categories.yaml: %w", err)
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal categories.yaml: %w", err)
	}

	c.LookUpMap = make(map[string]string)
	for _, cat := range c.Categories {
		c.LookUpMap[cat] = ""
	}
	return nil

}
