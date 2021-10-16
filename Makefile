# Optionally set these args as environment vars in the shell. You
# could also pass them as parameters of `make`.
# For example: make test FLAGS=-race
CMD?=api
FLAGS?=-v

default: lint test

lint:
	@golangci-lint run -v ./...

test:
	@go test $(FLAGS) ./...
