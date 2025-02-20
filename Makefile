.PHONY: docs

default: help

## all : Builds the program for both architecture
all: client server

## client : Builds the client from ./cmd/client
client:
	go build ./cmd/client

## server : Builds the server from ./cmd/server
server:
	go build ./cmd/server

## help : Shows this help
help: Makefile
	@printf ">] Excuses\n\n"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@printf ""
