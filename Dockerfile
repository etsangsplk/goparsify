FROM golang:latest

WORKDIR /goparsify

COPY ./ /goparsify

RUN go mod download

ENTRYPOINT go run example/main.go