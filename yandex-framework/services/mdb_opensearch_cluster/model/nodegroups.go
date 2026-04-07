package model

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

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

func sameSubnets(ctx context.Context, l types.List, s []string) bool {
	if l.IsUnknown() {
		return false
	}

	if len(l.Elements()) != len(s) {
		// special case for unique environment
		if len(l.Elements()) == 1 {
			return containsExactSameSubnet(ctx, l, s)
		}
		return false
	}

	slices.Sort(s)

	fromList := make([]string, 0, len(l.Elements()))
	l.ElementsAs(ctx, &fromList, false)
	slices.Sort(fromList)

	return slices.Equal(fromList, s)
}

func containsExactSameSubnet(ctx context.Context, l types.List, s []string) bool {
	// for some environment it can be so
	fromList := make([]string, 0, len(l.Elements()))
	l.ElementsAs(ctx, &fromList, false)
	for _, elem := range s {
		if elem != fromList[0] {
			return false
		}
	}

	return true
}
