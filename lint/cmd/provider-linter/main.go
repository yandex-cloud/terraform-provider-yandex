package main

import (
	"github.com/bflad/tfproviderlint/xpasses/XR005"
	"github.com/bflad/tfproviderlint/xpasses/XS001"
	"github.com/yandex-cloud/terraform-provider-yandex/lint/pkg/checks/XS003"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

var checks = []*analysis.Analyzer{
	XR005.Analyzer,
	XS001.Analyzer,
	XS003.Analyzer,
}

func main() {
	multichecker.Main(checks...)
}
