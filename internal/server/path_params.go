package server

import (
	"net/url"
	"strings"
)

func pathParamAfterPrefix(path string, prefix string) (string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", false
	}

	value := strings.TrimPrefix(path, prefix)
	value = strings.TrimSpace(value)

	if value == "" {
		return "", false
	}

	if strings.Contains(value, "/") {
		return "", false
	}

	decodedValue, err := url.PathUnescape(value)
	if err != nil {
		return "", false
	}

	decodedValue = strings.TrimSpace(decodedValue)

	if decodedValue == "" {
		return "", false
	}

	if strings.Contains(decodedValue, "/") {
		return "", false
	}

	return decodedValue, true
}
