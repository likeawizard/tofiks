# Go test and benchmark targets — included from the root Makefile.
# All paths are relative to the project root since make is invoked from there.

test-suite:
	go test -v -timeout 0 ./test_suite

test-short:
	go test -short -v -timeout 0 ./test_suite

test-pv:
	go test -run=TestValidPV -v -timeout 0 ./test_suite/

test-perft:
	go test -run=TestPerft -v -timeout 0 ./test_suite/

test-mate:
	go test -run=TestMate -v -timeout 0 ./test_suite/

fuzz-entry:
	go test -run=- -fuzz=FuzzEntry -v ./test_suite/

bench-search:
	go test -run=^$$ -bench='Benchmark(PVS|Quiescence|IDSearch)' -benchmem ./test_suite/

run-bench:
	go test -bench=. -count=20 -benchmem -cpu=1,2,4,12 ./test_suite/ > new.bench
