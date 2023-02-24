package urnx

import (
	"regexp"
)

func validateNamespace(namespace string) bool {
	r := regexp.MustCompile("^[A-Za-z0-9-]{1,30}$")
	rx := r.MatchString(namespace)

	return rx
}

func validateResourceType(name string) bool {
	r := regexp.MustCompile("^[A-Za-z0-9-]{1,255}$")
	rx := r.MatchString(name)

	return rx
}
