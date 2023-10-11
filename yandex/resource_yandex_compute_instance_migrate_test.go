package yandex

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestComputeInstanceMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
	}{
		"change resources from set to list": {
			StateVersion: 0,
			Attributes: map[string]string{
				"resources.#":                        "1",
				"resources.1690069307.core_fraction": "100",
				"resources.1690069307.cores":         "1",
				"resources.1690069307.memory":        "2",
			},
			Expected: map[string]string{
				"resources.#":               "1",
				"resources.0.core_fraction": "100",
				"resources.0.cores":         "1",
				"resources.0.memory":        "2",
			},
		},
	}

	for tn, tc := range cases {
		runInstanceMigrateTest(t, tn, tc.StateVersion, tc.Attributes, tc.Expected, nil)
	}
}

func TestComputeInstanceMigrateState_empty(t *testing.T) {
	if os.Getenv(resource.TestEnvVar) == "" {
		t.Skipf("Network access not allowed; use %s=1 to enable", resource.TestEnvVar)
	}
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceComputeInstanceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = resourceComputeInstanceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}

func runInstanceMigrateTest(t *testing.T, testName string, version int, attributes, expected map[string]string, meta interface{}) {
	is := &terraform.InstanceState{
		ID:         "sometestid",
		Attributes: attributes,
	}
	_, err := resourceComputeInstanceMigrateState(version, is, meta)
	if err != nil {
		t.Fatal(err)
	}

	for k, v := range expected {
		if attributes[k] != v {
			t.Fatalf(
				"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
				testName, k, expected[k], k, attributes[k], attributes)
		}
	}

	for k, v := range attributes {
		if expected[k] != v {
			t.Fatalf(
				"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
				testName, k, expected[k], k, attributes[k], attributes)
		}
	}
}
