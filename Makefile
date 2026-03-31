BINARY := openaccess-sync
GO     := go

.PHONY: build test clean tidy vet

build:
	$(GO) build -o $(BINARY) .

test:
	$(GO) test ./... -v

clean:
	$(GO) clean
	rm -f $(BINARY)

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...
