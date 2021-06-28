DOCKER_TAG ?= $(shell git rev-parse HEAD)

.PHONY: default
default: test build

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
build:
	CGO_ENABLED=0 go build ./cmd/skuttle

.PHONY: test
test:
	go install github.com/onsi/ginkgo/ginkgo@v1.16.4
	ginkgo -r -skipPackage integration

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: integration-test
integration-test: build
	go install sigs.k8s.io/kind@v0.11.1
	integration/test.sh
