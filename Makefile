GO_BIN?=$(shell pwd)/.bin
GOCI_LINT_VERSION?=v1.64.5

SHELL:=env PATH=$(GO_BIN):$(PATH) $(SHELL)

# Format the code
format::
	golangci-lint run --fix -v ./...

# Generate the go code from the proto files
generate-proto::
	go tool buf generate --template buf.gen.yaml

# Detech any breaking change in proto
lint-breaking::
	go tool buf breaking --against 'https://github.com/codeandlearn1991/news-grpc.git#branch=main'

lint-go::
	golangci-lint run -v ./...

# Lint the proto files
lint-proto::
	go tool buf lint --config buf.yaml

# Run all linters
lint:: lint-breaking lint-go lint-proto

# Install tool in local bin
install-tools::
	mkdir -p ${GO_BIN}
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ${GO_BIN} ${GOCI_LINT_VERSION}
	go install tool

# Run tidy
tidy::
	go mod tidy -v
