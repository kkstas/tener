# Development Setup

## Prerequisites for local development

- Go
- Docker & docker-compose
- Make

## Build and run

To build and run the Docker containers for the app and dynamodb-local:

```bash
make dev-build && make dev-start
```

## Hot Reload

Hot reloading (with browser refresh) is enabled via `air`, app is proxied from port 8081 and accessible at
```sh
http://localhost:8080
```

# Makefile Targets

- `dev-build` - builds app & local DynamoDB containers
- `dev-start` - starts the containers
- `dev-down` - stops and removes the containers
- `build-lambda` - builds the app as Lambda function and packages it into a zip file
- `push-lambda` - updates the Lambda function on AWS using the built zip file
