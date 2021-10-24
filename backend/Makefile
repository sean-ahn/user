.PHONY: run
run:
	go run -v -race cmd/main.go

.PHONY: format
format:
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/incu6us/goimports-reviser/v2@latest
	find . -name "*.go" -print0 | xargs -0 -I {} goimports-reviser -rm-unused -project-name github.com/sean-ahn/user -file-path {}
	gofmt -s -w .

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: generate-mock
generate-mock:
	@go get github.com/golang/mock/gomock
	@go install github.com/golang/mock/mockgen@v1.6.0
	go generate ./...

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
