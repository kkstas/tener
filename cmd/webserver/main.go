package main

import (
	"log"
	"net/http"

	"github.com/kkstas/tjener/internal/database"
	"github.com/kkstas/tjener/internal/server"
)

func main() {
	app := server.NewApplication(database.CreateDynamoDBClient())
	if err := http.ListenAndServe(":8081", app); err != nil {
		log.Fatal(err)
	}
}
