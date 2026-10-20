MODULE := github.com/RovierrHQ/envy
BIN    := bin
APP    := envy

.PHONY: all build test tidy cross clean

all: build

build:
	go build -o $(BIN)/$(APP)$(shell go env GOEXE) ./cmd/envy

test:
	go test ./...

tidy:
	go mod tidy

# Cross-compile single native binaries for release targets.
cross:
	mkdir -p $(BIN)
	GOOS=windows GOARCH=amd64 go build -o $(BIN)/$(APP)-windows-amd64.exe ./cmd/envy
	GOOS=darwin  GOARCH=amd64 go build -o $(BIN)/$(APP)-darwin-amd64 ./cmd/envy
	GOOS=darwin  GOARCH=arm64 go build -o $(BIN)/$(APP)-darwin-arm64 ./cmd/envy
	GOOS=linux   GOARCH=amd64 go build -o $(BIN)/$(APP)-linux-amd64 ./cmd/envy

clean:
	rm -rf $(BIN)
