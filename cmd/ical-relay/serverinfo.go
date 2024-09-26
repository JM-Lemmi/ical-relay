package main

import "net/http"

var serverversion string
var clientversion string

func initVersions(url string) {
	serverversion = binname + "/" + version
	clientversion = "Go-http-client/1.1 (" + binname + "/" + version + "; +" + url + ")"
}

func serverHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		w.Header().Set("Server", serverversion)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

// This implements a custom http Client that adds a User-Agent header to all requests.
// Usage:
//   client := http.Client{Transport: NewUseragentTransport(nil)}

type AddHeaderTransport struct {
	T http.RoundTripper
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", clientversion)
	return adt.T.RoundTrip(req)
}

func NewUseragentTransport(T http.RoundTripper) *AddHeaderTransport {
	if T == nil {
		T = http.DefaultTransport
	}
	return &AddHeaderTransport{T}
}
