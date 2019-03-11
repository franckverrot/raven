BINARY_NAME=raven

all: run
.PHONY: deps
deps:
	go get ./...

.PHONY: build
build: deps
	go build -o $(BINARY_NAME)

.PHONY: run
run: build
	./raven


.PHONY: envoy
envoy:
	examples/run_envoy.sh