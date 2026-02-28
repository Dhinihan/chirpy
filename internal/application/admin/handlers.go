package admin

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func NewApiConfig() *apiConfig {
	return &apiConfig{}
}

func RegisterHandlers(cfg *apiConfig, serverMux *http.ServeMux) {
	serverMux.HandleFunc("GET /admin/metrics", cfg.HandleMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.HandleReset)
}

func (cfg *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w, req)
		},
	)
}

func (cfg *apiConfig) HandleMetrics(
	w http.ResponseWriter,
	req *http.Request,
) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	template := `<html><body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body></html>`
	fmt.Fprintf(w, template, cfg.fileserverHits.Load())
}

func (cfg *apiConfig) HandleReset(
	w http.ResponseWriter,
	req *http.Request,
) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits.Load())
}
