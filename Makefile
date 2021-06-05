DOCKER_TAG ?= $(shell git rev-parse HEAD)

.PHONY: default
default: build

.PHONY: docker-build
docker-build:
	docker build \
		-t skuttle:$(DOCKER_TAG) \
		-t skuttle:latest \
		--cache-from skuttle:$(DOCKER_TAG) \
		.

.PHONY: clean
clean:
	rm -f skuttle

.PHONY: build
build: test
	CGO_ENABLED=0 go build ./cmd/skuttle

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	go fmt ./...
