package XS003_test

import (
	"github.com/yandex-cloud/terraform-provider-yandex/lint/pkg/checks/XS003"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestXS003Analyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, XS003.Analyzer, "a")
}
