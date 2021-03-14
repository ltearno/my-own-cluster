export APP_NAME := my-own-cluster

IMAGE := my-own-cluster:latest
COMMIT := latest

all: run-serve

.PHONY: build-prepare
build-prepare:
	@echo "updating dependencies..."
	GOPATH=$(dir $(realpath $(firstword $(MAKEFILE_LIST)))) GO111MODULE="off" go get github.com/jteeuwen/go-bindata/...
	@sudo apt install -y libgbm-dev libegl-dev

.PHONY: build-embed-assets
build-embed-assets:
	@echo "embedding assets..."
	@./bin/go-bindata -o assetsgen/assets.go -pkg assetsgen assets/...

.PHONY: build-apis
build-apis:
	cd api-generator && node index.js && cd ..

.PHONY: build
build: build-embed-assets
	@echo "build binaries..."
	@go build github.com/ltearno/my-own-cluster

build-releases: build-embed-assets
	@echo building release files...
	@./build-releases.sh

.PHONY: run-serve
run-serve: build-embed-assets tls.cert.pem
	@echo "run binaries..."
	@go run github.com/ltearno/my-own-cluster serve
	# -trace true

tls.cert.pem:
	@echo Generating TLS key files, you can leave default values everywhere by typing [ENTER] until the end
	@openssl req -x509 -newkey rsa:4096 -keyout tls.key.pem -nodes -out tls.cert.pem -days 3650 -subj "/C=FR/O=Rezilio/ST=Occitanie/L=Toulouse/CN=localhost"

docker-build: tls.cert.pem
	@echo building docker image
	@docker build . --build-arg UID=$(shell id -u) --build-arg GID=$(shell id -g) -t my-own-cluster:latest

docker-run:
	@docker run --rm -it -p 8443:8443 -v $(shell pwd)/my-own-cluster-database-provisional:/data/my-own-cluster-database-provisional $(IMAGE)

run-server:
	docker stop $(APP_NAME) || echo "$(APP_NAME) already stopped"
	docker rm $(APP_NAME) || echo "$(APP_NAME) already removed"
	mkdir -p $(HOME)/$(APP_NAME)-data
	docker run --name $(APP_NAME) -d --restart always \
		-u $(shell id -u) \
		-p 9870:8443 \
		-p 9871:8444 \
		-v $(HOME)/$(APP_NAME)-data:/data/my-own-cluster-database-provisional \
	    $(IMAGE)

.PHONY: core-api
core-api:
	@cd core-api && make all

clean-db:
	rm -rf my-own-cluster-database-provisional/

.PHONY: install
install: build
	sudo cp my-own-cluster /usr/local/bin/

.PHONY: install-systemd
install-systemd: install
	sudo cp my-own-cluster.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable my-own-cluster.service