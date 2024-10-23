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

Hot reloading (with browser refresh) is enabled via `air`, app is proxied from
port 8081 and accessible at

```sh
http://localhost:8080
```

# Environment variables

| Variable         | Description                                             | Type                                                   | Required | Default                                |
| ---------------- | ------------------------------------------------------- | ------------------------------------------------------ | -------- | -------------------------------------- |
| `DDB_TABLE_NAME` | DynamoDB table name                                     | `string`                                               | true     | -                                      |
| `LOG_LEVEL`      | Max log level app will emit                             | One of: `trace` `debug` `info` `error` `fatal` `panic` | false    | `trace` on webserver, `warn` on lambda |
| `TOKEN_SECRET`   | secret key for signing and verifying HMAC-SHA256 tokens | `string`                                               | true     | -                                      |
