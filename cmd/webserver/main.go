package main

import (
	"log"
	"net/http"

	"github.com/kkstas/tjener/internal/server"
)

func main() {
	if err := http.ListenAndServe(":8081", server.NewApplication()); err != nil {
		log.Fatal(err)
	}
}
