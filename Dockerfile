FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /chatterbox-api ./cmd/api

FROM alpine:latest
COPY --from=builder /chatterbox-api /chatterbox-api
EXPOSE 8080
ENTRYPOINT ["/chatterbox-api"]
