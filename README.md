# Development Setup

## Prerequisites

- Docker
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
