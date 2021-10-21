.PHONY: run
run:
	go run -v -race cmd/main.go

.PHONY: format
format:
	@go install golang.org/x/tools/cmd/goimports@latest
	goimports -local github.com/sean-ahn -w .
	gofmt -s -w .

.PHONY: lint
lint:
	golangci-lint run ./...
