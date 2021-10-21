.PHONY: format-proto
format-proto:
	find . -name "*.proto" -print0 | xargs -0 -I {} clang-format -i {}

.PHONY: check-proto
check-proto:
	find . -name "*.proto" -print0 | xargs -0 -I {} clang-format --dry-run --Werror {}

.PHONY: lint-proto
lint-proto:
	make check-proto
	buf lint proto
