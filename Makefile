PROTOC_GEN_PATH := $(shell go env GOPATH)/bin
VERSION := 1.0.0
COMMIT_HASH := $(shell git rev-parse HEAD)

generate:
	protoc -I=. -I=$(PROTOC_GEN_PATH)/../pkg/mod -I=/usr/local/include \
		--go_out=internal/generated --go_opt=paths=source_relative \
		--go-grpc_out=internal/generated --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=internal/generated --grpc-gateway_opt paths=source_relative,generate_unbound_methods=true \
		--openapiv2_out=api/ --openapiv2_opt generate_unbound_methods=true \
		proto/api.proto

	wire internal/wiring/wire.go

build:
	go build \
		-ldflags "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH)" \
		-o build/ \
		cmd/*.go

clean:
	rm -rf build/

run-server:
	go run cmd/*.go server