# go commands
GO=go
GOBUILD=$(GO) build
GOCLEAN=$(GO) clean
GOTEST=$(GO) test

# binary names
DATASETCREATOR_BINARY=./bin/datasetcreator.exe
LANGUAGESERVER_BINARY=./bin/languageserver.exe

.PHONY: languageserver

all: build

build: datasetcreator languageserver

datasetcreator:
	$(GOBUILD) -o $(DATASETCREATOR_BINARY) ./cmd/datasetcreator

languageserver:
	$(GOBUILD) -o $(LANGUAGESERVER_BINARY) ./cmd/languageserver

clean:
	$(GOCLEAN)
	rm -f $(DATASETCREATOR_BINARY)
	rm -f $(LANGUAGESERVER_BINARY)