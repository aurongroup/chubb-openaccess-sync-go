CSVEXPORT_BINARY    	:= cmd/csvexport/csvexport
FULLEXPORT_BINARY		:= cmd/fullexport/fullexport
SYNC_BINARY				:= cmd/sync/sync
MIGRATE_BINARY			:= cmd/migrate/migrate
TESTHARNESS_BINARY		:= cmd/testharness/testharness
GO        				:= go
LDFLAGS   				:= -ldflags="-s -w"

.PHONY: build build-debug build-windows build-windows-debug test clean tidy vet

build: build-csvexport build-fullexport build-sync build-migrate build-testharness

build-csvexport:
	$(GO) build $(LDFLAGS) -o $(CSVEXPORT_BINARY) cmd/csvexport/main.go

clean-csvexport:
	rm -f $(CSVEXPORT_BINARY)

build-fullexport:
	$(GO) build $(LDFLAGS) -o $(FULLEXPORT_BINARY) cmd/fullexport/main.go

clean-fullexport:
	rm -f $(FULLEXPORT_BINARY)

build-sync:
	$(GO) build $(LDFLAGS) -o $(SYNC_BINARY) cmd/sync/main.go

clean-sync:
	rm -f $(SYNC_BINARY)

build-migrate:
	$(GO) build $(LDFLAGS) -o $(MIGRATE_BINARY) cmd/migrate/main.go

clean-migrate:
	rm -f $(MIGRATE_BINARY)

build-testharness:
	$(GO) build $(LDFLAGS) -o $(TESTHARNESS_BINARY) cmd/testharness/main.go

clean-testharness:
	rm -f $(TESTHARNESS_BINARY)

build-debug:
	$(GO) build -o $(BINARY) .

build-windows:
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BINARY).exe .

build-windows-debug:
	GOOS=windows GOARCH=amd64 $(GO) build -o $(BINARY).exe .

test:
	$(GO) test ./... -v

clean: clean-csvexport clean-fullexport clean-sync clean-migrate clean-testharness

tidy:
	$(GO) mod tidy

vet:
	$(GO) vet ./...
