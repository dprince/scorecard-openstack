.PHONY: all
all: build

.PHONY: build
build: fmt vet ## Build manager binary.
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o scorecard-openstack main.go

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy
