package tjener

import (
	"fmt"
	"net/http"
)

func NewServer() http.Handler {
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
