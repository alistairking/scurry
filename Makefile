GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINDIR=./bin
CLI=scurry

all: cli

dev: mod-tidy codegen pkg cli test

cli:
	mkdir -p $(BINDIR)
	$(GOBUILD) -o $(BINDIR)/$(CLI) -v ./cmd/scurry

pkg:
	$(GOBUILD) ./

test:
	$(GOTEST) -v ./...

clean:
	$(GOCLEAN)
	rm -f $(BINDIR)

mod-tidy:
	$(GOMOD) tidy

codegen:
	go get -d github.com/alvaroloes/enumer
	go generate ./

run: cli
	$(BINDIR)/$(CLI)
