package filesystem

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

// GetPathForGeneratedContent - get valid output path for created content
func GetPathForGeneratedContent(pathToRepo, tplType, tplName, serviceName, resourceName string) string {
	var (
		parts    []string
		fileName string
	)
	parts = append(parts, pathToRepo, "yandex-framework", "services", serviceName, resourceName)

	if tplType == "datasource" {
		fileName = "datasource"
	} else if tplType == "resource" {
		fileName = "resource"
	}

	if tplName != "default" {
		fileName = fmt.Sprintf("%s_%s.go", fileName, tplName)
	} else {
		fileName = fmt.Sprintf("%s.go", fileName)
	}
	parts = append(parts, fileName)

	return path.Join(parts...)
}

// WriteContent - copy content from io.Reader to the file with path
func WriteContent(outputPath string, override bool, content io.Reader) error {
	if _, err := os.Stat(outputPath); !errors.Is(err, os.ErrNotExist) && !override {
		return fmt.Errorf("file with path (%s) already exists. Use force flag or delete exists file", outputPath)
	}
	dir, _ := path.Split(outputPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("prepare directory for path(%s): %w", dir, err)
	}

	f, err := os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()

	if err != nil {
		return fmt.Errorf("open file (%s) for writing: %w", outputPath, err)
	}

	if _, err := io.Copy(f, content); err != nil {
		return fmt.Errorf("write content to file (%s) : %w", outputPath, err)
	}

	return nil
}
