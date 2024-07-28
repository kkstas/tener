package main

import (
	"fmt"
	"net/http"

	lambda "github.com/aws/aws-lambda-go/lambda"
	httpadapter "github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

func newServer() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /hello", http.HandlerFunc(handlerGetHello))
	mux.Handle("POST /hello", http.HandlerFunc(handlerPostHello))
	mux.Handle("/", http.HandlerFunc(notFound))

	return mux
}

func handlerGetHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "invoked GET /hello")
}

func handlerPostHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "invoked POST hello")
}

func notFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func main() {
	adapter := httpadapter.New(newServer())
	lambda.Start(adapter.ProxyWithContext)
}

