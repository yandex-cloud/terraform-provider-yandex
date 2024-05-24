package templates

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"text/template"
)

// IsExist - search template in embedded file system
func IsExist(fileSystem fs.FS, name, tplType string) bool {
	f, err := fileSystem.Open(fmt.Sprintf("templates/%s/%s.tmpl", tplType, name))
	if err != nil {
		return false
	}
	_ = f.Close()

	return true
}

// Generate - execute template with given name and vars
func Generate(fileSystem fs.FS, tplType, name string, vars any) (io.Reader, error) {
	var (
		tpl = template.New(fmt.Sprintf("%s.tmpl", name))
	)

	parsed, err := tpl.ParseFS(fileSystem, fmt.Sprintf("templates/%s/%s.tmpl", tplType, name))
	if err != nil {
		return nil, fmt.Errorf("find template with name (%s) type: (%s): %w", name, tplType, err)
	}
	output := bytes.NewBuffer(make([]byte, 0, 2048))

	if err := parsed.Execute(output, vars); err != nil {
		return nil, fmt.Errorf("execute template with name (%s) type (%s): %w", name, tplType, err)
	}

	return output, nil
}

// Format - call go fmt for generated code
func Format(input io.Reader) (io.Reader, error) {
	rawSrc := bytes.NewBuffer(make([]byte, 0, 2048))
	if _, err := io.Copy(rawSrc, input); err != nil {
		return nil, err
	}

	formatted, err := format.Source(rawSrc.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}

	return bytes.NewBuffer(formatted), nil
}
