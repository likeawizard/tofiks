GOAMD64VERSION=v3

build-perft:
	GOAMD64=${GOAMD64VERSION} go build -o perft cmd/perft/main.go

build-analyze:
	GOAMD64=${GOAMD64VERSION} go build -o analyze cmd/analyze/main.go

build-lichess:
	GOAMD64=${GOAMD64VERSION} go build -o lichess cmd/lichess/main.go

build-self-play:
	GOAMD64=${GOAMD64VERSION} go build -o self-play cmd/self-play/main.go

build-single-eval:
	GOAMD64=${GOAMD64VERSION} go build -o single-eval cmd/single-eval/main.go

clean:
	go clean
	rm perft
	rm analyze
	rm lichess
	rm self-play
	rm single-eval

build: build-perft build-analyze build-lichess build-self-play build-single-eval

clean-build: clean build