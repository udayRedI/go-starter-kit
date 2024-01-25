package lib

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

// decodeURL parses the URL path to extract app name and action.
func decodeURI(req *http.Request) (string, string, string) {
	pathTokens := strings.Split(strings.Trim(req.URL.Path, "/"), "/")

	switch len(pathTokens) {
	case 0:
		return "", "", ""
	case 1:
		return pathTokens[0], "get", ""
	case 2:
		return pathTokens[0], pathTokens[1], ""
	case 3:
		return pathTokens[0], pathTokens[1], pathTokens[2]
	default:
		return "", "", ""
	}
}

func splitStrings(values []string, delimiter string) []string {
	var splitValues []string
	for _, v := range values {
		splitValues = append(splitValues, strings.Split(v, delimiter)...)
	}
	return splitValues
}

func decodeQueryParams(query url.Values) (string, map[string][]string, []string, int, int, []string, map[string]string, bool, map[string][]string, string) {
	freeTextQuery := ""
	filter := make(map[string][]string)
	sort := []string{}
	page := 1
	pageSize := 10
	fields := []string{}
	scrollId := ""
	context := make(map[string]string)
	uiReady := false
	other := make(map[string][]string)

	for param, values := range query {
		if len(param) == 0 {
			continue
		}

		for _, r := range param[:1] {
			if unicode.IsUpper(r) {
				filter[param] = splitStrings(values, ",")
			} else {
				switch param {
				case "free_text_query":
				case "q":
					if len(values) > 0 {
						freeTextQuery = values[0]
					}
				case "sort":
				case "s":
					sort = splitStrings(values, ",")
				case "page":
				case "p":
					if len(values) > 0 {
						p, err := strconv.Atoi(values[0])
						if err == nil {
							page = p
						}
					}
				case "pageSize":
				case "ps":
					if len(values) > 0 {
						ps, err := strconv.Atoi(values[0])
						if err == nil {
							if ps > 10000 {
								pageSize = 10000
							} else {
								pageSize = ps
							}
						}
					}
				case "scrollId":
					if len(values) > 0 {
						scrollId = values[0]
					}
				case "fields":
				case "f":
					fields = splitStrings(values, ",")
				case "context":
					for _, c := range splitStrings(values, ",") {
						c = strings.TrimSpace(c)
						tokens := strings.SplitN(c, ":", 2)
						if len(tokens) != 2 {
							continue
						}

						contextType := strings.TrimSpace(tokens[0])
						contextValue := strings.TrimSpace(tokens[1])
						if contextType == "" || contextValue == "" {
							continue
						}

						context[contextType] = contextValue
					}
				case "uiReady":
				case "ur":
					uiReady = true
				default:
					other[param] = splitStrings(values, ",")
				}
			}
		}
	}

	return freeTextQuery, filter, sort, page, pageSize, fields, context, uiReady, other, scrollId
}
