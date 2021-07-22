package main

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/api/server"

	"net/http"
)

func main() {
	srv := server.New()

	server := http.Server {
		Addr: ":9000",
		Handler: srv,
	}

	panic(server.ListenAndServe())
}