package helpers

import (
	"os"
	"strings"
)

func EnforceHTTP(url string) string {
	if !strings.HasPrefix(url, "http") {
		return "https://" + url
	}
	return url
}

func RemoveDomainError(url string) bool {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "www.")
	return strings.SplitN(url, "/", 1)[0] != os.Getenv("DOMAIN")
}
