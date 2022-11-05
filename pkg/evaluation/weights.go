package eval

import (
	"fmt"
	"os"

	"github.com/likeawizard/tofiks/pkg/board"
	"gopkg.in/yaml.v3"
)

const WEIGHT_PATH = "./weights/weights.yml"

var PieceWeights = [6]int{}

// Based on L. Kaufman - rook and knight values are adjusted by the number of pawns on the board
var PiecePawnBonus = [6][9]int{
	{},
	{},
	{-25, -19, -13, -6, 0, 6, 13, 19, 25},
	{50, 37, 25, 12, 0, -12, -25, -37, -50},
	{},
	{},
}

func LoadWeights() (*Weights, error) {
	var weights Weights
	wFile, err := os.Open(WEIGHT_PATH)
	if err != nil {
		return nil, err
	}
	defer wFile.Close()

	d := yaml.NewDecoder(wFile)
	err = d.Decode(&weights)
	if err != nil {
		return nil, err
	}
	return &weights, nil
}

func initPieceWeightLUT() {
	PieceWeights[board.PAWNS] = weights.Pieces.Pawn
	PieceWeights[board.BISHOPS] = weights.Pieces.Bishop
	PieceWeights[board.KNIGHTS] = weights.Pieces.Knight
	PieceWeights[board.ROOKS] = weights.Pieces.Rook
	PieceWeights[board.QUEENS] = weights.Pieces.Queen
	PieceWeights[board.KINGS] = weights.Pieces.King
	fmt.Println(PieceWeights)
}

type Weights struct {
	Moves  Moves  `yaml:"moves"`
	Pieces Pieces `yaml:"pieces"`
	Knight Knight `yaml:"knight"`
	Bishop Bishop `yaml:"bishop"`
	Pawn   Pawn   `yaml:"pawn"`
}

type Pieces struct {
	Pawn   int `yaml:"pawn"`
	Knight int `yaml:"knight"`
	Bishop int `yaml:"bishop"`
	Rook   int `yaml:"rook"`
	Queen  int `yaml:"queen"`
	King   int `yaml:"king"`
}

type Knight struct {
	Center22 int `yaml:"center22"`
	Center44 int `yaml:"center44"`
	InnerRim int `yaml:"innerRim"`
	OuterRim int `yaml:"outerRim"`
}

type Bishop struct {
	MajorDiag int `yaml:"majorDiag"`
	MinorDiag int `yaml:"minorDiag"`
}

type Pawn struct {
	Passed    int `yaml:"passed"`
	Protected int `yaml:"protected"`
	Doubled   int `yaml:"doubled"`
	Isolated  int `yaml:"isolated"`
	Center22  int `yaml:"center22"`
	Center44  int `yaml:"center44"`
	Advance   int `yaml:"advance"`
}

type Moves struct {
	Move    int `yaml:"move"`
	Capture int `yaml:"capture"`
}
