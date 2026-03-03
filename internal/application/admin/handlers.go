package admin

import (
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/Dhinihan/chirpy/internal/application"
	"github.com/Dhinihan/chirpy/internal/database"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	Db             *database.Queries
}

func NewApiConfig(db *database.Queries) *ApiConfig {
	return &ApiConfig{Db: db}
}

func RegisterHandlers(cfg *ApiConfig, serverMux *http.ServeMux) {
	serverMux.HandleFunc("GET /admin/metrics", cfg.HandleMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.HandleReset)
}

func (cfg *ApiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w, req)
		},
	)
}

func (cfg *ApiConfig) HandleMetrics(
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

func (cfg *ApiConfig) HandleReset(
	w http.ResponseWriter,
	req *http.Request,
) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	if err := cfg.Db.ResetUsers(req.Context()); err != nil {
		application.RespondWithError(w, 500, "Erro ao limpar usuários", err)
		return
	}

	fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits.Load())
}
