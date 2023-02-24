package urnx

import (
	"regexp"
)

var (
	nsrx *regexp.Regexp
	rtrx *regexp.Regexp
)

func init() {
	nsrx = regexp.MustCompile("^[A-Za-z0-9-]{1,30}$")
	rtrx = regexp.MustCompile("^[A-Za-z0-9-]{1,255}$")
}

func validateNamespace(namespace string) bool {
	rx := nsrx.MatchString(namespace)

	return rx
}

func validateResourceType(name string) bool {
	rx := rtrx.MatchString(name)

	return rx
}
