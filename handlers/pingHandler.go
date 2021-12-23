package handlers

import "net/http"

var pingByte = []byte{}

func PingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(pingByte)
		return
	})
}
