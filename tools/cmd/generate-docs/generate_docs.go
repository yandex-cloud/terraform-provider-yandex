package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yandex-cloud/terraform-provider-yandex/tools/cmd/pkg/categories"
	"gopkg.in/yaml.v3"
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

// ===============================================================
// SDKv2 Resource & Data-source pairs List where empty Description
// fields should be filled for nested Attributes blocks and
// Attributes in such nested blocks.
//
// When Resource and Data-source pair has been migrated to
// Terraform plugin framework, it should be removed from this list.
// ================================================================
var SDKv2ResourcesList = []string{
	"alb_backend_group",
	"alb_http_router",
	"alb_load_balancer",
	"alb_target_group",
	"alb_virtual_host",
	"api_gateway",
	"audit_trails_trail",
	"backup_policy",
	"cdn_origin_group",
	"cdn_resource",
	"cm_certificate",
	"compute_disk",
	"compute_disk_placement_group",
	"compute_filesystem",
	"compute_gpu_cluster",
	"compute_image",
	"compute_instance",
	"compute_instance_group",
	"compute_placement_group",
	"compute_snapshot",
	"compute_snapshot_schedule",
	"container_registry",
	"container_registry_ip_permission",
	"container_repository",
	"container_repository_lifecycle_policy",
	"dataproc_cluster",
	"dns_zone",
	"function",
	"function_scaling_policy",
	"function_trigger",
	"iam_service_account",
	"iam_workload_identity_federated_credential",
	"iam_workload_identity_oidc_federation",
	"iot_core_broker",
	"iot_core_device",
	"iot_core_registry",
	"kms_asymmetric_encryption_key",
	"kms_asymmetric_signature_key",
	"kms_symmetric_key",
	"kubernetes_cluster",
	"kubernetes_node_group",
	"lb_network_load_balancer",
	"lb_target_group",
	"loadtesting_agent",
	"lockbox_secret",
	"lockbox_secret_version",
	"logging_group",
	"mdb_clickhouse_cluster",
	"mdb_greenplum_cluster",
	"mdb_greenplum_user",
	"mdb_greenplum_resource_group",
	"mdb_kafka_cluster",
	"mdb_kafka_connector",
	"mdb_kafka_topic",
	"mdb_kafka_user",
	"mdb_mongodb_cluster",
	"mdb_mysql_cluster",
	"mdb_mysql_database",
	"mdb_mysql_user",
	"mdb_postgresql_cluster",
	"mdb_postgresql_database",
	"mdb_postgresql_user",
	"mdb_redis_cluster",
	"mdb_clickhouse_database",
	"mdb_clickhouse_user",
	"message_queue",
	"monitoring_dashboard",
	"organizationmanager_group",
	"organizationmanager_os_login_settings",
	"organizationmanager_saml_federation",
	"organizationmanager_saml_federation_user_account",
	"organizationmanager_user_ssh_key",
	"resourcemanager_cloud",
	"resourcemanager_folder",
	"serverless_container",
	"serverless_eventrouter_bus",
	"serverless_eventrouter_connector",
	"serverless_eventrouter_rule",
	"smartcaptcha_captcha",
	"sws_advanced_rate_limiter_profile",
	"sws_security_profile",
	"sws_waf_profile",
	"vpc_address",
	"vpc_gateway",
	"vpc_network",
	"vpc_private_endpoint",
	"vpc_route_table",
	"vpc_security_group",
	"vpc_subnet",
	"ydb_database_dedicated",
	"ydb_database_serverless",
}

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

	// Copy Descriptions of nested blocks (attributes & blocks)
	// from SDKv2 resource generated doc to the SDKv2 data-source generated doc
	wd, _ := os.Getwd()
	for _, resName := range SDKv2ResourcesList {
		srcName := fmt.Sprintf("%s/docs/resources/%s.md", wd, resName)
		dstName := fmt.Sprintf("%s/docs/data-sources/%s.md", wd, resName)
		FixDataSourceDescriptions(dstName, GetResourceDescriptions(srcName))
	}

	err := filepath.Walk(resourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed while traversing directory %s: %w", path, err)
		}

		// Process only Markdown files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			return replaceTimeoutBlock(path)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to post process docs %w", err)
	}
	return nil
}

func replaceTimeoutBlock(filePath string) error {
	//log.Printf("Replacing timeout block in doc %s\n", filePath)
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

	//log.Printf("Post processed file: %s\n", filePath)
	return nil
}

func regroupTemplatesFiles(templatesDir, tmpDir string, categories categories.Categories) error {
	log.Println("Reordering templates directory for sdk provider")
	return filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("failed while traversing directory %s: %w", path, err)
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
				log.Printf("Category is not registered %s", cat)
				return nil
			}
		}

		var file string
		filename += ".tmpl"
		if strings.HasPrefix(filename, "d_") {
			file = filepath.Join(tmpDir, "data-sources", filename[2:])
		} else if strings.HasPrefix(filename, "r_") {
			file = filepath.Join(tmpDir, "resources", filename[2:])
		} else if filename == "index.md.tmpl" {
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

func FixDataSourceDescriptions(FileName string, Data map[string]map[string]map[string]string) {

	blockPrefix := "Nested Schema for "
	nestedPrefix := "(see [below for nested schema]"
	currentBlock := ""

	tmpSuffix := ".tmp"
	targetFile := FileName + tmpSuffix

	// Open DataSource file for Read
	file, err := os.Open(FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fileTmp, err := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer fileTmp.Close()

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		strLine := string(line)
		strNew := ""

		// Check line for block prefix
		if strings.Contains(strLine, blockPrefix) {
			currentBlock = strings.Split(strLine, "`")[1]
		}

		// Check line for attribute
		if strings.HasPrefix(strLine, "- `") {
			if len(currentBlock) > 0 {
				// Parse attribute name
				attrName := strings.Split(strLine, "`")[1]
				attrDescr := ""
				// if attribute description has nested reference to another block
				if strings.Contains(strLine, nestedPrefix) {
					attrDescr = strings.Split(strLine, nestedPrefix)[1]
				}

				// Build new line
				val, key := Data[currentBlock][attrName]
				if key {
					// regular attribute
					if len(attrDescr) == 0 {
						strNew = fmt.Sprintf("%s %s\n", strLine, val["descr"])
						// nested block with reference
					} else {
						strNew = fmt.Sprintf("- `%s` (%s) %s %s%s\n",
							attrName, val["type"], val["descr"], nestedPrefix, attrDescr)
					}
					strLine = strNew
				}
			}
		}

		// Write line to the the Temp file (.tmp)
		_, err = fileTmp.Write([]byte(strLine + "\n"))
		if err != nil {
			log.Fatal(err)
		}
	}
	// Rename temp filename to the target filename
	err = os.Rename(targetFile, strings.TrimSuffix(targetFile, tmpSuffix))
	if err != nil {
		log.Fatal(err)
	}
}

func GetResourceDescriptions(FileName string) map[string]map[string]map[string]string {

	blockPrefix := "Nested Schema for "
	currentBlock := ""

	data := make(map[string]map[string]map[string]string)

	// Open Resource file for Read
	file, err := os.Open(FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		aData := string(line)

		// Check line for block prefix
		if strings.Contains(aData, blockPrefix) {
			currentBlock = strings.Split(aData, "`")[1]
			continue
			// Check line for attribute prefix
		} else if strings.HasPrefix(aData, "- `") {
			if len(currentBlock) > 0 {
				aName, aType, aDescr, aLink := "", "", "", ""
				// Parse the attribute Name and clean it
				aVal := strings.Split(aData, "` (")
				aName = strings.TrimPrefix(aVal[0], "- `")
				// Parse the attribute Type
				aType = strings.Split(aVal[1], ")")[0]

				// if link to nested block was found. Save it separated from description
				aBuf := strings.Split(aVal[1], "(see [below for nested schema](#nestedblock--")
				if len(aBuf) > 1 {
					aDescr = strings.TrimSpace(strings.Split(aBuf[0], aType+")")[1])
					aLink = strings.TrimSuffix(aBuf[1], "))")
					// If it regular attribute
				} else {
					aPrefix := fmt.Sprintf("`%s` (%s)", aName, aType)
					aDescr = strings.TrimSpace(strings.Split(aData, aPrefix)[1])
				}

				// Save data. Skip "timeouts" block and attributes with empty descriptions
				if len(aDescr) > 1 && currentBlock != "timeouts" {
					// If current Block already exists just add attribute to it
					if entry, ok := data[currentBlock]; ok {
						entry[aName] = map[string]string{
							"type":  aType,
							"descr": aDescr,
							"link":  aLink,
						}
						// If current Block is new, init it and add attribute
					} else {
						data[currentBlock] = make(map[string]map[string]string)
						data[currentBlock][aName] = map[string]string{
							"type":  aType,
							"descr": aDescr,
							"link":  aLink,
						}
					}
				}
			}
		}

	}
	return data
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
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("Error running tfplugindocs: %s\n", err)
		return
	}
	err = postProcessingDocs(filepath.Join(docsDir, "resources"))
	if err != nil {
		log.Fatalf("Error post processing docs: %s\n", err)
		return
	}
}
