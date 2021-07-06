test-short:
	@echo "Testing Go packages..."
	@go test ./... -cover -count=1 -short

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run --timeout 3m
