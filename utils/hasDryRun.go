package utils

import "net/http"

func HasDryRun(r *http.Request) bool {

	q := r.URL.Query()
	if q.Has("dry") {
		return true
	}
	if q.Has("dry-run") {
		return true
	}
	if q.Has("dryRun") {
		return true
	}
	if r.Header.Get("dry-run") != "" {
		return true
	}
	if r.Header.Get("dry") != "" {
		return true
	}
	if r.Header.Get("dryrun") != "" {
		return true
	}
	if r.Header.Get("x-dry") != "" {
		return true
	}
	if r.Header.Get("x-dry-run") != "" {
		return true
	}
	if r.Header.Get("x-dryrun") != "" {
		return true
	}
	return false
}
