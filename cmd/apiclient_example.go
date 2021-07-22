package main

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/api/client"

	"fmt"
	"log"
	"os"
	"regexp"
	"time"
)

func ApiExample(api client.ClientAPI) {
	ires, err := api.Set("K", "V", 0)
	if err != nil {
		log.Fatalln(err)
	}
	res := ires.(string)
	fmt.Printf("SET result: %s\n", res)


	ires, err = api.Get("K")
	if err != nil {
		log.Fatalln(err)
	}
	res = ires.(string)
	fmt.Printf("GET result: %s\n", res)


	deleted, err := api.Del("K", "KK")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("DEL result: %d\n", deleted)
}

func main() {
	if len(os.Args) < 2  {
		log.Fatalln("Expected target api-server's address in first arg")
	}
	if match, _ := regexp.MatchString(".+:[0-9]+", os.Args[1]); !match  {
		log.Fatalln("Expected target api-server's address in form host:port")
	}

	c, err := client.NewClient("http://" + os.Args[1], time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	api := client.NewAPI(c)
	fmt.Println("api-client successfully created")

	ApiExample(api)
}