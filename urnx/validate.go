package urnx

import (
	"regexp"
)

func validateNamespace(namespace string) (bool, error) {
	rx, err := regexp.MatchString("^[A-Za-z0-9-]{1,30}$", namespace)
	if err != nil {
		return false, err
	}

	return rx, nil
}
