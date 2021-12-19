.PHONY:
.SILENT:

build:
	go build -o ./.bin/leafsite cmd/web/*.go

run: build
	./.bin/leafsite