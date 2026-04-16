# Tofiks UCI chess engine - Yes, I am a dog and, yes, I play chess. Get over it!
![Tofiks](resources/tofiks_logo.jpeg)
## Installation

[![Latest Release](https://img.shields.io/github/v/release/likeawizard/tofiks)](https://github.com/likeawizard/tofiks/releases/latest)

Pre-built 64-bit binaries are available on the [releases page](https://github.com/likeawizard/tofiks/releases/latest).
### Compile yourself
* Clone repository or copy files
* Make sure you have Go 1.26+ installed

**Using Make (recommended):**
* `make build` will compile for your native architecture.
* In the Makefile you can change the `GOAMD64VERSION` variable. The values range from v1 to v4 and include various sets of extended instructions. A higher v-value should be expected to have better performance but might have more specific architecture and instruction set extension requirements. Most modern CPUs should be AVX2 compatible which is v3, v4 requires AVX512 support.

**Using go build directly:**
```
go build -o tofiks cmd/tofiks/main.go
```

**Profile-Guided Optimization (PGO):** The repository includes `cmd/tofiks/default.pgo`, a CPU profile. Go 1.21+ automatically uses a `default.pgo` file in the build directory for profile-guided optimization.

## Feature Overview

### Search
* Principal Variation Search with aspiration windows
* Null Move pruning
* Late Move Reductions / Late Move Pruning
* Singular Extensions
* Transposition table with aging
* Quiescence search
* MVV-LVA move ordering
* History Heuristic
* Killer Move Heuristic
* Counter Move Heuristic

### Evaluation
* Tapered eval (middlegame / endgame phase interpolation)
* Piece-Square Tables (texel-tuned)
* Piece mobility and capture bonuses
* King safety (center distance, pawn shield) and endgame king activity
* Pawn structure (doubled, isolated, backward, blocked, connected/candidate passers)
* Passed pawn bonuses by rank with king-proximity scaling in endgame
* Outpost evaluation for knights and bishops
* Bishop pair bonus
* Rook on open/semi-open files
* Kaufman piece-value adjustments

### Other
* PolyGlot opening book support
* Texel tuner with streaming Adam optimizer
* Supported UCI commands and options:
   * `uci` — engine responds with id and supported options
   * `go` — wtime, btime, winc, binc, movestogo, depth, movetime, ponder, infinite
   * `setoption name <option> value <value>`
       * Ponder (default false)
       * OwnBook (default false) — if a PolyGlot `book.bin` is in the same directory as the executable the engine will load it
       * Hash — Transposition Table size in MB
       * Move Overhead — lag compensation (negative increment to reduce allotted thinking time)
* Non-UCI commands:
    * `go perft <depth>` — leaf node count grouped by legal moves, with NPS for performance benchmarking

## Acknowledgments
* [Lichess](https://lichess.org/) and the Community for nurturing my love for chess and offering a free open source platform for chess. Tofiks can be found playing on lichess under its [bot account](https://lichess.org/@/likeawizard-bot).
* The authors of [Official lichess bot client](https://github.com/ShailChoksi/lichess-bot). It saved a lot of time and headaches adopting a reliable solution to interact with the Lichess API instead of coding all myself.
* The authors of [Blunder engine](https://github.com/algerbrex/blunder). It is a great reference to practical examples of engine principles and algorithms with its well documented code.
* Everyone on [TalkChess.com Forums](https://talkchess.com/forum3/) who have been helpful and patient explaining even very basic concepts to me.
* [Chess Programming Wiki](https://www.chessprogramming.org/Main_Page) and its maintainers for creating a very comprehensive resource for everything chess programming related.