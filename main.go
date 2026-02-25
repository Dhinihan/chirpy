package main

import (
	"fmt"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()
	serverMux.Handle(
		"/app/",
		http.StripPrefix(
			"/app",
			http.FileServer(http.Dir("./app")),
		),
	)
	serverMux.HandleFunc("/healthz", handleHealthZ)
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
