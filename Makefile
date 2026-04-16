EXE = tofiks
GOAMD64VERSION=v3
FASTCHESS=$(HOME)/fastchess/fastchess
OPENINGS=$(HOME)/fastchess/UHO_Lichess_4852_v1.epd
TOFIKS_PROD=$(HOME)/tofiks/tofiks

# Domain-specific Makefiles. cmd/tofiks/Makefile must be included first so its
# `default` target remains the make-default when no target is specified.
include cmd/tofiks/Makefile
include test_suite/Makefile
include cmd/texel/Makefile

cutechess:
	@-rm games.pgn
	cutechess-cli -engine conf=tofiks -engine conf=tofiks-1.2 -each proto=uci tc=1+0.1 timemargin=50 -games 2 -rounds 5000 -concurrency 4 -repeat -openings file=${OPENINGS} format=epd -pgnout games.pgn -recover

test-cutechess: build
	cutechess-cli -engine conf=tofiks -engine conf=tofiksProd -each proto=uci tc=0.5+0.05 timemargin=50 -games 2 -rounds 2000 -concurrency 7 -repeat -sprt elo0=0 elo1=5 alpha=0.05 beta=0.05 -openings file=${OPENINGS} format=epd -recover

pgo-cutechess: build
	cutechess-cli -engine conf=tofiks arg=-pgo -engine conf=tofiksProd -each proto=uci tc=2+0.2 timemargin=50 -rounds 1 -openings file=${OPENINGS} format=epd -recover

memprof-cutechess: build
	cutechess-cli -engine conf=tofiks arg=-memprof -engine conf=tofiksProd -each proto=uci tc=30+1 timemargin=50 -rounds 1 -openings file=${OPENINGS} format=epd -recover

fastchess: build
	@-rm games.pgn
	${FASTCHESS} -engine cmd=./tofiks name=tofiks-dev -engine cmd=${TOFIKS_PROD} name=tofiks-prod -each proto=uci tc=1+0.1 timemargin=50 -rounds 5000 -concurrency 4 -repeat -openings file=${OPENINGS} format=epd order=random -pgnout file=games.pgn -recover

test-fastchess-ltc: build
	${FASTCHESS} -engine cmd=./tofiks name=tofiks-dev -engine cmd=${TOFIKS_PROD} name=tofiks-prod -each proto=uci tc=60+0.6 timemargin=50 -rounds 2000 -concurrency 7 -repeat -sprt elo0=-5 elo1=15 alpha=0.1 beta=0.1 -openings file=${OPENINGS} format=epd order=random -recover

test-fastchess-stc: build
	${FASTCHESS} -engine cmd=./tofiks name=tofiks-dev -engine cmd=${TOFIKS_PROD} name=tofiks-prod -each proto=uci tc=10+0 timemargin=50 -rounds 2000 -concurrency 7 -repeat -sprt elo0=-5 elo1=15 alpha=0.1 beta=0.1 -openings file=${OPENINGS} format=epd order=random -recover

test-fastchess-vstc: build
	${FASTCHESS} -engine cmd=./tofiks name=tofiks-dev -engine cmd=${TOFIKS_PROD} name=tofiks-prod -each proto=uci tc=0.5+0.05 timemargin=50 -rounds 2000 -concurrency 7 -repeat -sprt elo0=-5 elo1=15 alpha=0.1 beta=0.1 -openings file=${OPENINGS} format=epd order=random -recover

test-fastchess-vvstc: build
	${FASTCHESS} -engine cmd=./tofiks name=tofiks-dev -engine cmd=${TOFIKS_PROD} name=tofiks-prod -each proto=uci tc=0.2+0.02 timemargin=50 -rounds 2000 -concurrency 7 -repeat -sprt elo0=-5 elo1=15 alpha=0.1 beta=0.1 -openings file=${OPENINGS} format=epd order=random -recover

pgo-fastchess: build
	${FASTCHESS} -engine cmd=./tofiks args=-pgo name=tofiks-dev -engine cmd=${TOFIKS_PROD} name=tofiks-prod -each proto=uci tc=2+0.2 timemargin=50 -rounds 1 -openings file=${OPENINGS} format=epd order=random -recover

memprof-fastchess: build
	${FASTCHESS} -engine cmd=./tofiks args=-memprof name=tofiks-dev -engine cmd=${TOFIKS_PROD} name=tofiks-prod -each proto=uci tc=30+1 timemargin=50 -rounds 1 -openings file=${OPENINGS} format=epd order=random -recover

spsa-init:
	@echo '{"params":[{"name":"ExampleParam","type":"int","value":100,"min":50,"max":150,"c_end":10,"r_end":0.002}],"spsa_alpha":0.602,"spsa_gamma":0.101,"spsa_A_ratio":0.1,"spsa_iterations":10000,"spsa_pairs_per":32,"spsa_reporting_type":"BULK","spsa_distribution_type":"SINGLE"}' | jq . > spsa.json
	@echo "Created spsa.json template — replace ExampleParam with your params"

lint:
	go tool golangci-lint run

lint-fix:
	go tool golangci-lint run --fix
