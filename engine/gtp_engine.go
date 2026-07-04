package engine

import (
	"fmt"
	"strconv"
	"strings"

	"golearn/board"
)

const gtpColumnLetters = "ABCDEFGHJKLMNOPQRST"

type GTPEngine struct {
	c    *GTPClient
	size int
}

func NewGTPEngine(c *GTPClient) *GTPEngine {
	return &GTPEngine{c: c, size: 19}
}

func (e *GTPEngine) Client() *GTPClient { return e.c }

func (e *GTPEngine) Name() (string, error) { return e.c.Command("name") }

func (e *GTPEngine) BoardSize(size int) error {
	if _, err := e.c.Command(fmt.Sprintf("boardsize %d", size)); err != nil {
		return err
	}
	e.size = size
	return nil
}

func (e *GTPEngine) ClearBoard() error {
	_, err := e.c.Command("clear_board")
	return err
}

func (e *GTPEngine) Komi(k float64) error {
	_, err := e.c.Command(fmt.Sprintf("komi %g", k))
	return err
}

func (e *GTPEngine) Play(c board.Color, m Move) error {
	coord := "pass"
	if !m.Pass {
		coord = gtpCoord(m.X, m.Y, e.size)
	}
	color := "B"
	if c == board.White {
		color = "W"
	}
	_, err := e.c.Command(fmt.Sprintf("play %s %s", color, coord))
	return err
}

func (e *GTPEngine) GenMove(c board.Color) (Move, error) {
	color := "B"
	if c == board.White {
		color = "W"
	}
	resp, err := e.c.Command("genmove " + color)
	if err != nil {
		return Move{}, err
	}
	resp = strings.TrimSpace(resp)
	switch strings.ToUpper(resp) {
	case "PASS":
		return Move{Pass: true}, nil
	case "RESIGN":
		return Move{Resign: true}, nil
	}
	x, y, err := parseGTPCoord(resp, e.size)
	if err != nil {
		return Move{}, err
	}
	return Move{X: x, Y: y}, nil
}

func (e *GTPEngine) Undo() error {
	_, err := e.c.Command("undo")
	return err
}

func (e *GTPEngine) FinalScore() (string, error) {
	return e.c.Command("final_score")
}

func (e *GTPEngine) Close() error { return e.c.Close() }

// gtpCoord converts internal (x, y) where (0,0) is top-left to a GTP
// coordinate like "D4". GTP columns skip 'I'; row 1 is the bottom row.
func gtpCoord(x, y, size int) string {
	if x < 0 || x >= len(gtpColumnLetters) {
		return ""
	}
	col := gtpColumnLetters[x]
	row := size - y
	return fmt.Sprintf("%c%d", col, row)
}

func parseGTPCoord(s string, size int) (int, int, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if len(s) < 2 {
		return 0, 0, fmt.Errorf("bad coord %q", s)
	}
	col := strings.IndexByte(gtpColumnLetters, s[0])
	if col < 0 {
		return 0, 0, fmt.Errorf("bad column in %q", s)
	}
	row, err := strconv.Atoi(s[1:])
	if err != nil {
		return 0, 0, fmt.Errorf("bad row in %q", s)
	}
	return col, size - row, nil
}
