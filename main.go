package main

import (
	"fmt"
	"net/http"

	"github.com/Dhinihan/chirpy/internal/application/admin"
	"github.com/Dhinihan/chirpy/internal/application/api"
)

func main() {
	appHandler := http.StripPrefix(
		"/app",
		http.FileServer(http.Dir("./app")),
	)

	cfg := admin.NewApiConfig()

	serverMux := http.NewServeMux()
	serverMux.Handle(
		"/app/",
		cfg.MiddlewareMetricsInc(appHandler),
	)
	api.RegisterHandlers(serverMux)
	admin.RegisterHandlers(cfg, serverMux)
	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Servidor parou por: \n%s\n", err.Error())
	}
}
