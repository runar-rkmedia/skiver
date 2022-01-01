package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Extracts queryparameters from path and query-params.
// This is useful for caching
func extractParams(r *http.Request, basePath string) (url.Values, error) {
	q := r.URL.Query()

	qpath := strings.Replace(r.URL.Path, basePath, "", 1)
	qpath, _, _ = strings.Cut(qpath, ".")
	qpath = strings.TrimPrefix(qpath, "/")
	fmt.Println(qpath, basePath)

	qq, err := url.ParseQuery(qpath)
	if err != nil {
		return q, err
	}

	for k, v := range qq {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	return q, nil
}
