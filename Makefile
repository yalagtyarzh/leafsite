.PHONY:
.SILENT:

build:
	go build -o ./.bin/leafsite cmd/web/*.go

run: build
#Specify dbname, dbuser required, dbpass and production are optional 
	./.bin/leafsite -dbname= -dbuser= -production=false -dbpass=