package engine

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"golearn/board"
)

// fakeServer reads lines from r and responds via w using simple GTP rules.
func fakeServer(r io.Reader, w io.Writer) {
	br := bufio.NewReader(r)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return
		}
		cmd := strings.TrimSpace(line)
		switch {
		case cmd == "name":
			io.WriteString(w, "= fakeEngine\n\n")
		case strings.HasPrefix(cmd, "boardsize"):
			io.WriteString(w, "=\n\n")
		case strings.HasPrefix(cmd, "clear_board"):
			io.WriteString(w, "=\n\n")
		case strings.HasPrefix(cmd, "komi"):
			io.WriteString(w, "=\n\n")
		case strings.HasPrefix(cmd, "play"):
			io.WriteString(w, "=\n\n")
		case strings.HasPrefix(cmd, "genmove"):
			io.WriteString(w, "= D4\n\n")
		case strings.HasPrefix(cmd, "undo"):
			io.WriteString(w, "=\n\n")
		case strings.HasPrefix(cmd, "final_score"):
			io.WriteString(w, "= B+12.5\n\n")
		case cmd == "bogus":
			io.WriteString(w, "? unknown command\n\n")
		default:
			io.WriteString(w, "=\n\n")
		}
	}
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

func newTestClient(t *testing.T) *GTPClient {
	t.Helper()
	clientToServer := newPipe()
	serverToClient := newPipe()
	go fakeServer(clientToServer.r, serverToClient.w)
	return NewGTPClient(serverToClient.r, clientToServer.w, nopCloser{})
}

type pipePair struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func newPipe() pipePair {
	r, w := io.Pipe()
	return pipePair{r: r, w: w}
}

func TestGTPCommandSuccess(t *testing.T) {
	c := newTestClient(t)
	body, err := c.Command("name")
	if err != nil {
		t.Fatal(err)
	}
	if body != "fakeEngine" {
		t.Errorf("body=%q, want fakeEngine", body)
	}
}

func TestGTPCommandError(t *testing.T) {
	c := newTestClient(t)
	_, err := c.Command("bogus")
	if err == nil {
		t.Error("expected error for bogus command")
	}
}

func TestGTPEngineFullCycle(t *testing.T) {
	c := newTestClient(t)
	e := NewGTPEngine(c)
	if err := e.BoardSize(9); err != nil {
		t.Fatal(err)
	}
	if err := e.ClearBoard(); err != nil {
		t.Fatal(err)
	}
	if err := e.Komi(6.5); err != nil {
		t.Fatal(err)
	}
	if err := e.Play(board.Black, Move{X: 3, Y: 5}); err != nil {
		t.Fatal(err)
	}
	mv, err := e.GenMove(board.White)
	if err != nil {
		t.Fatal(err)
	}
	if mv.X != 3 || mv.Y != 5 {
		t.Errorf("genmove returned (%d,%d), want (3,5)", mv.X, mv.Y)
	}
	if score, err := e.FinalScore(); err != nil || score != "B+12.5" {
		t.Errorf("score=%q err=%v", score, err)
	}
}

func TestGTPCoord(t *testing.T) {
	cases := []struct {
		x, y, size int
		want       string
	}{
		{0, 8, 9, "A1"},
		{3, 5, 9, "D4"},
		{8, 0, 9, "J9"},
		{0, 18, 19, "A1"},
		{18, 0, 19, "T19"},
	}
	for _, c := range cases {
		got := gtpCoord(c.x, c.y, c.size)
		if got != c.want {
			t.Errorf("gtpCoord(%d,%d,%d)=%q, want %q", c.x, c.y, c.size, got, c.want)
		}
		x, y, err := parseGTPCoord(c.want, c.size)
		if err != nil {
			t.Errorf("parseGTPCoord(%q): %v", c.want, err)
			continue
		}
		if x != c.x || y != c.y {
			t.Errorf("parseGTPCoord(%q)=(%d,%d), want (%d,%d)", c.want, x, y, c.x, c.y)
		}
	}
}

func TestMockEnginePlays(t *testing.T) {
	m := NewMockEngine()
	if err := m.BoardSize(9); err != nil {
		t.Fatal(err)
	}
	if err := m.ClearBoard(); err != nil {
		t.Fatal(err)
	}
	if err := m.Komi(6.5); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		c := board.Black
		if i%2 == 1 {
			c = board.White
		}
		mv, err := m.GenMove(c)
		if err != nil {
			t.Fatalf("genmove %d: %v", i, err)
		}
		if mv.Resign {
			t.Errorf("mock should not resign")
		}
	}
}

func TestMockEngineRejectsIllegal(t *testing.T) {
	m := NewMockEngine()
	m.BoardSize(9)
	m.ClearBoard()
	if err := m.Play(board.Black, Move{X: 4, Y: 4}); err != nil {
		t.Fatal(err)
	}
	if err := m.Play(board.White, Move{X: 4, Y: 4}); err == nil {
		t.Error("expected error placing on occupied point")
	}
}
