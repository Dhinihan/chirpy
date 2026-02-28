package main

import (
	"encoding/json"
	"fmt"
	"io"
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
	serverMux.HandleFunc("GET /api/healthz", handleHealthZ)
	serverMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)
	serverMux.HandleFunc("GET /admin/metrics", cfg.handleMetrics)
	serverMux.HandleFunc("POST /admin/reset", cfg.handleReset)
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

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	data, err := io.ReadAll(req.Body)
	if err != nil {
		respondWithError(w, 500, "Something went wrong", err)
		return
	}
	var requestData struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(data, &requestData); err != nil {
		respondWithError(w, 400, "expected json with 'body' key", err)
		return
	}
	if requestData.Body == "" {
		respondWithError(w, 400, "Chirp not informed", nil)
		return
	}
	if len(requestData.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}
	respondWithJson(w, 200, struct {
		Valid bool `json:"valid"`
	}{Valid: true})
	return
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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	template := `<html><body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body></html>`
	fmt.Fprintf(w, template, cfg.fileserverHits.Load())
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

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		fmt.Println(err.Error())
	}
	respondWithJson(w, code, struct {
		Error string `json:"error"`
	}{Error: msg})
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	js, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(js))
}
