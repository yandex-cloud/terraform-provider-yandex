package yandex

import "github.com/yandex-cloud/go-genproto/yandex/cloud/access"

func memberType(ab *access.AccessBinding) string {
	return ab.Subject.Type
}

func memberID(ab *access.AccessBinding) string {
	return ab.Subject.Id
}

func userAccountIDs(p *Policy) []string {
	usersMap := map[string]bool{}
	for _, b := range p.Bindings {
		if memberType(b) == "userAccount" {
			usersMap[memberID(b)] = true
		}
	}
	userIDs := []string{}
	for userID := range usersMap {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
