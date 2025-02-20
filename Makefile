.PHONY: docs

default: help

## build_all : Builds the program for both architecture
build_all: build_client build_server

## build_client : Builds the client from ./cmd/client
build_client:
	go build ./cmd/client

## build_server : Builds the server from ./cmd/server
build_server:
	go build ./cmd/server

## help : Shows this help
help: Makefile
	@printf ">] Excuses\n\n"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@printf ""
