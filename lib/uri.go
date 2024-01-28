package lib

import (
	"net/http"
	"strings"
)

// decodeURL parses the URL path to extract app name and action.
func decodeURI(req *http.Request) (string, string) {
	pathTokens := strings.Split(strings.Trim(req.URL.Path, "/"), "/")

	switch len(pathTokens) {
	case 1:
		return pathTokens[0], ""
	case 2:
		return pathTokens[0], pathTokens[1]
	default:
		return "", ""
	}
}

func splitStrings(values []string, delimiter string) []string {
	var splitValues []string
	for _, v := range values {
		splitValues = append(splitValues, strings.Split(v, delimiter)...)
	}
	return splitValues
}
