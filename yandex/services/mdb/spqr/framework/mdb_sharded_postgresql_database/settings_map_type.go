package mdb_sharded_postgresql_database

type SettingsAttributeInfoProvider struct{}

var attrProvider = &SettingsAttributeInfoProvider{}

var settingsEnumNames = map[string]map[int32]string{}

var settingsEnumValues = map[string]map[string]int32{}

var listAttributes = map[string]struct{}{}

func (p *SettingsAttributeInfoProvider) GetSettingsEnumNames() map[string]map[int32]string {
	return settingsEnumNames
}

func (p *SettingsAttributeInfoProvider) GetSettingsEnumValues() map[string]map[string]int32 {
	return settingsEnumValues
}

func (p *SettingsAttributeInfoProvider) GetSetAttributes() map[string]struct{} {
	return listAttributes
}
