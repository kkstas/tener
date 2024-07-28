package main

import (
	"github.com/kkstas/tjener/internal/server"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

func main() {
	adapter := httpadapter.New(server.NewApplication())
	lambda.Start(adapter.ProxyWithContext)
}
