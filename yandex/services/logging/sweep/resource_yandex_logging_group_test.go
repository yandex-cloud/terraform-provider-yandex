package sweep

import (
	"strings"
	"testing"
)

func TestSweepYandexLoggingGroupSkip(t *testing.T) {
	t.Setenv("YC_TOKEN", "")
	t.Setenv("YC_SERVICE_ACCOUNT_KEY_FILE", "")

	t.Run("enabled", func(t *testing.T) {
		t.Setenv("YC_SKIP_LOGGING_GROUP_SWEEPING", "1")
		if err := sweepAllYandexLoggingGroup(""); err != nil {
			t.Fatalf("skipping logging sweeper must not require credentials: %v", err)
		}
	})

	t.Run("default", func(t *testing.T) {
		t.Setenv("YC_SKIP_LOGGING_GROUP_SWEEPING", "")
		err := sweepAllYandexLoggingGroup("")
		if err == nil || !strings.Contains(err.Error(), "YC_TOKEN or YC_SERVICE_ACCOUNT_KEY_FILE must be set") {
			t.Fatalf("logging sweeper must still require credentials by default, got: %v", err)
		}
	})
}
