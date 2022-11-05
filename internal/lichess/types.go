package lichess

const (
	EVENT_CHALLENGE  = "challenge"
	EVENT_GAME_START = "gameStart"
	EVENT_GAME_END   = "gameEnd"
	GAME_EVENT_FULL  = "gameFull"
	GAME_EVENT_STATE = "gameState"
)

type StreamEvent struct {
	Type      string    `json:"type"`
	Challenge Challenge `json:"challenge"`
	Game      Game      `json:"game"`
}

type ChallengeRsp struct {
	In  []Challenge `json:"in"`
	Out []Challenge `json:"out"`
}

type Challenge struct {
	ID          string      `json:"id"`
	URL         string      `json:"url"`
	Status      string      `json:"status"`
	Challenger  Challenger  `json:"challenger"`
	DestUser    DestUser    `json:"destUser"`
	Variant     Variant     `json:"variant"`
	Rated       bool        `json:"rated"`
	Speed       string      `json:"speed"`
	TimeControl TimeControl `json:"timeControl"`
	Color       string      `json:"color"`
	FinalColor  string      `json:"finalColor"`
	Perf        Perf        `json:"perf"`
	Direction   string      `json:"direction"`
}
type Challenger struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Title       interface{} `json:"title"`
	Rating      int         `json:"rating"`
	Provisional bool        `json:"provisional"`
	Online      bool        `json:"online"`
}
type DestUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Rating      int    `json:"rating"`
	Provisional bool   `json:"provisional"`
	Online      bool   `json:"online"`
}
type Variant struct {
	Key   string `json:"key"`
	Name  string `json:"name"`
	Short string `json:"short"`
}
type TimeControl struct {
	Type string `json:"type"`
}
type Perf struct {
	Icon string `json:"icon"`
	Name string `json:"name"`
}

type GamesRsp struct {
	NowPlaying []NowPlaying `json:"nowPlaying"`
}

// type Variant struct {
// 	Key string `json:"key"`
// 	Name string `json:"name"`
// }
type Opponent struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Rating   int    `json:"rating"`
}
type NowPlaying struct {
	FullID   string   `json:"fullId"`
	GameID   string   `json:"gameId"`
	Fen      string   `json:"fen"`
	Color    string   `json:"color"`
	LastMove string   `json:"lastMove"`
	Source   string   `json:"source"`
	Variant  Variant  `json:"variant"`
	Speed    string   `json:"speed"`
	Perf     string   `json:"perf"`
	Rated    bool     `json:"rated"`
	HasMoved bool     `json:"hasMoved"`
	Opponent Opponent `json:"opponent"`
	IsMyTurn bool     `json:"isMyTurn"`
}

type Compat struct {
	Bot   bool `json:"bot"`
	Board bool `json:"board"`
}
type Game struct {
	GameID      string   `json:"gameId"`
	FullID      string   `json:"fullId"`
	Color       string   `json:"color"`
	Fen         string   `json:"fen"`
	HasMoved    bool     `json:"hasMoved"`
	IsMyTurn    bool     `json:"isMyTurn"`
	LastMove    string   `json:"lastMove"`
	Opponent    Opponent `json:"opponent"`
	Perf        string   `json:"perf"`
	Rated       bool     `json:"rated"`
	SecondsLeft int      `json:"secondsLeft"`
	Source      string   `json:"source"`
	Speed       string   `json:"speed"`
	Variant     Variant  `json:"variant"`
	Compat      Compat   `json:"compat"`
}

type GameState struct {
	Type       string  `json:"type"`
	ID         string  `json:"id"`
	Moves      string  `json:"moves"`
	Wtime      int     `json:"wtime"`
	Btime      int     `json:"btime"`
	Winc       int     `json:"winc"`
	Binc       int     `json:"binc"`
	Status     string  `json:"status"`
	Rated      bool    `json:"rated"`
	Variant    Variant `json:"variant"`
	Clock      Clock   `json:"clock"`
	Speed      string  `json:"speed"`
	Perf       Perf    `json:"perf"`
	CreatedAt  int64   `json:"createdAt"`
	White      White   `json:"white"`
	Black      Black   `json:"black"`
	InitialFen string  `json:"initialFen"`
	State      State   `json:"state"`
}

type Clock struct {
	Initial   int `json:"initial"`
	Increment int `json:"increment"`
}

type White struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provisional bool   `json:"provisional"`
	Rating      int    `json:"rating"`
	Title       string `json:"title"`
}
type Black struct {
	ID     string      `json:"id"`
	Name   string      `json:"name"`
	Rating int         `json:"rating"`
	Title  interface{} `json:"title"`
}
type State struct {
	Type   string `json:"type"`
	Moves  string `json:"moves"`
	Wtime  int    `json:"wtime"`
	Btime  int    `json:"btime"`
	Winc   int    `json:"winc"`
	Binc   int    `json:"binc"`
	Status string `json:"status"`
}

type MoveQueue struct {
	Fen    string
	GameID string
}
