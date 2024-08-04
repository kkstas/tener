package main

import (
	"log"
	"net/http"
)

func main() {
	app := initApplicationAndDDB()
	if err := http.ListenAndServe(":8081", app); err != nil {
		log.Fatal(err)
	}
}
