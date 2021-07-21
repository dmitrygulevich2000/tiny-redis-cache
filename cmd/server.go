package main

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/api"

	"net/http"
)

func main() {
	api := api.New()

	mux := http.NewServeMux()
	mux.HandleFunc("/set/", api.HandleSet)
	mux.HandleFunc("/get/", api.HandleGet)
	mux.HandleFunc("/del/", api.HandleDel)

	server := http.Server {
		Addr: ":9000",
		Handler: mux,
	}

	panic(server.ListenAndServe())
}