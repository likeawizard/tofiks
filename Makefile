GOAMD64VERSION=v3

build-tofiks:
	GOAMD64=${GOAMD64VERSION} go build -o tofiks cmd/tofiks/main.go

build-linux:
	GOAMD64=${GOAMD64VERSION} GOOS=linux go build -o tofiks cmd/tofiks/main.go

build-windows:
	GOAMD64=${GOAMD64VERSION} GOOS=windows go build -o tofiks.exe cmd/tofiks/main.go

clean:
	go clean

build: clean build-tofiks