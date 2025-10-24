# Aegis Protocol Makefile
PROG_CLIENT=bin/aegis-mcp-client
PROG_SERVER=bin/aegis-mcp-server
PROG_TERMINAL=bin/aegis-terminal
PROG_API=bin/aegis-api
SRCS_CLIENT=./cmd/aegis-mcp-client
SRCS_SERVER=./cmd/aegis-mcp-server
SRCS_TERMINAL=./cmd/aegis-terminal
SRCS_API=./cmd/aegis-api

# Version info
COMMIT_HASH=$(shell git rev-parse --short HEAD || echo "GitNotFound")
BUILD_DATE=$(shell date -u '+%Y-%m-%d %H:%M:%S')
BUILD_FLAGS=-ldflags "-s -w -X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\""

# Default target
all: build-client build-server build-terminal build-api

# Create bin directory
$(shell mkdir -p bin)

# Build targets
build-client:
	go build ${BUILD_FLAGS} -o ${PROG_CLIENT} ${SRCS_CLIENT}

build-server:
	go build ${BUILD_FLAGS} -o ${PROG_SERVER} ${SRCS_SERVER}

build-terminal:
	go build ${BUILD_FLAGS} -o ${PROG_TERMINAL} ${SRCS_TERMINAL}

build-api:
	go build ${BUILD_FLAGS} -o ${PROG_API} ${SRCS_API}

# Cross-compilation targets
build-all: build-linux build-windows build-darwin build-arm

build-linux:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-linux ${SRCS_CLIENT}
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-linux ${SRCS_SERVER}
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-linux ${SRCS_TERMINAL}
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_API}-linux ${SRCS_API}

build-windows:
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-windows.exe ${SRCS_CLIENT}
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-windows.exe ${SRCS_SERVER}
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-windows.exe ${SRCS_TERMINAL}
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_API}-windows.exe ${SRCS_API}

build-darwin:
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-darwin ${SRCS_CLIENT}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-darwin ${SRCS_SERVER}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-darwin ${SRCS_TERMINAL}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_API}-darwin ${SRCS_API}

build-arm:
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-arm ${SRCS_CLIENT}
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-arm ${SRCS_SERVER}
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-arm ${SRCS_TERMINAL}
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_API}-arm ${SRCS_API}

# Development targets
run-client:
	go run ${SRCS_CLIENT}

run-server:
	go run ${SRCS_SERVER}

run-terminal:
	go run ${SRCS_TERMINAL}

run-api:
	go run ${SRCS_API}

# Cleanup
clean:
	rm -f ${PROG_CLIENT}* ${PROG_SERVER}* ${PROG_TERMINAL}* ${PROG_API}* bin/*.exe

.PHONY: all build-client build-server build-terminal build-api build-all build-linux build-windows build-darwin build-arm run-client run-server run-terminal run-api clean
