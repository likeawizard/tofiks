package lichess

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/likeawizard/tofiks/internal/board"
	"github.com/likeawizard/tofiks/internal/config"
	eval "github.com/likeawizard/tofiks/internal/evaluation"
)

const (
	apiUrl = "https://lichess.org/api"

	accountPath   = "/account"
	challengePath = "/challenge"
	gamesPath     = "/playing"
)

type LichessConnector struct {
	Client        *http.Client
	token         string
	Config        *config.Config
	FinishedGames sync.Map
	Ponder        bool
}

func (lc *LichessConnector) request(path, method string, payload io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, apiUrl+path, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+lc.token)
	if payload != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := lc.Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	} else {
		return body, nil
	}
}

func NewLichessConnector(c *config.Config) *LichessConnector {
	return &LichessConnector{
		Client: &http.Client{},
		token:  c.Lichess.APIToken,
		Config: c,
		Ponder: c.Lichess.Ponder,
	}
}

func (lc *LichessConnector) CheckActiveGames() []NowPlaying {
	method := http.MethodGet
	path := fmt.Sprintf("%s%s", accountPath, gamesPath)
	body, err := lc.request(path, method, nil)
	if err != nil {
		return nil
	}

	games := &GamesRsp{}
	err = json.Unmarshal(body, &games)
	if err != nil {
		return nil
	} else {
		return games.NowPlaying
	}
}

func (lc *LichessConnector) MakeMove(gameID, move string) error {
	method := http.MethodPost
	path := fmt.Sprintf("/bot/game/%s/move/%s", gameID, move)

	_, err := lc.request(path, method, nil)

	return err
}

func (lc *LichessConnector) ResignGame(gameID string) error {
	method := http.MethodPost
	path := fmt.Sprintf("/bot/game/%s/resign", gameID)

	_, err := lc.request(path, method, nil)

	return err
}

func (lc *LichessConnector) GetChallenges() ([]Challenge, error) {
	method := http.MethodGet
	path := challengePath

	body, err := lc.request(path, method, nil)
	if err != nil {
		return nil, err
	}

	challenges := &ChallengeRsp{}

	err = json.Unmarshal(body, &challenges)
	if err != nil {
		return nil, err
	} else {
		return challenges.In, nil
	}
}

func (lc *LichessConnector) HandleChallenges(ch []Challenge) {
	for _, c := range ch {
		if ok, reason := lc.ShouldAccept(c); ok {
			lc.Accept(c)
		} else {
			lc.Decline(c, reason)
		}
	}
}

func (lc *LichessConnector) Accept(c Challenge) error {
	method := http.MethodPost
	path := fmt.Sprintf("%s/%s/accept", challengePath, c.ID)
	_, err := lc.request(path, method, nil)
	return err

}

func (lc *LichessConnector) Decline(c Challenge, reason string) error {
	method := http.MethodPost
	if reason == "" {
		reason = "generic"
	}
	data := url.Values{}
	data.Set("reason", reason)
	payload := data.Encode()

	path := fmt.Sprintf("%s/%s/decline", challengePath, c.ID)

	_, err := lc.request(path, method, strings.NewReader(payload))
	return err
}

func (lc *LichessConnector) OpenEventStream() (*json.Decoder, error) {
	req, err := http.NewRequest(http.MethodGet, apiUrl+"/stream/event", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+lc.token)
	resp, err := lc.Client.Do(req)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	return decoder, nil
}

func (lc *LichessConnector) ListenToEvents(decoder *json.Decoder) {
	for decoder.More() {
		var e StreamEvent
		err := decoder.Decode(&e)
		if err != nil {
			fmt.Printf("Error decoding stream event: %s\n", err)
		}

		switch e.Type {
		case EVENT_CHALLENGE:
			//fmt.Printf("New Challenge: %v\n", e.Challenge)
			lc.HandleChallenges([]Challenge{e.Challenge})
		case EVENT_GAME_START:
			//fmt.Printf("New Game: %v\n", e.Game)
			go lc.ListenToGame(e.Game)
		case "gameFinish":
			lc.MarkGameForCancelation(e.Game.GameID)
		default:
			//fmt.Printf("Unhandled event: %s\n", e.Type)
		}
	}
}

func (lc *LichessConnector) MarkGameForCancelation(gameId string) {
	lc.FinishedGames.Store(gameId, struct{}{})
}

func (lc *LichessConnector) ListenToGame(game Game) {
	//https://lichess.org/api/bot/game/stream/{gameId}
	path := fmt.Sprintf("%s/bot/game/stream/%s", apiUrl, game.GameID)
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		fmt.Printf("Error opening GameStream: %s\n", err)
	}
	req.Header.Add("Authorization", "Bearer "+lc.token)
	resp, err := lc.Client.Do(req)
	if err != nil {
		fmt.Printf("Error opening GameStream: %s\n", err)
	}

	var isWhite bool
	var e *eval.EvalEngine
	var b *board.Board = &board.Board{}
	decoder := json.NewDecoder(resp.Body)
	var ctx context.Context
	var cancel context.CancelFunc
	var timeManagment TimeManagment

	var best, ponder board.Move
	var pv []board.Move
	for decoder.More() {
		var gs GameState
		err := decoder.Decode(&gs)
		if err != nil {
			fmt.Printf("Error decoding GameState: %s\n", err)
			continue
		} else {
			fmt.Printf("Event: %s in %s\n", gs.Type, game.GameID)
		}

		if _, ok := lc.FinishedGames.LoadAndDelete(game.GameID); ok {
			if cancel != nil {
				cancel()
			}

			fmt.Printf("Game over ID: %s\n", game.GameID)

			return
		}

		switch gs.Type {
		case GAME_EVENT_FULL:
			fmt.Printf("Game started: %s\n", game.GameID)
			//TODO: hardcoded player id
			isWhite = gs.White.ID == "likeawizard-bot"
			if gs.InitialFen == "startpos" {
				b.InitDefault()
			} else {
				b.ImportFEN(gs.InitialFen)
			}
			b.PlayMovesUCI(gs.State.Moves)
			e, err = eval.NewEvalEngine(b, lc.Config)
			if err != nil {
				fmt.Printf("Error loading eval engine: %s\n", err)
				cancel()
				return
			}
			timeManagment = *NewTimeManagement(gs, isWhite)

			if isWhite == (e.Board.Side == board.WHITE) {
				fmt.Printf("My turn in %s. (FEN: %s) Thinking...\n", game.GameID, e.Board.ExportFEN())
				fmt.Printf("TimeManagment: time to think:%v, effective lag: %v\n", timeManagment.AllotTime(), timeManagment.Lag)
				ctx, cancel := timeManagment.GetTimeoutContext()
				best, ponder = e.GetMove(ctx, &pv, false)
				if best == 0 {
					return
				}
				defer cancel()
				fmt.Printf("Playing %s in %s (FEN: %s)\n", best, game.GameID, e.Board.ExportFEN())
				timeManagment.StartStopWatch()
				lc.MakeMove(game.GameID, best.String())
			}
		case GAME_EVENT_STATE:
			//fmt.Printf("New move in: %s\n", game.GameID)
			timeManagment.UpdateClock(gs)
			moves := strings.Fields(gs.Moves)
			var lastMove string

			if len(moves) != 0 {
				lastMove = moves[len(moves)-1]
				if cancel != nil {
					cancel()
				}
				e.Board.MoveUCI(lastMove)
			}

			if isWhite == (e.Board.Side == board.WHITE) {
				if lc.Ponder {
					if lastMove == ponder.String() {
						fmt.Printf("Ponder hit with: %s PV: %v\n", ponder, pv)
					} else {
						fmt.Printf("Ponder miss. Expected: %s, Played: %s\n", ponder, lastMove)
						pv = []board.Move{}
					}
				}

				fmt.Printf("My turn in %s. (FEN: %s) Thinking...\n", game.GameID, e.Board.ExportFEN())
				fmt.Printf("TimeManagment: time to think:%v, effective lag: %v\n", timeManagment.AllotTime(), timeManagment.Lag)
				ctx, cancel = timeManagment.GetTimeoutContext()
				defer cancel()
				best, ponder = e.GetMove(ctx, &pv, false)
				if best == 0 {
					return
				}
				fmt.Printf("Playing %s in %s (FEN: %s)\n", best, game.GameID, e.Board.ExportFEN())
				timeManagment.StartStopWatch()
				lc.MakeMove(game.GameID, best.String())
			} else {
				timeManagment.MeasureLag()
				if lc.Ponder {
					ctx, cancel = timeManagment.GetPonderContext()
					umove := e.Board.MakeMove(ponder)
					e.GetMove(ctx, &pv, true)
					umove()
				}
			}

		default:
			fmt.Printf("Unhandled game state: %s\n", gs.Type)
		}

	}
}
