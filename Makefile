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

test-suite:
	go test -v -timeout 0 ./test_suite

test-pv:
	go test -run=TestValidPV -v -timeout 0 ./test_suite/

test-perft:
	go test -run=TestPerft -v -timeout 0 ./test_suite/

test-mate:
	go test -run=TestMate -v -timeout 0 ./test_suite/

test-entry:
	go test -fuzz=FuzzEntry -v -timeout 0 ./test_suite/

run-bench:
	go test -run=BenchmarkMake -bench=. -benchtime=10s -benchmem -cpu=1,2,4,12 ./test_suite/

cutechess:
	@-rm games.pgn
	cutechess-cli -engine conf=tofiks -engine conf=tofiks-1.2 -each proto=uci tc=1+0.1 -games 2 -rounds 5000 -concurrency 4 -repeat -openings file=/home/arturs/cutechess/Arasan.pgn format=pgn plies=10 -pgnout games.pgn -recover

remove-dup:
	@echo "Removing duplicates"
	@gawk -i inplace '!seen[$0]++' texel_data.txt
	@echo "Randomizing position order"
	@shuf -o texel_data.txt texel_data.txt

make run-texel:
	GOAMD64=${GOAMD64VERSION} go build -o texel cmd/texel/main.go
	./texel -f "rand.txt" -c 8 -i 100 -lim 100000 > texel_out.txt