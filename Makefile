.PHONY: build install proto clean

build:
	go build -o bin/protoc-gen-ai-context ./cmd/protoc-gen-ai-context

install:
	go install ./cmd/protoc-gen-ai-context

proto:
	buf generate

clean:
	rm -rf bin/
