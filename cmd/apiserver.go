package main

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/api/server"

	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Expected server port in first arg")
	}
	port, err := strconv.Atoi(os.Args[1]) 
	if err != nil {
		log.Fatalln("Expected server port in first arg")
	}
	
	srv := server.New()
	server := http.Server {
		Addr: ":" + strconv.Itoa(port),
		Handler: srv,
	}

	panic(server.ListenAndServe())
}