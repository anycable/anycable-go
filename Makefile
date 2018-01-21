ifdef VERSION
else
	VERSION := $(shell sh -c 'git describe --always --tags')
endif

ifdef GOBIN
PATH := $(GOBIN):$(PATH)
else
PATH := $(subst :,/bin:,$(GOPATH))/bin:$(PATH)
endif

# Standard build
default: prepare build

# Install current version
install:
	go install ./...

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/anycable-go cmd/anycable-go/main.go

build-all:
	rm -rf ./dist
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-linux-arm" .
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-linux-arm64" .
	env CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-linux-386" .
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-linux-amd64" .
	env CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-win-386" .
	env CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-win-amd64" .
	env CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-macos-386" .
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-macos-amd64" .
	env CGO_ENABLED=0 GOOS=freebsd GOARCH=arm go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-freebsd-arm" .
	env CGO_ENABLED=0 GOOS=freebsd GOARCH=386 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-freebsd-386" .
	env CGO_ENABLED=0 GOOS=freebsd GOARCH=amd64 go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o "dist/anycable-go-$(VERSION)-freebsd-amd64" .

s3-deploy:
	aws s3 cp --acl=public-read ./dist/anycable-go-$(VERSION)-linux-386 "s3://anycable/builds/$(VERSION)/anycable-go-$(VERSION)-linux-386"

downloads-md:
	ruby etc/generate_downloads.rb

release: build-all s3-deploy downloads-md

docker-release: dockerize
	docker push "anycable/anycable-go:$(VERSION)"

dockerize:
	CGO_ENABLED=0 GOOS=linux go build -ldflags "-X main.version=$(VERSION)" -a -installsuffix cgo -o .docker/anycable-go .
	docker build -t "anycable/anycable-go:$(VERSION)" .

# Run server
run:
	go run ./*.go

build-protos:
	protoc --proto_path=./etc --go_out=plugins=grpc:./protos ./etc/rpc.proto

test:
	go test .

test-cable:
	go build -o tmp/anycable-go-test .
	anyt -c "tmp/anycable-go-test -headers=cookie,x-api-token" --target-url="ws://localhost:8080/cable"
	anyt -c "tmp/anycable-go-test -headers=cookie,x-api-token -ssl_key=etc/ssl/server.key -ssl_cert=etc/ssl/server.crt -addr=localhost:8443" --target-url="wss://localhost:8443/cable"

test-ci: prepare test test-cable

# Get dependencies and use gdm to checkout changesets
prepare:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

gen-ssl:
	mkdir -p tmp/ssl
	openssl genrsa -out tmp/ssl/server.key 2048
	openssl req -new -x509 -sha256 -key tmp/ssl/server.key -out tmp/ssl/server.crt -days 3650

vet:
	go vet ./...

fmt:
	go fmt ./...
