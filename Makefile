PROTOC_GEN_PATH := $(shell go env GOPATH)/bin

generate:
	protoc -I=. -I=$(PROTOC_GEN_PATH)/../pkg/mod -I=/usr/local/include \
		--go_out=internal/generated --go_opt=paths=source_relative \
		--go-grpc_out=internal/generated --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=internal/generated --grpc-gateway_opt paths=source_relative,generate_unbound_methods=true \
		--openapiv2_out=api/ --openapiv2_opt generate_unbound_methods=true \
		proto/api.proto

	wire internal/wiring/wire.go