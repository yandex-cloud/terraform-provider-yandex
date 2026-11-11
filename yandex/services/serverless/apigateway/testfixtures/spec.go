package testfixtures

import _ "embed"

//go:embed main.yaml
var Spec string

//go:embed main_updated.yaml
var UpdatedSpec string

//go:embed canary.yaml
var ParameterizedSpec string
