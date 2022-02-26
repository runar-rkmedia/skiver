package handlers

import (
	"net/http"
	"net/url"
	"strings"
)

// Extracts queryparameters from path and query-params.
// This is useful for caching
func ExtractParams(r *http.Request) (url.Values, error) {
	q := r.URL.Query()

	qpath := StripBeforeLast(r.URL.Path, "/")

	// qpath, _, _ = strings.Cut(qpath, ".")
	qpath = strings.TrimPrefix(qpath, "/")

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

// Removes the characters before the last occureance of `sep` and the `sep` if available
func StripBeforeLast(s, sep string) string {
	if idx := strings.LastIndex(s, sep); idx != -1 {
		return s[idx+1:]
	}
	return s
}
