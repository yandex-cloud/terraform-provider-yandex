package model

type withName interface {
	GetName() string
}

func GetGroupByName[T withName](groups []T) map[string]T {
	groupsByName := make(map[string]T, len(groups))
	for _, g := range groups {
		groupsByName[g.GetName()] = g
	}

	return groupsByName
}

func getGroupNames[T withName](groups []T) []string {
	names := make([]string, len(groups))
	for i, g := range groups {
		names[i] = g.GetName()
	}

	return names
}
