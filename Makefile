.PHONY: all build run clean serve test

all: run

build:
	go build -o ssg ./cmd/ssg

run: build
	./ssg

clean:
	rm -rf public ssg

serve: run
	go run ./cmd/server

test:
	go test ./...
