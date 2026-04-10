# Tofiks engine build and release targets — included from the root Makefile.
# All paths are relative to the project root since make is invoked from there.
# Variables (EXE, GOAMD64VERSION) are inherited from the root Makefile.

VERSION=v1.4.0
RELEASE_DIR=release

default:
	GOAMD64=${GOAMD64VERSION} go build -gcflags=-B -o $(EXE) cmd/tofiks/main.go

build-tofiks: default

build-linux:
	GOAMD64=${GOAMD64VERSION} GOOS=linux go build -o tofiks cmd/tofiks/main.go

build-windows:
	GOAMD64=${GOAMD64VERSION} GOOS=windows go build -o tofiks.exe cmd/tofiks/main.go

build-debug:
	GOAMD64=${GOAMD64VERSION} go build -tags debug -gcflags=-B -o debug_tofiks cmd/tofiks/main.go

build: clean build-tofiks

clean:
	go clean

release: clean
	@mkdir -p $(RELEASE_DIR)
	GOOS=linux GOARCH=amd64 GOAMD64=v3 go build -gcflags=-B -o $(RELEASE_DIR)/tofiks-$(VERSION)-linux-avx2 cmd/tofiks/main.go
	GOOS=linux GOARCH=amd64 GOAMD64=v4 go build -gcflags=-B -o $(RELEASE_DIR)/tofiks-$(VERSION)-linux-avx512 cmd/tofiks/main.go
	GOOS=windows GOARCH=amd64 GOAMD64=v3 go build -gcflags=-B -o $(RELEASE_DIR)/tofiks-$(VERSION)-windows-avx2.exe cmd/tofiks/main.go
	GOOS=windows GOARCH=amd64 GOAMD64=v4 go build -gcflags=-B -o $(RELEASE_DIR)/tofiks-$(VERSION)-windows-avx512.exe cmd/tofiks/main.go
	GOOS=darwin GOARCH=arm64 go build -gcflags=-B -o $(RELEASE_DIR)/tofiks-$(VERSION)-darwin-arm64 cmd/tofiks/main.go
