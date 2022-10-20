FROM golang:alpine as builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download 

COPY ./url-collector /app/url-collector

WORKDIR /app/url-collector

RUN go build -o main .

# Reduce image size
FROM alpine:latest

COPY --from=builder /app/url-collector/main .

ENTRYPOINT [ "/main" ]