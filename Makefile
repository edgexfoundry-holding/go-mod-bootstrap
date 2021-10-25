.PHONY: test

GO=CGO_ENABLED=0 GO111MODULE=on go

tidy:
	go mod tidy

test:
	$(GO) test ./... -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")
	[ "`gofmt -l $$(find . -type f -name '*.go'| grep -v "/vendor/")`" = "" ]

vendor:
	go mod vendor
