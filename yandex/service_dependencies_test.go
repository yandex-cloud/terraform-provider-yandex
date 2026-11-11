package yandex

import (
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

// This guards the import direction after moving code out of the provider
// builder. Existing acceptance-test bridges are intentionally outside its scope.
func TestServiceDependencyDirection(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the source tree")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), ".."))
	const module = "github.com/yandex-cloud/terraform-provider-yandex"
	const services = "yandex/services"
	ownerOf := func(path string) string {
		if relative, ok := strings.CutPrefix(path, services+"/"); ok {
			parts := strings.Split(relative, "/")
			for i, part := range parts {
				switch part {
				case "pluginsdk", "framework", "internal", "acceptance", "sweep", "testfixtures":
					return strings.Join(parts[:i], "/")
				}
			}
		}
		return ""
	}
	for path, want := range map[string]string{
		"yandex/services/mdb/postgresql/pluginsdk/resource.go":    "mdb/postgresql",
		"yandex/services/mdb/mysql/framework/cluster":             "mdb/mysql",
		"yandex/services/serverless/functions/acceptance/test.go": "serverless/functions",
		"yandex/services/cic/internal/logging":                    "cic",
		"internal/provider/framework/config":                      "",
	} {
		if got := ownerOf(path); got != want {
			t.Errorf("ownerOf(%q) = %q; want %q", path, got, want)
		}
	}
	for _, scope := range []struct {
		path string
	}{
		{services},
		{"yandex/internal/sdkcommon"},
		{"yandex/providerconfig"},
		{"internal/provider/framework/config"},
	} {
		t.Run(scope.path, func(t *testing.T) {
			start := filepath.Join(root, filepath.FromSlash(scope.path))
			err := filepath.WalkDir(start, func(path string, entry fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if entry.IsDir() {
					if entry.Name() == "sweep" || entry.Name() == "testfixtures" {
						return filepath.SkipDir
					}
					return nil
				}
				if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
					return nil
				}
				source, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
				if err != nil {
					return err
				}
				relative, err := filepath.Rel(root, path)
				if err != nil {
					return err
				}
				owner := ownerOf(filepath.ToSlash(relative))
				for _, imported := range source.Imports {
					name, err := strconv.Unquote(imported.Path.Value)
					if err != nil {
						return err
					}
					dependency, local := strings.CutPrefix(name, module+"/")
					if !local {
						continue
					}
					if dependency == "yandex" || dependency == "internal/provider/framework" ||
						dependency == "pkg/testhelpers" || strings.HasPrefix(dependency, "pkg/testhelpers/") ||
						strings.HasPrefix(dependency, "internal/acceptance") || strings.Contains(dependency, "/sweep") || strings.Contains(dependency, "/testfixtures") {
						t.Errorf("%s imports %s", relative, dependency)
					}
					if importedOwner := ownerOf(dependency); importedOwner != "" && importedOwner != owner {
						t.Errorf("%s imports service %s via %s", relative, importedOwner, dependency)
					}
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
