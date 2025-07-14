
_default:
  @just --list

api: build
  LISTEN_ADDR=127.0.0.1:8080 ./bin/chedul-backend 

build:
  go build -o bin/chedul-backend cmd/api/main.go
