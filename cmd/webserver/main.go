package main

import (
	"log"
	"net/http"
	"time"
)

func main() {
	app := initApplicationAndDDB()

	server := &http.Server{
		Addr:              ":8081",
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           app,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
