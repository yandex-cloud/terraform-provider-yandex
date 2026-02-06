package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type FrontMatter struct {
	Subcategory string `yaml:"subcategory"`
	PageTitle   string `yaml:"page_title"`
	Description string `yaml:"description"`
}

type TocItem struct {
	Name  string    `yaml:"name"`
	Href  string    `yaml:"href,omitempty"`
	Items []TocItem `yaml:"items,omitempty"`
	Title string    `yaml:"title"`
}

func main() {
	flag.Parse()
	docsDir := flag.Arg(0)
	if docsDir == "" {
		log.Fatalln("Docs directory is not set, please provider docs directory path")
		return
	}

	toc := TocItem{
		Title: "Yandex Cloud Provider",
		Href:  "index.md",
	}

	err := processDirectory(filepath.Join(docsDir, "data-sources"), "Data Sources", &toc)

	if err != nil {
		log.Fatalf("Error while processing data-sources dir: %s\n", err)
		return
	}
	err = processDirectory(filepath.Join(docsDir, "resources"), "Resources", &toc)

	if err != nil {
		log.Fatalf("Error while processing resource dir: %s\n", err)
		return
	}

	sortTocItems(&toc.Items)

	tocFile, err := os.Create(filepath.Join(docsDir, "toc.yaml"))
	if err != nil {
		log.Fatalf("Error while create toc.yaml: %s\n", err)
		return
	}
	defer tocFile.Close()

	encoder := yaml.NewEncoder(tocFile)
	err = encoder.Encode(&toc)
	if err != nil {
		log.Fatalf("Error while encoding toc.yaml: %s\n", err)
	}
}

func processDirectory(dir string, itemType string, toc *TocItem) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed while travising directory %s: %w", path, err)
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		fm, err := parseFrontMatter(path)

		if err != nil {
			return fmt.Errorf("failed to parse frnt matter %s: %w", path, err)
		}

		var subcategoryItem *TocItem
		for i := range toc.Items {
			if toc.Items[i].Name == fm.Subcategory {
				subcategoryItem = &toc.Items[i]
				break
			}
		}

		if subcategoryItem == nil {
			subcategoryItem = &TocItem{
				Name: fm.Subcategory,
			}
			toc.Items = append(toc.Items, *subcategoryItem)
			subcategoryItem = &toc.Items[len(toc.Items)-1]
		}

		var itemTypeItem *TocItem
		for i := range subcategoryItem.Items {
			if subcategoryItem.Items[i].Name == itemType {
				itemTypeItem = &subcategoryItem.Items[i]
				break
			}
		}

		if itemTypeItem == nil {
			itemTypeItem = &TocItem{
				Name: itemType,
			}
			subcategoryItem.Items = append(subcategoryItem.Items, *itemTypeItem)
			itemTypeItem = &subcategoryItem.Items[len(subcategoryItem.Items)-1]
		}

		itemTypeItem.Items = append(itemTypeItem.Items, TocItem{
			Name: strings.TrimSuffix(info.Name(), ".md"),
			Href: filepath.ToSlash(strings.TrimPrefix(path, "docs/")),
		})
		return nil
	})
}

func parseFrontMatter(path string) (*FrontMatter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontMatter := false
	var frontMatterLines []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if !inFrontMatter {
				inFrontMatter = true
			} else {
				break
			}
		} else if inFrontMatter {
			frontMatterLines = append(frontMatterLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file %s: %w", path, err)
	}

	fm := FrontMatter{}
	err = yaml.Unmarshal([]byte(strings.Join(frontMatterLines, "\n")), &fm)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml file header %s: %w", path, err)
	}

	return &fm, nil
}

func sortTocItems(items *[]TocItem) {
	sort.SliceStable(*items, func(i, j int) bool {
		return (*items)[i].Name < (*items)[j].Name
	})
	for i := range *items {
		if len((*items)[i].Items) > 0 {
			sortTocItems(&((*items)[i].Items))
		}
	}
}
