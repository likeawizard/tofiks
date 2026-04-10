# Texel tuning targets — included from the root Makefile.
# All paths are relative to the project root since make is invoked from there.
# Variables (GOAMD64VERSION, FASTCHESS, TOFIKS_PROD) are inherited from the root Makefile.

remove-dup:
	@echo "Removing duplicates from texel_data.txt"
	@gawk -i inplace '!seen[$$0]++' texel_data.txt
	@echo "Randomizing position order"
	@shuf -o texel_data.txt texel_data.txt
	@rm -f texel_data.bin
	@echo "Removed cache (texel_data.bin) — will rebuild on next tune"

run-texel:
	GOAMD64=${GOAMD64VERSION} go build -o texel cmd/texel/main.go
	./texel -f texel_data.txt -i 1000

texel-data:
	${FASTCHESS} \
	-engine cmd=${TOFIKS_PROD} name=tofiks1 \
	-engine cmd=${TOFIKS_PROD} name=tofiks2 \
	-openings file=$(HOME)/Downloads/UHO_Lichess_4852_v1.epd format=epd order=random \
	-each tc=0.2+0.02 option.Hash=64 \
	-rounds 15000 -repeat \
	-concurrency 6 \
	-draw movenumber=40 movecount=8 score=10 \
	-resign movecount=3 score=600 \
	-pgnout file=selfplay.pgn notation=uci min=true \
	-recover
