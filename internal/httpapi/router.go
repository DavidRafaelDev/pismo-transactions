package httpapi

import "net/http"

func NewRouter(p Pinger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", HealthHandler(p))
	return mux
}
