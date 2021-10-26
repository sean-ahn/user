.PHONY: format
format:
	find . -name "*.proto" -print0 | xargs -0 -I {} clang-format -i {}

.PHONY: check
check:
	find . -name "*.proto" -print0 | xargs -0 -I {} clang-format --dry-run --Werror {}

.PHONY: lint
lint:
	@$(MAKE) check
	buf lint

.PHONY: generate
generate:
	buf generate

.PHONY: clean
clean:
	rm -rf gen/
