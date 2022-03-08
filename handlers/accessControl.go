package handlers

import (
	"fmt"
	"net/http"
	"time"
)

type AccessControl struct {
	AllowOrigin string
	MaxAge      time.Duration
}

var (
	accessControl = AccessControl{
		AllowOrigin: "_any_",
		MaxAge:      24 * time.Hour,
	}
	accessControlMaxAgeString = fmt.Sprintf("%.f", accessControl.MaxAge.Seconds())
)

func AddAccessControl(r *http.Request, rw http.ResponseWriter) {

	h := rw.Header()
	switch accessControl.AllowOrigin {
	case "_any_":
		h.Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	default:
		h.Set("Access-Control-Allow-Origin", accessControl.AllowOrigin)
	}
	h.Set("Access-Control-Allow-Headers", "x-request-id, content-type, jmes-path")
	h.Set("Access-Control-Max-Age", accessControlMaxAgeString)
	if r.Method == "OPTIONS" {
		h.Set("Cache-Control", "public, max-age=%0.f"+accessControlMaxAgeString)
		h.Set("Vary", "origin")

		return
	}
}
