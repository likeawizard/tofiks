GOAMD64VERSION=v3

build-tofiks:
	GOAMD64=${GOAMD64VERSION} go build -o tofiks cmd/tofiks/main.go

clean:
	go clean
	rm tofiks

build: build-tofiks

clean-build: clean build