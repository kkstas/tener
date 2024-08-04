package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

func main() {
	app := initApplication()
	lambda.Start(httpadapter.New(app).ProxyWithContext)
}
