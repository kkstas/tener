FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache curl
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.10/tailwindcss-linux-x64 \
        && chmod +x tailwindcss-linux-x64 \
        && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

COPY go.mod .

RUN go install github.com/air-verse/air@v1.61.1 && \
        TEMPL_VERSION=$(grep "github.com/a-h/templ" go.mod | sed -E 's/.*v([0-9]+\.[0-9]+\.[0-9]+).*/\1/') && \
        echo "Installing templ version: v${TEMPL_VERSION}" && \
        go install github.com/a-h/templ/cmd/templ@v${TEMPL_VERSION}

COPY . ./
RUN go mod download && go mod verify

COPY . ./
EXPOSE 8080
