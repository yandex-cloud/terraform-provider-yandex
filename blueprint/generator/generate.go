package generator

import (
	"context"
	"embed"
	"fmt"
	"io"

	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/command/generate"
	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/filesystem"
	"github.com/yandex-cloud/terraform-provider-yandex/blueprint/templates"
)

var (
	//go:embed templates/*
	fs embed.FS
)

type Opts func(generator *Generator)

func WithTemplateName(name string) Opts {
	return func(generator *Generator) {
		generator.tplName = name
	}
}

func WithTemplateType(t string) Opts {
	return func(generator *Generator) {
		generator.tplType = t
	}
}

func WithOverrideFiles(force bool) Opts {
	return func(generator *Generator) {
		generator.override = force
	}
}

func WithSkipComments(skip bool) Opts {
	return func(generator *Generator) {
		generator.skipComments = skip
	}
}

type Generator struct {
	tplType      string
	tplName      string
	resourceName string
	serviceName  string
	tplVars      any
	override     bool
	skipComments bool
}

func New(serviceName, resourceName string, opts ...Opts) *Generator {
	def := &Generator{
		tplName:      "default",
		resourceName: resourceName,
		serviceName:  serviceName,
	}

	for _, opt := range opts {
		opt(def)
	}

	return def
}

func (g *Generator) Generate(_ context.Context, output io.Writer) error {
	fmt.Fprintf(
		output,
		"Start generating %s from template: %s for service: %s entity: %s ... \n",
		g.tplType, g.tplName, g.serviceName, g.resourceName,
	)
	content, err := g.generate()
	if err != nil {
		return fmt.Errorf("generate main file: %w", err)
	}

	pathWithGeneratedFile, err := g.saveToFile(content)
	if err != nil {
		return fmt.Errorf("save main file: %w", err)
	}

	_, _ = fmt.Fprintf(output, "File sucessfully generated and placed by path: %s \n", pathWithGeneratedFile)
	return nil
}

func (g *Generator) generate() (io.Reader, error) {
	if !templates.IsExist(fs, g.tplName, g.tplType) {
		return nil, fmt.Errorf("template with provided name (%s) and type (%s) doesn't exist", g.tplName, g.tplType)
	}

	content, err := templates.Generate(
		fs,
		g.tplType,
		g.tplName,
		variablesForTemplate(g.tplType, g.tplName, g.serviceName, g.resourceName, g.skipComments),
	)
	if err != nil {
		return nil, fmt.Errorf("generate template (%s) : %w", g.tplName, err)
	}

	formattedContent, err := templates.Format(content)
	if err != nil {
		return nil, fmt.Errorf("generate template (%s) : %w", g.tplName, err)
	}

	return formattedContent, nil
}

func (g *Generator) saveToFile(content io.Reader) (string, error) {
	outputPath := filesystem.GetPathForGeneratedContent(generate.PathToRepo, g.tplType, g.tplName, g.serviceName, g.resourceName)
	if err := filesystem.WriteContent(outputPath, g.override, content); err != nil {
		return "", fmt.Errorf("write generated template to file (%s): %w", outputPath, err)
	}

	return outputPath, nil
}
