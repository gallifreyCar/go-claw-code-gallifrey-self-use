.PHONY: all build install clean test lint deps

# 变量
BINARY_NAME=go-claw-code
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DIR=bin
GO=go
GOFLAGS=-v

# 构建信息
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

all: deps build

## deps: 安装依赖
deps:
	$(GO) mod download
	$(GO) mod tidy

## build: 构建二进制文件
build:
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

## install: 安装到 $GOPATH/bin
install:
	$(GO) install $(LDFLAGS) ./cmd/$(BINARY_NAME)

## clean: 清理构建产物
clean:
	rm -rf $(BUILD_DIR)
	$(GO) clean

## test: 运行测试
test:
	$(GO) test -v -race -coverprofile=coverage.out ./...

## coverage: 查看测试覆盖率
coverage: test
	$(GO) tool cover -html=coverage.out

## lint: 代码检查
lint:
	@which golangci-lint > /dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

## fmt: 格式化代码
fmt:
	$(GO) fmt ./...
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .

## run: 本地运行
run:
	$(GO) run ./cmd/$(BINARY_NAME)

## docker-build: 构建 Docker 镜像
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

## help: 显示帮助
help:
	@echo "可用命令:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'