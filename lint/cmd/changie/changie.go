package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"github.com/miniscruff/changie/cmd"
	"gopkg.in/yaml.v3"
)

var (
	bodyRe = regexp.MustCompile("^[a-zA-Z_0-9]+:\\s.+$")
)

type unreleased struct {
	Kind string `yaml:"kind"`
	Body string `yaml:"body"`
	Time string `yaml:"time"`
}

func main() {
	if err := isChangieAbleToGenerateChangelog(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "changie error duriong dry run: %s \n", err)
		os.Exit(1)
	}

	if err := checkUnreleasedBody(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "invalid body of changie unreleased: %s \n", err)
		os.Exit(1)
	}
}

func isChangieAbleToGenerateChangelog() (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = errors.New("handled panic during changie batch command execution. Reason - invalid unreleased note format")
		}
	}()

	err = cmd.RootCmd().Execute()

	return err
}

func checkUnreleasedBody() error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cur dir: %w", err)
	}

	var (
		pathToUnreleased = path.Join(root, ".changes", "unreleased")
		raw              = bytes.NewBuffer(make([]byte, 0, 1024))
		changieFileInfo  unreleased
	)
	if err := filepath.Walk(pathToUnreleased, func(path string, fsInfo fs.FileInfo, err error) error {
		if fsInfo.IsDir() || filepath.Ext(path) != ".yaml" {
			return nil
		}

		dsc, err := os.OpenFile(path, os.O_RDONLY, 0600)
		if err != nil {
			return fmt.Errorf("open changie changieFileInfo file (%s) : %w", path, err)
		}

		defer dsc.Close()
		defer raw.Reset()

		if _, err := io.Copy(raw, dsc); err != nil {
			return fmt.Errorf("read changie changieFileInfo file (%s) : %w", path, err)
		}

		if err := handleYamlFile(raw.Bytes(), &changieFileInfo); err != nil {
			return fmt.Errorf("analyse body of unreleaed file(%s): %w", path, err)
		}

		return nil

	}); err != nil {
		return fmt.Errorf("check changieFileInfo info: %w", err)
	}

	return nil
}

func handleYamlFile(content []byte, changieFileInfo *unreleased) error {
	if err := yaml.Unmarshal(content, &changieFileInfo); err != nil {
		return fmt.Errorf("parse changie unreleased file: %w", err)
	}

	if !validateChangieBody(changieFileInfo.Body) {
		return fmt.Errorf("have body (%s) in changie file. Valid body format: ('Service: feature description')", changieFileInfo.Body)
	}

	return nil
}

func validateChangieBody(input string) bool {
	return bodyRe.MatchString(input)
}
