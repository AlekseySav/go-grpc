GOBIN ?= $(shell go env GOPATH)/bin
GRPC_GATEWAY := $(shell go list -m -f '{{.Dir}}' github.com/grpc-ecosystem/grpc-gateway/v2)

.PHONY: gen
gen: gen-tools proto

.PHONY: gen-tools
gen-tools:
	go mod download
	GOBIN="$(GOBIN)" go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	GOBIN="$(GOBIN)" go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
	GOBIN="$(GOBIN)" go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.28.0
	GOBIN="$(GOBIN)" go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.28.0


.PHONY: proto
proto: proto/v1/blog.proto
	@echo $(GRPC_GATEWAY)
	@mkdir -p gen/blog swagger
	protoc -I proto/v1 \
		-I $(GRPC_GATEWAY) \
		-I third_party/googleapis \
		proto/v1/blog.proto \
		--go_out=gen/blog --go_opt=paths=source_relative \
		--go-grpc_out=gen/blog --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=gen/blog --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=swagger --openapiv2_opt=logtostderr=true
