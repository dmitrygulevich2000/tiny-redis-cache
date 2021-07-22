package main

import (
	"github.com/dmitrygulevich2000/tiny-redis-cache/api/client"

	"fmt"
	"log"
	"time"
)

func main() {
	c, err := client.NewClient("http://localhost:9000", time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	api := client.NewAPI(c)
	fmt.Println("Successfully created")


	ires, err := api.Set("K", "V", 0)
	if err != nil {
		log.Fatalln(err)
	}
	res := ires.(string)
	fmt.Printf("Set result: %s\n", res)


	ires, err = api.Get("K")
	if err != nil {
		log.Fatalln(err)
	}
	res = ires.(string)
	fmt.Printf("Get result: %s\n", res)


	deleted, err := api.Del("K", "KK")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("Del result: %d\n", deleted)
}