package scanner

type (
	ScanStructOption interface {
		applyScanStructOption(settings *scanStructSettings)
	}
	tagName                       string
	allowMissingColumnsFromSelect struct{}
	allowMissingFieldsInStruct    struct{}
)

var (
	_ ScanStructOption = tagName("")
	_ ScanStructOption = allowMissingColumnsFromSelect{}
	_ ScanStructOption = allowMissingFieldsInStruct{}
)

func (allowMissingFieldsInStruct) applyScanStructOption(settings *scanStructSettings) {
	settings.AllowMissingFieldsInStruct = true
}

func (allowMissingColumnsFromSelect) applyScanStructOption(settings *scanStructSettings) {
	settings.AllowMissingColumnsFromSelect = true
}

func (name tagName) applyScanStructOption(settings *scanStructSettings) {
	settings.TagName = string(name)
}

func WithTagName(name string) tagName {
	return tagName(name)
}

func WithAllowMissingColumnsFromSelect() allowMissingColumnsFromSelect {
	return allowMissingColumnsFromSelect{}
}

func WithAllowMissingFieldsInStruct() allowMissingFieldsInStruct {
	return allowMissingFieldsInStruct{}
}
