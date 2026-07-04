package game

import (
	"bytes"
	"strings"
	"testing"

	"golearn/board"
)

func TestNewGame(t *testing.T) {
	g := New(9, 6.5)
	if g.Turn() != board.Black {
		t.Error("first turn should be Black")
	}
	if g.Board.Size != 9 {
		t.Errorf("size=%d, want 9", g.Board.Size)
	}
	if g.Komi != 6.5 {
		t.Errorf("komi=%v, want 6.5", g.Komi)
	}
}

func TestPlayAlternatesTurns(t *testing.T) {
	g := New(9, 0)
	if err := g.Play(0, 0); err != nil {
		t.Fatal(err)
	}
	if g.Turn() != board.White {
		t.Error("turn should be White after Black plays")
	}
	if err := g.Play(1, 1); err != nil {
		t.Fatal(err)
	}
	if g.Turn() != board.Black {
		t.Error("turn should be Black again")
	}
}

func TestKoRejected(t *testing.T) {
	g := New(9, 0)
	moves := [][2]int{
		{1, 0}, // B
		{2, 0}, // W
		{0, 1}, // B
		{1, 1}, // W
		{1, 2}, // B
		{2, 2}, // W
		{8, 8}, // B (filler)
		{3, 1}, // W
		{2, 1}, // B captures W at (1,1)
	}
	for i, m := range moves {
		if err := g.Play(m[0], m[1]); err != nil {
			t.Fatalf("move %d %v: %v", i, m, err)
		}
	}
	if err := g.Play(1, 1); err != ErrKo {
		t.Errorf("expected ErrKo, got %v", err)
	}
}

func TestPassPassFinishesGame(t *testing.T) {
	g := New(9, 0)
	if err := g.Pass(); err != nil {
		t.Fatal(err)
	}
	if g.IsFinished() {
		t.Error("game shouldn't be finished after one pass")
	}
	if err := g.Pass(); err != nil {
		t.Fatal(err)
	}
	if !g.IsFinished() {
		t.Error("game should be finished after two passes")
	}
}

func TestResign(t *testing.T) {
	g := New(9, 0)
	if err := g.Play(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := g.Resign(); err != nil {
		t.Fatal(err)
	}
	if !g.IsFinished() {
		t.Error("game should be finished after resign")
	}
	if g.ResignedBy() != board.White {
		t.Errorf("resigned by=%v, want White", g.ResignedBy())
	}
	if g.Winner() != board.Black {
		t.Errorf("winner=%v, want Black", g.Winner())
	}
}

func TestUndoMove(t *testing.T) {
	g := New(9, 0)
	if err := g.Play(4, 4); err != nil {
		t.Fatal(err)
	}
	if err := g.Play(3, 3); err != nil {
		t.Fatal(err)
	}
	if err := g.Undo(); err != nil {
		t.Fatal(err)
	}
	if g.Board.Get(3, 3) != board.Empty {
		t.Error("undo didn't remove last move")
	}
	if g.Turn() != board.White {
		t.Errorf("turn=%v, want White", g.Turn())
	}
	if err := g.Undo(); err != nil {
		t.Fatal(err)
	}
	if g.Board.Get(4, 4) != board.Empty {
		t.Error("second undo didn't remove first move")
	}
	if g.Turn() != board.Black {
		t.Errorf("turn=%v, want Black", g.Turn())
	}
}

func TestUndoRestoresCaptures(t *testing.T) {
	g := New(9, 0)
	g.Play(0, 0) // B
	g.Play(8, 8) // W
	g.Play(1, 0) // B
	g.Play(7, 7) // W
	g.Play(0, 1) // B captures? No, (0,0) is B and we need to capture a W.
	// Reset: instead, let B capture W.
	g = New(9, 0)
	g.Play(8, 8) // dummy B
	g.Play(0, 0) // W
	g.Play(1, 0) // B
	g.Play(7, 7) // dummy W
	g.Play(0, 1) // B captures W(0,0)
	if g.Captures(board.Black) != 1 {
		t.Fatalf("B captures=%d, want 1", g.Captures(board.Black))
	}
	if err := g.Undo(); err != nil {
		t.Fatal(err)
	}
	if g.Board.Get(0, 0) != board.White {
		t.Error("undo didn't restore captured stone")
	}
	if g.Captures(board.Black) != 0 {
		t.Errorf("B captures=%d, want 0 after undo", g.Captures(board.Black))
	}
}

func TestSGFRoundtrip(t *testing.T) {
	g := New(9, 6.5)
	g.Play(4, 4)
	g.Play(3, 3)
	g.Play(5, 5)
	g.Pass()

	var buf bytes.Buffer
	if err := WriteSGF(&buf, g, "Alice", "Bob"); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "SZ[9]") {
		t.Errorf("missing SZ[9] in %q", out)
	}
	if !strings.Contains(out, "PB[Alice]") {
		t.Errorf("missing PB in %q", out)
	}

	parsed, err := ReadSGF(strings.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Black != "Alice" || parsed.White != "Bob" {
		t.Errorf("names=%q,%q", parsed.Black, parsed.White)
	}
	if parsed.Game.MoveCount() != 4 {
		t.Errorf("moves=%d, want 4", parsed.Game.MoveCount())
	}
	if parsed.Game.Board.Get(4, 4) != board.Black {
		t.Error("position mismatch at (4,4)")
	}
	if parsed.Game.Board.Get(3, 3) != board.White {
		t.Error("position mismatch at (3,3)")
	}
	if parsed.Game.Komi != 6.5 {
		t.Errorf("komi=%v, want 6.5", parsed.Game.Komi)
	}
}

func TestSGFCoord(t *testing.T) {
	if got := sgfCoord(0, 0); got != "aa" {
		t.Errorf("sgfCoord(0,0)=%q, want aa", got)
	}
	if got := sgfCoord(3, 4); got != "de" {
		t.Errorf("sgfCoord(3,4)=%q, want de", got)
	}
	x, y, err := parseSGFCoord("de")
	if err != nil {
		t.Fatal(err)
	}
	if x != 3 || y != 4 {
		t.Errorf("got (%d,%d), want (3,4)", x, y)
	}
}
