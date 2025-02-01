BUILD_DIR = bin

build: cmd/printer/main.go
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/printer cmd/printer/main.go

clean:
	@rm -rf $(BUILD_DIR)

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	golangci-lint run

.PHONY: build clean test fmt lint
