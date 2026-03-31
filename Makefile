BINARY    := openaccess-sync
GO        := go
LDFLAGS   := -ldflags="-s -w"

.PHONY: build build-debug build-windows build-windows-debug test clean tidy vet

build:
	$(GO) build $(LDFLAGS) -o $(BINARY) .

build-debug:
	$(GO) build -o $(BINARY) .

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY).exe .

build-windows-debug:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY).exe .

test:
	$(GO) test ./... -v

clean:
	$(GO) clean
	rm -f $(BINARY) $(BINARY).exe

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...
