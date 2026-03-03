package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/application/api"
	"github.com/Dhinihan/chirpy/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	appHandler := http.StripPrefix(
		"/app",
		http.FileServer(http.Dir("./app")),
	)

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Erro ao abrir conexão com o banco de dados:\n%s", err.Error())
	}
	cfg := admin.NewApiConfig(database.New(db))

	serverMux := http.NewServeMux()
	serverMux.Handle(
		"/app/",
		cfg.MiddlewareMetricsInc(appHandler),
	)
	api.RegisterHandlers(cfg, serverMux)
	admin.RegisterHandlers(cfg, serverMux)
	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("Servidor parou por: \n%s\n", err.Error())
	}
}
