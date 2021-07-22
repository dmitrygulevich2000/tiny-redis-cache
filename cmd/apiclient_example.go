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
	var (
		resString string
		resISlice []interface{}
		resInt int
	)

	ires, err := api.Set("K", "V", time.Second)
	if err != nil {
		log.Fatalln(err)
	}
	resString = ires.(string)
	fmt.Printf("SET(\"K\", \"V\", 1s) result: %s\n", resString)

	ires, err = api.Get("K")
	if err != nil {
		log.Fatalln(err)
	}
	resString = ires.(string)
	fmt.Printf("GET(\"K\") result: %s\n", resString)

	ires, err = api.Set("KK", []string{"a", "b"}, 0)
	if err != nil {
		log.Fatalln(err)
	}
	resString = ires.(string)
	fmt.Printf("SET(\"KK\", [\"a\", \"b\"], 0) result: %s\n", resString)

	ires, err = api.Get("KK")
	if err != nil {
		log.Fatalln(err)
	}
	resISlice = ires.([]interface{})
	fmt.Printf("GET(\"KK\") result: %#v\n", resISlice)


	fmt.Println("Sleepping 1s...")
	time.Sleep(time.Second)


	ires, err = api.Get("K")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("GET(\"K\") result: %#v\n", ires)

	resInt, err = api.Del("K", "KK")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("DEL(\"K\", \"KK\", \"KKK\") result: %d\n", resInt)
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