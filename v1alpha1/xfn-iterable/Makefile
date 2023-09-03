.PHONY: test
test: go-mod-tidy fmt
	go test ./... -coverprofile cover.out

.PHONY: fmt
fmt:
	go fmt ./...

LOCAL = $(shell head -n 1 go.mod | awk '{print $$2}')
FILES = $(shell find . -type f -name '*.go' -not -name 'zz_generated*' -not -path '*/.*' -not -path '*/_*')
.PHONY: fmt-imports
fmt-imports:
	goimports -w --local ${LOCAL} ${FILES}

.PHONY: go-mod-tidy
go-mod-tidy:
	go mod tidy
