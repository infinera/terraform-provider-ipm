TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=infinera.com
NAMESPACE=poc
NAME=ipm
BINARY=terraform-provider-${NAME}
VERSION=`git describe --tags --abbrev=0`
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

registry:
    mkdir -p assets
	go build -o ${BINARY}_v${VERSION}
	zip ${BINARY}_${VERSION}_${OS_ARCH}.zip ${BINARY}_v${VERSION}
	shasum -a 256 *.zip > ${BINARY}_${VERSION}_SHA256SUMS
	gpg --detach-sign ${BINARY}_${VERSION}_SHA256SUMS

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

build/docker:
	rm -rf terraform-ipm-modules
	git clone ssh://git@bitbucket.infinera.com:7999/mar/terraform-ipm-modules.git
	docker build -t sv-artifactory.infinera.com/marvel/ipm/ipm-services:1.0.0 .
	rm -rf terraform-ipm-modules

default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
