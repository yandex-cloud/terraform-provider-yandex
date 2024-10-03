// Package XS003 defines an Analyzer that checks for
// Schema/Resource where the Description field is configured.
package XS003

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

const doc = `Check for Schema and Resource that Description is configured.

The XS003 analyzer reports cases where schemas/resources 
are missing a Description, which is generally useful for providers that 
automatically generate documentation based on schema information.`

const analyzerName = "XS003"

// Analyzer defines the schema/resource description analyzer.
var Analyzer = &analysis.Analyzer{
	Name:             analyzerName,
	Doc:              doc,
	Run:              run,
	RunDespiteErrors: true,
}

var attributesSet = map[string]struct{}{
	"BoolAttribute":         {},
	"DynamicAttribute":      {},
	"Float32Attribute":      {},
	"Float64Attribute":      {},
	"Int32Attribute":        {},
	"Int64Attribute":        {},
	"ListAttribute":         {},
	"MapAttribute":          {},
	"NumberAttribute":       {},
	"ObjectAttribute":       {},
	"SetAttribute":          {},
	"StringAttribute":       {},
	"ListNestedAttribute":   {},
	"MapNestedAttribute":    {},
	"SetNestedAttribute":    {},
	"SingleNestedAttribute": {},
	"Schema":                {},
}

// run performs the analysis on the provided package.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			// Check if the node is a composite literal (e.g., struct literal)
			compositeLit, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}

			// Check if the type of the composite literal is schema.Something
			selExpr, ok := compositeLit.Type.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			_, exists := attributesSet[selExpr.Sel.Name]
			if !exists {
				return false
			}

			pkgIdent, ok := selExpr.X.(*ast.Ident)
			if !ok || pkgIdent.Name != "schema" {
				return false
			}

			// Check if the Description field is present
			descriptionPresent := false
			for _, element := range compositeLit.Elts {
				kvExpr, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}

				keyIdent, ok := kvExpr.Key.(*ast.Ident)
				if ok && keyIdent.Name == "Description" {
					descriptionPresent = true
					break
				}
			}

			// Report if the Description field is missing
			if !descriptionPresent {
				pass.Reportf(compositeLit.Pos(), "%s: description field should be configured", analyzerName)
			}

			return true
		})
	}

	return nil, nil
}
