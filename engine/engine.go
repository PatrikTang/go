package engine

import "golearn/board"

type Move struct {
	Pass   bool
	Resign bool
	X, Y   int
}

type Engine interface {
	Name() (string, error)
	BoardSize(size int) error
	ClearBoard() error
	Komi(k float64) error
	Play(c board.Color, m Move) error
	GenMove(c board.Color) (Move, error)
	Undo() error
	FinalScore() (string, error)
	Close() error
}

type Difficulty int

const (
	Beginner Difficulty = iota
	Easy
	Intermediate
	Strong
)

func (d Difficulty) String() string {
	switch d {
	case Beginner:
		return "Beginner"
	case Easy:
		return "Easy"
	case Intermediate:
		return "Intermediate"
	case Strong:
		return "Strong"
	}
	return "Unknown"
}

func ParseDifficulty(s string) Difficulty {
	switch s {
	case "Easy":
		return Easy
	case "Intermediate":
		return Intermediate
	case "Strong":
		return Strong
	}
	return Beginner
}
