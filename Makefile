export GOPATH := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
export APP_NAME := my-own-cluster

COMMIT := latest

all: run-serve

.PHONY: build-prepare
build-prepare:
	@echo "updating dependencies..."
	@./update-dependencies.sh

.PHONY: build-embed-assets
build-embed-assets:
	@echo "embedding assets..."
	@./bin/go-bindata -o src/my-own-cluster/assetsgen/assets.go -pkg assetsgen assets/...

.PHONY: build
build: build-embed-assets
	@echo "build binaries..."
	@go build my-own-cluster

build-releases: build-embed-assets
	@echo building release files...
	@./build-releases.sh

.PHONY: run-serve
run-serve: build-embed-assets tls.cert.pem
	@echo "run binaries..."
	@go run my-own-cluster serve

tls.cert.pem:
	@echo Generating TLS key files, you can leave default values everywhere by typing [ENTER] until the end
	@openssl req -x509 -newkey rsa:4096 -keyout tls.key.pem -nodes -out tls.cert.pem -days 365

docker-build: tls.cert.pem
	@echo building docker image
	@docker build . -t my-own-cluster:latest

docker-run:
	@docker run --rm -it -p 8443:8443 -v $(shell pwd)/my-own-cluster-database-provisional:/data/my-own-cluster-database-provisional my-own-cluster:latest