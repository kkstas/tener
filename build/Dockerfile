FROM golang:1.23-alpine

WORKDIR /app

RUN apk add --no-cache curl
RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.10/tailwindcss-linux-x64 \
        && chmod +x tailwindcss-linux-x64 \
        && mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/air-verse/air@latest

COPY . ./
RUN go mod download && go mod verify

COPY . ./
EXPOSE 8080
