CONFIG_PATH=${HOME}/.boom/
NAME := boom
GO := go
ROOT_PACKAGE := $(GIT_PROVIDER)/joostvdg/$(NAME)
GO_VERSION := $(shell $(GO) version | sed -e 's/^[^0-9.]*\([0-9.]*\).*/\1/')
PACKAGE_DIRS := $(shell $(GO) list ./... | grep -v /vendor/)
PKGS := $(shell go list ./... | grep -v /vendor | grep -v generated)
CGO_ENABLED = 0


.PHONY: init
init:
	mkdir -p ${CONFIG_PATH}

.PHONY: gencert
gencert:
	cfssl gencert \
		-initca test/ca-csr.json | cfssljson -bare ca

	cfssl gencert \
		-ca=ca.pem \
		-ca-key=ca-key.pem \
		-config=test/ca-config.json \
		-profile=server \
		test/server-csr.json | cfssljson -bare server

		cfssl gencert \
			-ca=ca.pem \
			-ca-key=ca-key.pem \
			-config=test/ca-config.json \
			-profile=client \
			-cn="root" \
			test/client-csr.json | cfssljson -bare root-client

		cfssl gencert \
			-ca=ca.pem \
			-ca-key=ca-key.pem \
			-config=test/ca-config.json \
			-profile=client \
			-cn="nobody" \
			test/client-csr.json | cfssljson -bare nobody-client

	mv *.pem *.csr ${CONFIG_PATH}

.PHONY: linux
linux:
	CGO_ENABLE=0 GOOS=linux GOARCH=amd64 go build -o bin/boom ./cmd/boom

.PHONY: server
server:
	CGO_ENABLE=0 go build -o bin/boom ./cmd/boom-server

.PHONY: client
client:
	CGO_ENABLE=0 go build -o bin/boom-client ./cmd/boom-client

test:
	CGO_ENABLED=$(CGO_ENABLED) $(GO) test $(PACKAGE_DIRS) -test.v -coverprofile cp.out

fmt:
	@gofmt -s -w -l **/*.go