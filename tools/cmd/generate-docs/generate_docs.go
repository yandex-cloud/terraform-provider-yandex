package main

import (
	"flag"
	"fmt"
	"github.com/yandex-cloud/terraform-provider-yandex/tools/cmd/pkg/categories"
	"gopkg.in/yaml.v3"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const description = "A string that can be" +
	" [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as \"30s\" or \"2h45m\"." +
	" Valid time units are \"s\" (seconds), \"m\" (minutes), \"h\" (hours)."

var newDescriptions = map[string]string{
	"create":  description,
	"delete":  description + " Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.",
	"update":  description,
	"read":    description + " Read operations occur during any refresh or planning operation when refresh is enabled.",
	"default": description,
}

const (
	defaultTemplatesDir = "templates"
	defaultDocsDir      = "docs"
)

// Header представляет структуру YAML-заголовка
type Header struct {
	Subcategory string `yaml:"subcategory"`
}

func extractSubcategory(input []byte) (string, error) {
	content := string(input)

	parts := strings.Split(content, "---")
	if len(parts) < 2 {
		return "", fmt.Errorf("не найден YAML-заголовок")
	}

	var header Header
	err := yaml.Unmarshal([]byte(parts[1]), &header)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга YAML: %v", err)
	}

	return header.Subcategory, nil
}

func postProcessingDocs(resourcePath string) error {
	log.Printf("Post processing resource %s", resourcePath)

	err := filepath.Walk(resourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed while travising directory %s: %w", path, err)
		}

		// Process only Markdown files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			return replaceTimeoutBlock(path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to post hande docs %w", err)
	}
	return nil
}

func replaceTimeoutBlock(filePath string) error {
	log.Printf("Replacing timeout block in doc %s\n", filePath)
	input, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	content := string(input)

	reTimeoutsBlock := regexp.MustCompile(`(?s)\<a id=\"nestedblock--timeouts\"\>\<\/a\>
\s*### Nested Schema for \x60timeouts\x60
\s+Optional:\n\n((- \x60(create|default|delete|read|update))\x60 \(String\)(\n)?)+`)

	matches := reTimeoutsBlock.FindStringSubmatch(content)
	if len(matches) > 1 {
		blockContent := matches[0]

		for key, desc := range newDescriptions {
			reField := regexp.MustCompile(fmt.Sprintf(`-\s*\x60` + key + `\x60\s*\(String\)(.*?)`))
			blockContent = reField.ReplaceAllString(blockContent, fmt.Sprintf("- `%s` (String) %s", key, desc))
		}

		content = strings.Replace(content, matches[0], blockContent, 1)
	}

	// Write the updated content back to the file
	err = os.WriteFile(filePath, []byte(content), os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	log.Printf("Post processed file: %s\n", filePath)
	return nil
}

func regroupTemplatesFiles(templatesDir, tmpDir string, categories categories.Categories) error {
	log.Println("Reordering templates directory for sdk provider")
	return filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed while travising directory %s: %w", path, err)
		}

		if d.IsDir() {
			return nil
		}

		filename := d.Name()
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}
		// We are checking that cats are registered
		if strings.HasPrefix(filename, "d_") || strings.HasPrefix(filename, "r_") {
			cat, err := extractSubcategory(data)
			if err != nil {
				log.Printf("Failed to extract subcategory for %s", path)
				return nil
			}
			_, ok := categories.LookUpMap[cat]
			if !ok {
				log.Printf("Category is not registed %s", cat)
				return nil
			}
		}

		var file string
		if strings.HasPrefix(filename, "d_") {
			file = filepath.Join(tmpDir, "data-sources", filename[2:])
		} else if strings.HasPrefix(filename, "r_") {
			file = filepath.Join(tmpDir, "resources", filename[2:])
		} else if filename == "index.md" {
			file = filepath.Join(tmpDir, filename)
			err = os.WriteFile(file, data, os.FileMode(0644))
			return err
		} else {
			return nil
		}
		err = os.WriteFile(file, data, os.FileMode(0644))
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", filename, err)
		}
		return nil
	})
}

func main() {
	flag.Parse()
	templatesDir := flag.Arg(0)
	if templatesDir == "" {
		log.Println("Template directory is not set, using default")
		templatesDir = defaultTemplatesDir
	}
	docsDir := flag.Arg(1)
	if docsDir == "" {
		log.Println("Docs directory is not set, using default")
		docsDir = defaultDocsDir
	}
	tmpDir, err := os.MkdirTemp(".", "templates-")

	if err != nil {
		log.Fatalln("Error creating temporary dir")
		return
	}

	datasourceDir := filepath.Join(tmpDir, "data-sources")
	if err := os.MkdirAll(datasourceDir, os.ModePerm); err != nil {
		log.Fatalln("Unable to create temporary dir data-sources")
		return
	}
	resourceDir := filepath.Join(tmpDir, "resources")
	if err := os.MkdirAll(resourceDir, os.ModePerm); err != nil {
		log.Fatalln("Unable to create temporary dir data-sources")
		return
	}

	defer os.RemoveAll(tmpDir)

	var categories_ categories.Categories
	err = categories_.LoadCategoriesMapping(filepath.Join(templatesDir, "categories.yaml"))

	if err != nil {
		log.Fatalf("Error loading category.yaml: %s", err)
		return
	}

	if err := regroupTemplatesFiles(templatesDir, tmpDir, categories_); err != nil {
		log.Fatalf("Error regrouping templates files: %v", err)
		return
	}

	log.Println("Running tfplugindocs")
	cmd := exec.Command("go", "run", "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs", "generate",
		"--provider-name", "yandex",
		"--website-source-dir", tmpDir,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running tfplugindocs: %s\n", err)
		return
	}
	err = postProcessingDocs(filepath.Join(docsDir, "resources"))
	if err != nil {
		log.Fatalf("Error post proccessing docs: %s\n", err)
		return
	}

}
