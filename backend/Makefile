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

.PHONY: generate-sql
generate-sql:
	@go install github.com/volatiletech/sqlboiler/v4@latest
	@go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-mysql@latest
	cat sqlboiler.tmpl.toml | gomplate > sqlboiler.toml
	sqlboiler --wipe --no-tests --no-auto-timestamps -p model -o ./model mysql
	@$(MAKE) format

.PHONY: test
test:
	@go install github.com/rakyll/gotest@latest
	gotest -p 1 -race -cover -v ./...
