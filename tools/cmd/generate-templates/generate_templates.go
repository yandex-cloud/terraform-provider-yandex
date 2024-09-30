package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yandex-cloud/terraform-provider-yandex/tools/cmd/pkg/categories"
	"github.com/yandex-cloud/terraform-provider-yandex/yandex"
	yandex_framework "github.com/yandex-cloud/terraform-provider-yandex/yandex-framework/provider"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const defaultTemplate = `---
subcategory: "{{.SubCategory}}"
page_title: "Yandex: {{.Name}}"
description: |-
{{ .Description | plainmarkdown | trimspace | prefixlines "  " }}
---

# {{.Name}}

{{ .Description | trimspace }}

{{- /* Uncomment this block as you add .ExampleFile

## Example Usage

{{tffile .ExampleFile }}

*/ -}}

{{ .SchemaMarkdown | trimspace }}


{{- /* Uncomment this block as you add .ImportFile

## Import

Import is supported using the following syntax:

{{codefile "shell" .ImportFile }}

*/ -}}
`

type servicesSet struct {
	resources   map[string]struct{}
	dataSources map[string]struct{}
}

func getProviderServicesSet(ctx context.Context, sdkProvider *schema.Provider, frameworkProvider provider.Provider) (*servicesSet, error) {
	log.Println("Loading services set from provider schema")
	data := servicesSet{
		resources:   make(map[string]struct{}),
		dataSources: make(map[string]struct{}),
	}

	for name, _ := range sdkProvider.DataSourcesMap {
		data.dataSources[strings.TrimPrefix(name, "yandex_")] = struct{}{}
	}

	for name, _ := range sdkProvider.ResourcesMap {
		data.resources[strings.TrimPrefix(name, "yandex_")] = struct{}{}
	}

	for _, dataSource := range frameworkProvider.DataSources(ctx) {
		req := datasource.MetadataRequest{ProviderTypeName: "yandex"}
		resp := datasource.MetadataResponse{}
		dataSource().Metadata(context.Background(), req, &resp)
		data.dataSources[strings.TrimPrefix(resp.TypeName, "yandex_")] = struct{}{}
	}

	for _, resource_ := range frameworkProvider.Resources(ctx) {
		req := resource.MetadataRequest{ProviderTypeName: "yandex"}
		resp := resource.MetadataResponse{}
		resource_().Metadata(context.Background(), req, &resp)
		data.resources[strings.TrimPrefix(resp.TypeName, "yandex_")] = struct{}{}
	}

	return &data, nil
}

func getTemplatesServicesSet(templatesDir string) (*servicesSet, error) {
	log.Println("Loading service set from templates dir")
	data := servicesSet{
		resources:   make(map[string]struct{}),
		dataSources: make(map[string]struct{}),
	}
	// Walk through all files in the root directory
	err := filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		filename := strings.TrimSuffix(d.Name(), ".md.tmpl")

		if strings.Contains(path, "data-sources") {
			data.dataSources[filename] = struct{}{}
		} else if strings.Contains(path, "resources") {
			data.resources[filename] = struct{}{}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error while travising templates dir: %w", err)
	}

	return &data, nil
}

func findMissingTemplatesService(tempSS, providerSS *servicesSet) *servicesSet {
	log.Println("Finding missing services in docs templates")
	data := servicesSet{
		resources:   make(map[string]struct{}),
		dataSources: make(map[string]struct{}),
	}
	for service, _ := range providerSS.dataSources {
		_, ok := tempSS.dataSources[service]
		if !ok {
			log.Printf("data-source %s is not in templates", service)
			data.dataSources[service] = struct{}{}
		}
	}

	for service, _ := range providerSS.resources {
		_, ok := tempSS.resources[service]
		if !ok {
			log.Printf("resource %s is not in templates", service)
			data.resources[service] = struct{}{}
		}
	}

	return &data
}

func generateMissingTemplates(ctx context.Context, services *servicesSet, templatesDir string, mapping categories.CategoryMapping) error {
	for dataSource, _ := range services.dataSources {
		template := renderTemplate(dataSource, "data-sources")
		filename := dataSource + ".md.tmpl"
		service, err := mapping.GetCategoryFromResource(dataSource)
		if err != nil {
			return fmt.Errorf("failed to find servcie for datasoruce: %w", err)
		}
		dir := filepath.Join(templatesDir, service, "data-sources")
		err = os.MkdirAll(dir, os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("failed to create dir: %w", err)
		}
		file := filepath.Join(dir, filename)
		err = os.WriteFile(file, []byte(template), os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("failed to save template for %s in file: %w", dataSource, err)
		}

	}

	for resource_, _ := range services.resources {
		template := renderTemplate(resource_, "resources")
		filename := resource_ + ".md.tmpl"
		service, err := mapping.GetCategoryFromResource(resource_)
		if err != nil {
			return fmt.Errorf("failed to find servcie for resource: %w", err)
		}
		dir := filepath.Join(templatesDir, service, "resources")
		err = os.MkdirAll(dir, os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("failed to create dir: %w", err)
		}
		file := filepath.Join(dir, filename)
		err = os.WriteFile(file, []byte(template), os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("failed to save template for %s in file: %w", resource_, err)
		}

	}

	return nil
}

func renderTemplate(service, type_ string) string {
	template := strings.ReplaceAll(defaultTemplate, ".ExampleFile",
		fmt.Sprintf("\"examples/%s/%s/example_1.tf\"", service, type_))

	template = strings.ReplaceAll(template, ".ImportFile",
		fmt.Sprintf("\"examples/%s/%s/import/import.sh\"", service, type_))

	return template
}

func main() {
	flag.Parse()
	var templatesDir = flag.Arg(0)
	if templatesDir == "" {
		log.Fatalln("Template directory is not set, please provider template directory path")
		return
	}

	ctx := context.Background()
	provider_ := yandex.NewSDKProvider()
	frameworkProvider := yandex_framework.NewFrameworkProvider()

	providerServicesSet, err := getProviderServicesSet(ctx, provider_, frameworkProvider)
	if err != nil {
		log.Fatalf("Failed to get provider schema: %s", err)
		return
	}

	templatesServicesSet, err := getTemplatesServicesSet(templatesDir)
	if err != nil {
		log.Fatalf("Failed to get services form templates: %s", err)
		return
	}

	diff := findMissingTemplatesService(templatesServicesSet, providerServicesSet)

	var categoryMapping categories.CategoryMapping

	err = categoryMapping.LoadCategoriesMapping(filepath.Join(templatesDir, "categories.yaml"))
	if err != nil {
		return
	}

	err = generateMissingTemplates(ctx, diff, templatesDir, categoryMapping)
	if err != nil {
		log.Fatalf("Failed to generate missing template: %s", err)
		return
	}

}
