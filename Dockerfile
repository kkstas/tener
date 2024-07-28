FROM golang:1.22 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go install github.com/air-verse/air@latest

EXPOSE 8080

CMD ["air"]
