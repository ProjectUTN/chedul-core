GOOSE_DBSTRING := "postgres://chedul:chedul@127.0.0.1:5432/chedul-db?sslmode=disable"
MIGRATION_PATH:="migrations"

_default:
  @just --list

api build_flag="": build
  docker compose up {{ if build_flag == "rebuild" { "--build" } else { "" } }}

watch short="":
  watchexec -r -e go -- just test {{ if short == "short" { "short" } else { "" } }}

build: fmt lint
  go mod tidy
  go build -o bin/chedul-api cmd/api/main.go

fmt:
  go fmt ./...

lint:
  gocritic check ./...

goose name:
  goose -dir={{MIGRATION_PATH}} create {{name}} sql

db-status:
  GOOSE_DRIVER=postgres GOOSE_DBSTRING={{GOOSE_DBSTRING}} goose -dir={{MIGRATION_PATH}} status

up:
  GOOSE_DRIVER=postgres GOOSE_DBSTRING={{GOOSE_DBSTRING}} goose -dir={{MIGRATION_PATH}} up

down:
  GOOSE_DRIVER=postgres GOOSE_DBSTRING={{GOOSE_DBSTRING}} goose -dir={{MIGRATION_PATH}} down

reset:
  GOOSE_DRIVER=postgres GOOSE_DBSTRING={{GOOSE_DBSTRING}} goose -dir={{MIGRATION_PATH}} reset

test short="" dir="...":
  CONFIG_DIR=../configuration/ gotestsum --format testname --debug -- {{ if short == "short" { "-short"} else {""} }} ./{{dir}}

update_slow_tests:
  go test -json -short ./... | gotestsum tool slowest --skip-stmt "testing.Short" --threshold 200ms
