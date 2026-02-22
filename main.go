package main

import (
	"fmt"
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: serverMux,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("Servidor parou por: \n%s\n", err.Error())
	}
}
