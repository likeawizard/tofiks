package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/likeawizard/tofiks/pkg/board"
)

func main() {
	var pgnPath, outPath string
	var skipPlies int
	flag.StringVar(&pgnPath, "pgn", "selfplay.pgn", "PGN input file")
	flag.StringVar(&outPath, "o", "texel_data.txt", "Output file")
	flag.IntVar(&skipPlies, "skip", 8, "Skip first N plies (opening book moves)")
	flag.Parse()

	games, err := parsePGN(pgnPath)
	if err != nil {
		log.Fatalf("Failed to parse PGN: %v", err)
	}
	log.Printf("Parsed %d games", len(games))

	type fenResult struct {
		fen    string
		result string
	}

	fenCh := make(chan fenResult, 1000)
	var wg sync.WaitGroup

	for _, g := range games {
		wg.Add(1)
		go func(g game) {
			defer wg.Done()
			result := ""
			switch g.result {
			case "1-0":
				result = "1"
			case "0-1":
				result = "0"
			case "1/2-1/2":
				result = "0.5"
			default:
				return
			}

			b := board.NewBoard(g.startFEN)
			for i, uci := range g.moves {
				if i >= skipPlies && !b.InCheck {
					fenCh <- fenResult{fen: b.ExportFEN(), result: result}
				}
				_, ok := b.MoveUCI(uci)
				if !ok {
					break
				}
			}
		}(g)
	}

	go func() {
		wg.Wait()
		close(fenCh)
	}()

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("Failed to create output: %v", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	count := 0
	for fr := range fenCh {
		fmt.Fprintf(w, "%s %s\n", fr.result, fr.fen)
		count++
	}
	w.Flush()
	log.Printf("Wrote %d positions to %s", count, outPath)
}

type game struct {
	startFEN string
	result   string
	moves    []string
}

func parsePGN(path string) ([]game, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var games []game
	var headers map[string]string
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 1024*1024), 1024*1024)

	headers = make(map[string]string)
	var moveLines []string

	flush := func() {
		if len(moveLines) == 0 {
			return
		}
		allMoves := strings.Join(moveLines, " ")
		// Remove move numbers like "7." "12."
		var moves []string
		for tok := range strings.FieldsSeq(allMoves) {
			// Skip move numbers, results, and comments.
			if strings.HasSuffix(tok, ".") {
				continue
			}
			if tok == "1-0" || tok == "0-1" || tok == "1/2-1/2" || tok == "*" {
				continue
			}
			moves = append(moves, tok)
		}

		startFEN := "startpos"
		if fen, ok := headers["FEN"]; ok {
			startFEN = fen
		}

		games = append(games, game{
			startFEN: startFEN,
			result:   headers["Result"],
			moves:    moves,
		})
		headers = make(map[string]string)
		moveLines = nil
	}

	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if len(moveLines) > 0 {
				flush()
			}
			// Parse header.
			line = line[1 : len(line)-1]
			parts := strings.SplitN(line, " ", 2)
			if len(parts) == 2 {
				val := strings.Trim(parts[1], "\"")
				headers[parts[0]] = val
			}
		} else if strings.TrimSpace(line) != "" {
			moveLines = append(moveLines, line)
		}
	}
	flush()

	return games, nil
}
