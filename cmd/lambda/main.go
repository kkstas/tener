package main

import (
	"github.com/kkstas/tjener/internal"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

func main() {
	adapter := httpadapter.New(tjener.NewServer())
	lambda.Start(adapter.ProxyWithContext)
}
