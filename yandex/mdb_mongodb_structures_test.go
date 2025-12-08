package yandex

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strings"
	"testing"
)

func TestMdbMongodbConfigUpdateFieldsMapCompleteness(t *testing.T) {
	expandPaths := extractPathsFromExpand(t)

	var missing []string
	for path := range expandPaths {
		if _, ok := mdbMongodbConfigUpdateFieldsMap[path]; !ok {
			missing = append(missing, path)
		}
	}

	if len(missing) > 0 {
		t.Errorf("mdbMongodbConfigUpdateFieldsMap is missing paths used in Expand:\n%s",
			strings.Join(missing, "\n"))
	}
}

func TestMdbMongodbConfigUpdateFieldsMapNoExtraPaths(t *testing.T) {
	expandPaths := extractPathsFromExpand(t)

	mustBeExtra := []string{
		"cluster_config.0.mongod",
		"cluster_config.0.mongos",
		"cluster_config.0.mongocfg",
	}

	var extra []string
	for path := range mdbMongodbConfigUpdateFieldsMap {
		if _, ok := expandPaths[path]; !ok {
			extra = append(extra, path)
		}
	}

	slices.Sort(extra)
	slices.Sort(mustBeExtra)

	if !slices.Equal(extra, mustBeExtra) {
		slices.DeleteFunc(extra, func(path string) bool {
			return slices.Contains(mustBeExtra, path)
		})
		t.Errorf("mdbMongodbConfigUpdateFieldsMap contains paths not used in Expand:\n%s",
			strings.Join(extra, "\n"))
	}
}

func TestMdbMongodbConfigUpdateFieldsMapNoDuplicateValues(t *testing.T) {
	seen := make(map[string]string)
	var duplicates []string

	for tfPath, apiPath := range mdbMongodbConfigUpdateFieldsMap {
		if existingTfPath, ok := seen[apiPath]; ok {
			duplicates = append(duplicates, apiPath+" (used by "+existingTfPath+" and "+tfPath+")")
		} else {
			seen[apiPath] = tfPath
		}
	}

	if len(duplicates) > 0 {
		t.Errorf("mdbMongodbConfigUpdateFieldsMap contains duplicate API paths:\n%s",
			strings.Join(duplicates, "\n"))
	}
}

func extractPathsFromExpand(t *testing.T) map[string]struct{} {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "mdb_mongodb_structures.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("failed to parse mdb_mongodb_structures.go: %v", err)
	}

	paths := make(map[string]struct{})

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.KeyValueExpr:
			if ident, ok := node.Key.(*ast.Ident); ok && ident.Name == "Expand" {
				ast.Inspect(node.Value, func(inner ast.Node) bool {
					extractGetPaths(inner, paths)
					return true
				})
				return false
			}
		}
		return true
	})

	return paths
}

func extractGetPaths(n ast.Node, paths map[string]struct{}) {
	call, ok := n.(*ast.CallExpr)
	if !ok {
		return
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}

	methodName := sel.Sel.Name
	if methodName != "Get" && methodName != "GetOk" {
		return
	}

	if len(call.Args) == 0 {
		return
	}

	lit, ok := call.Args[0].(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return
	}

	path := strings.Trim(lit.Value, `"`)
	paths[path] = struct{}{}
}
