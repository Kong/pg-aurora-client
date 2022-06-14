GO111MODULES=on
APP=pg-aurora-client
COMMIT_SHA=$(shell git rev-parse --short HEAD)
APP_VERSION=1.7

default: help

.PHONY: build
## build:
build:
	CGO_ENABLED=0 go build -o ${APP} cmd/main.go cmd/config.go cmd/helpers.go cmd/routes.go

.PHONY: docker-push
## docker-push: build and push image to docker hub
docker-push:
	docker build . -t  kongcloud/pg-aurora-client:${APP_VERSION}
	docker push kongcloud/pg-aurora-client:${APP_VERSION}
	docker push kongcloud/pg-aurora-client:${APP_VERSION}


.PHONY: clean
## clean: removes the binary files
clean:
	rm -f ${APP}

.PHONY: help
## help: Prints this help message
help:
	@echo "Usage: \n"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
