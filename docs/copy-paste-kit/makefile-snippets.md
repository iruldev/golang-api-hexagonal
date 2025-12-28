# Makefile Snippets

Add these to your project's `Makefile`:

```makefile
# Testing targets
.PHONY: test test-shuffle test-integration test-all gencheck

test:
	go test -v -race -count=1 ./...

test-shuffle:
	go test -v -race -count=1 -shuffle=on ./...

test-integration:
	go test -v -race -count=1 -tags=integration ./...

test-all: test test-integration

# Generation check
gencheck:
	go generate ./...
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Generated files are out of date"; \
		git diff; \
		exit 1; \
	fi
```
