include .env

.PHONY: dev-build dev-start clean build-lambda push-lambda

dev-build:
	docker compose -f docker-compose.yaml build

dev-start:
	docker compose -f docker-compose.yaml up

dev-down:
	docker compose -f docker-compose.yaml down

build-lambda:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./cmd/lambda
	zip lambda-handler.zip bootstrap

push-lambda: build-lambda
	aws lambda update-function-code --function-name ${DEV_FUNCTION_NAME} --zip-file fileb://lambda-handler.zip > /dev/null
	rm lambda-handler.zip
	rm bootstrap

prd-push-lambda: build-lambda
	aws lambda update-function-code --function-name ${PRD_FUNCTION_NAME} --zip-file fileb://lambda-handler.zip > /dev/null
	rm lambda-handler.zip
	rm bootstrap
