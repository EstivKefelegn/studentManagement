package utils

import "errors"

type ContextKey string

func AuthorizeUser(userRole string, allowedRoles ...string) (bool, error) {
	for _, alloallowedRole := range allowedRoles {
		if userRole == alloallowedRole {
			return true, nil
		}
	}
	return false, errors.New("user not authorized")
}
