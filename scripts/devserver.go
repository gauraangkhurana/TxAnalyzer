// Command devserver is a Docker-free local development helper - NOT part of
// the real deployment (that's docker-compose.yml + gateway/nginx.conf).
// It serves gateway/data/www as static files and reverse-proxies
// /v1/bank -> localhost:10001 and /v1/users -> localhost:10002, mirroring
// nginx.conf's routing so the frontend's relative fetch() calls work
// without CORS, exactly like they will under the real nginx setup.
//
// Usage (from the repo root, with bank-svc and user-svc already running
// locally on ports 10001/10002):
//
//	go run ./scripts/devserver.go
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func mustProxy(target string) *httputil.ReverseProxy {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("invalid proxy target %q: %v", target, err)
	}
	return httputil.NewSingleHostReverseProxy(u)
}

func main() {
	bankProxy := mustProxy("http://localhost:10001")
	userProxy := mustProxy("http://localhost:10002")

	mux := http.NewServeMux()
	mux.Handle("/v1/bank/", bankProxy)
	mux.Handle("/v1/users", userProxy)
	mux.Handle("/v1/users/", userProxy)
	mux.Handle("/", http.FileServer(http.Dir("gateway/data/www")))

	addr := ":8080"
	log.Printf("devserver listening on %s (static: gateway/data/www, proxying /v1/bank -> :10001, /v1/users -> :10002)", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
