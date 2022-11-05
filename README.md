# Chess engine written in Golang
## Installation
* Clone repository or copy files
* Make sure you have Golang 1.17+ installed
* On linux systems simply run `./build.sh` or manually compile each of the endpoints under `./cmd/**`
* You should have three separate executables `analyze, self-play, lichess, single-eval`
* `cp config.yml.dev config.yml`

## Setting up .env
* For `./self-play` you can set `init.startingFen` to set a custom starting position. (Comment out to use standard starting position)
* Set search depth `engine.depth` number of plies (half-moves) to look into. For example `depth=2` will look 2 plies deep- next move for the side to play, response from opposition.
* Set `engine.algorithm=alphabete(default)|minmax` to set what search algorithm will be used.
* For `./lichess` only. Set `lichess.apiToken` to your bot account token. **Do not publish this token as it is equivalent to your password.** To learn more about how to set up a lichess bot account see: [Upgrading To a Lichess Bot Account](https://lichess.org/api#operation/botAccountUpgrade)


## Usage
### Self Play
* Set up your `config.yml` file
* Run `./self-play`
* Sending an interrupt signal `Ctrl+c` will also output the full movelist of the game before shutdown
### Analyze
* Run `./analyze -fen="<position to analyze>"`
### Single Eval
* This executable is for development and debug purposes to check the evaluation function of a single position
* It will output the evaluation of the position and values for all pieces that contribute to it
* Run `./single-eval -fen="<position to analyze>"`
### Lichess
* Run `./lichess`.
* The bot will authenticate using the token from `config.yml` and listen for incomming challenges and moves