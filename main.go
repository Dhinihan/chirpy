package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

func main() {
	appHandler := http.StripPrefix(
		"/app",
		http.FileServer(http.Dir("./app")),
	)

	cfg := apiConfig{}

	serverMux := http.NewServeMux()
	serverMux.Handle(
		"/app/",
		cfg.middlewareMetricsInc(appHandler),
	)
	serverMux.HandleFunc("/healthz", handleHealthZ)
	serverMux.HandleFunc("/metrics", cfg.handleMetrics)
	serverMux.HandleFunc("/reset", cfg.handleReset)
	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Servidor parou por: \n%s\n", err.Error())
	}
}

func handleHealthZ(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, req *http.Request) {
			cfg.fileserverHits.Add(1)
			next.ServeHTTP(w, req)
		})

}

func (cfg *apiConfig) handleMetrics(
	w http.ResponseWriter,
	req *http.Request,
) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits.Load())
}

func (cfg *apiConfig) handleReset(
	w http.ResponseWriter,
	req *http.Request,
) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	cfg.fileserverHits.Store(0)
	fmt.Fprintf(w, "Hits: %d\n", cfg.fileserverHits.Load())
}
