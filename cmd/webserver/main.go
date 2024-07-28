package main

import (
	"log"
	"net/http"

	"github.com/kkstas/tjener/internal"
)

func main() {
	if err := http.ListenAndServe(":8080", tjener.NewServer()); err != nil {
		log.Fatal(err)
	}
}
