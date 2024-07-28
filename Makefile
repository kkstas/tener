FUNCTION_NAME='tjener-lambda'

build_lambda:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	zip lambda-handler.zip bootstrap

push_lambda: build_lambda
	aws lambda update-function-code --function-name $(FUNCTION_NAME) --zip-file fileb://lambda-handler.zip > /dev/null
	rm lambda-handler.zip
	rm bootstrap
