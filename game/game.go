package game

import (
	"errors"
	"fmt"

	"golearn/board"
)

var (
	ErrFinished = errors.New("game is finished")
	ErrNoMoves  = errors.New("no moves to undo")
	ErrKo       = errors.New("ko: position repeats")
)

type Move struct {
	Color    board.Color
	Pass     bool
	Resign   bool
	Point    board.Point
	Captures []board.Point
}

func (m Move) String() string {
	switch {
	case m.Pass:
		return fmt.Sprintf("%s pass", m.Color)
	case m.Resign:
		return fmt.Sprintf("%s resign", m.Color)
	default:
		return fmt.Sprintf("%s %d,%d", m.Color, m.Point.X, m.Point.Y)
	}
}

type Game struct {
	Board *board.Board
	Komi  float64

	turn              board.Color
	moves             []Move
	snapshots         []*board.Board
	captures          map[board.Color]int
	consecutivePasses int
	finished          bool
	resignedBy        board.Color
}

func New(size int, komi float64) *Game {
	b := board.New(size)
	return &Game{
		Board:     b,
		Komi:      komi,
		turn:      board.Black,
		snapshots: []*board.Board{b.Copy()},
		captures:  map[board.Color]int{board.Black: 0, board.White: 0},
	}
}

func (g *Game) Turn() board.Color          { return g.turn }
func (g *Game) Moves() []Move              { return g.moves }
func (g *Game) Captures(c board.Color) int { return g.captures[c] }
func (g *Game) IsFinished() bool           { return g.finished }
func (g *Game) ResignedBy() board.Color    { return g.resignedBy }
func (g *Game) MoveCount() int             { return len(g.moves) }

func (g *Game) Play(x, y int) error {
	if g.finished {
		return ErrFinished
	}
	test := g.Board.Copy()
	captures, err := test.Place(x, y, g.turn)
	if err != nil {
		return err
	}
	for _, prev := range g.snapshots {
		if prev.Equal(test) {
			return ErrKo
		}
	}
	g.Board = test
	g.captures[g.turn] += len(captures)
	g.moves = append(g.moves, Move{Color: g.turn, Point: board.Point{X: x, Y: y}, Captures: captures})
	g.snapshots = append(g.snapshots, test.Copy())
	g.turn = g.turn.Opponent()
	g.consecutivePasses = 0
	return nil
}

func (g *Game) Pass() error {
	if g.finished {
		return ErrFinished
	}
	g.moves = append(g.moves, Move{Color: g.turn, Pass: true})
	g.snapshots = append(g.snapshots, g.Board.Copy())
	g.consecutivePasses++
	g.turn = g.turn.Opponent()
	if g.consecutivePasses >= 2 {
		g.finished = true
	}
	return nil
}

func (g *Game) Resign() error {
	if g.finished {
		return ErrFinished
	}
	g.moves = append(g.moves, Move{Color: g.turn, Resign: true})
	g.resignedBy = g.turn
	g.finished = true
	return nil
}

func (g *Game) Undo() error {
	if len(g.moves) == 0 {
		return ErrNoMoves
	}
	last := g.moves[len(g.moves)-1]
	g.moves = g.moves[:len(g.moves)-1]

	if last.Resign {
		g.finished = false
		g.resignedBy = board.Empty
		g.turn = last.Color
		return nil
	}

	g.snapshots = g.snapshots[:len(g.snapshots)-1]
	g.Board = g.snapshots[len(g.snapshots)-1].Copy()

	if last.Pass {
		g.consecutivePasses--
		if g.finished && g.consecutivePasses < 2 {
			g.finished = false
		}
	} else {
		g.captures[last.Color] -= len(last.Captures)
		g.consecutivePasses = 0
		for i := len(g.moves) - 1; i >= 0 && g.moves[i].Pass; i-- {
			g.consecutivePasses++
		}
	}
	g.turn = last.Color
	return nil
}

// SetTurn forces the next player. Use only when loading a position from
// an external source (SGF) where moves may not strictly alternate.
func (g *Game) SetTurn(c board.Color) { g.turn = c }

func (g *Game) Score() (black, white float64) {
	return g.Board.Score(g.Komi)
}

func (g *Game) Winner() board.Color {
	if !g.finished {
		return board.Empty
	}
	if g.resignedBy != board.Empty {
		return g.resignedBy.Opponent()
	}
	bl, wh := g.Score()
	switch {
	case bl > wh:
		return board.Black
	case wh > bl:
		return board.White
	default:
		return board.Empty
	}
}
