package sweep

import "testing"

func TestSweepGroupsWithoutOrganizationIDIsNoop(t *testing.T) {
	t.Setenv("YC_TOKEN", "")
	t.Setenv("YC_SERVICE_ACCOUNT_KEY_FILE", "")

	if err := sweepAllGroupsForOrganization(""); err != nil {
		t.Fatalf("sweeper without organization ID must be a no-op: %v", err)
	}
}

func TestSweepUserSshKeysWithoutOrganizationIDIsNoop(t *testing.T) {
	t.Setenv("YC_TOKEN", "")
	t.Setenv("YC_SERVICE_ACCOUNT_KEY_FILE", "")

	if err := sweepAllUserSshKeysForOrganization(""); err != nil {
		t.Fatalf("sweeper without organization ID must be a no-op: %v", err)
	}
}

func TestLegacyOrganizationSweepersRequireAcceptanceMode(t *testing.T) {
	t.Setenv("TF_ACC", "")
	t.Setenv("YC_ORGANIZATION_ID", "organization")
	t.Setenv("YC_TOKEN", "")
	t.Setenv("YC_SERVICE_ACCOUNT_KEY_FILE", "")
	for name, run := range map[string]func(string) error{
		"groups":      sweepAllGroups,
		"federations": sweepAllSamlFederations,
		"SSH keys":    sweepAllUserSshKeys,
	} {
		if err := run(""); err != nil {
			t.Errorf("%s: expected no-op without TF_ACC: %v", name, err)
		}
	}
}
