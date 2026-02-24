.PHONY: build clean test install lint vet

VERSION ?= dev
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o phorge ./cmd/phorge

test:
	go test ./... -v

vet:
	go vet ./...

clean:
	rm -f phorge phorge.exe

install:
	go install $(LDFLAGS) ./cmd/phorge

lint: vet
	@echo "Lint complete"
