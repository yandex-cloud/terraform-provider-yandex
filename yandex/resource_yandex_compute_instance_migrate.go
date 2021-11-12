package yandex

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func resourceComputeInstanceMigrateState(v int, is *terraform.InstanceState, meta interface{}) (*terraform.InstanceState, error) {
	if is.Empty() {
		log.Println("[DEBUG] Empty InstanceState; nothing to migrate.")
		return is, nil
	}

	var err error

	switch v {
	case 0:
		log.Println("[INFO] Found Compute Instance State v0; migrating to v1")
		is, err = migrateStateV0toV1(is)
		if err != nil {
			return is, err
		}
		// when adding case 1, make sure to turn this into a fallthrough
		return is, err
	default:
		return is, fmt.Errorf("Unexpected schema version: %d", v)
	}
}

func migrateStateV0toV1(is *terraform.InstanceState) (*terraform.InstanceState, error) {
	log.Printf("[DEBUG] Attributes before migration: %#v", is.Attributes)

	newResourcesMap := make(map[string]string)

	for k, value := range is.Attributes {
		if !strings.HasPrefix(k, "resources.") {
			continue
		}

		if k == "resources.#" {
			continue
		}

		// Key is now of the form resources.%d.{core,core_fraction,memory}
		keyParts := strings.Split(k, ".")

		// Sanity check
		badFormat := false
		if len(keyParts) != 3 {
			badFormat = true
		} else if _, err := strconv.Atoi(keyParts[1]); err != nil {
			badFormat = true
		}

		if badFormat {
			return is, fmt.Errorf(
				"migration error: found resource key in unexpected format: %s", k)
		}

		// New key formed without Set hashcode, just List 0-indexed
		newKey := fmt.Sprintf("%s.0.%s", keyParts[0], keyParts[2])
		newResourcesMap[newKey] = value

		delete(is.Attributes, k)
	}

	for newKey, value := range newResourcesMap {
		is.Attributes[newKey] = value
	}

	log.Printf("[DEBUG] Attributes after migration: %#v", is.Attributes)
	return is, nil

}
