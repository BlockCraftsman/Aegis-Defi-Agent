# Aegis Protocol Makefile
PROG_CLIENT=bin/aegis-mcp-client
PROG_SERVER=bin/aegis-mcp-server
PROG_TERMINAL=bin/aegis-terminal
SRCS_CLIENT=./cmd/aegis-mcp-client
SRCS_SERVER=./cmd/aegis-mcp-server
SRCS_TERMINAL=./cmd/aegis-terminal

# Version info
COMMIT_HASH=$(shell git rev-parse --short HEAD || echo "GitNotFound")
BUILD_DATE=$(shell date -u '+%Y-%m-%d %H:%M:%S')
BUILD_FLAGS=-ldflags "-s -w -X \"main.BuildVersion=${COMMIT_HASH}\" -X \"main.BuildDate=${BUILD_DATE}\""

# Default target
all: build-client build-server build-terminal

# Create bin directory
$(shell mkdir -p bin)

# Build targets
build-client:
	go build ${BUILD_FLAGS} -o ${PROG_CLIENT} ${SRCS_CLIENT}

build-server:
	go build ${BUILD_FLAGS} -o ${PROG_SERVER} ${SRCS_SERVER}

build-terminal:
	go build ${BUILD_FLAGS} -o ${PROG_TERMINAL} ${SRCS_TERMINAL}

# Cross-compilation targets
build-all: build-linux build-windows build-darwin build-arm

build-linux:
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-linux ${SRCS_CLIENT}
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-linux ${SRCS_SERVER}
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-linux ${SRCS_TERMINAL}

build-windows:
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-windows.exe ${SRCS_CLIENT}
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-windows.exe ${SRCS_SERVER}
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-windows.exe ${SRCS_TERMINAL}

build-darwin:
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-darwin ${SRCS_CLIENT}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-darwin ${SRCS_SERVER}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-darwin ${SRCS_TERMINAL}

build-arm:
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_CLIENT}-arm ${SRCS_CLIENT}
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_SERVER}-arm ${SRCS_SERVER}
	GOOS=linux GOARCH=arm64 go build ${BUILD_FLAGS} -o ${PROG_TERMINAL}-arm ${SRCS_TERMINAL}

# Development targets
run-client:
	go run ${SRCS_CLIENT}

run-server:
	go run ${SRCS_SERVER}

run-terminal:
	go run ${SRCS_TERMINAL}

# Cleanup
clean:
	rm -f ${PROG_CLIENT}* ${PROG_SERVER}* ${PROG_TERMINAL}* bin/*.exe

.PHONY: all build-client build-server build-terminal build-all build-linux build-windows build-darwin build-arm run-client run-server run-terminal clean
