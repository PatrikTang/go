package board

import "testing"

func TestNewBoardEmpty(t *testing.T) {
	b := New(9)
	if b.Size != 9 {
		t.Errorf("Size=%d, want 9", b.Size)
	}
	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			if got := b.Get(x, y); got != Empty {
				t.Errorf("Get(%d,%d)=%v, want Empty", x, y, got)
			}
		}
	}
}

func TestPlaceAndGet(t *testing.T) {
	b := New(9)
	caps, err := b.Place(4, 4, Black)
	if err != nil {
		t.Fatal(err)
	}
	if len(caps) != 0 {
		t.Errorf("captures=%v, want none", caps)
	}
	if b.Get(4, 4) != Black {
		t.Error("stone not placed")
	}
}

func TestPlaceOccupied(t *testing.T) {
	b := New(9)
	if _, err := b.Place(4, 4, Black); err != nil {
		t.Fatal(err)
	}
	_, err := b.Place(4, 4, White)
	if err != ErrOccupied {
		t.Errorf("err=%v, want ErrOccupied", err)
	}
}

func TestPlaceOutOfBounds(t *testing.T) {
	b := New(9)
	if _, err := b.Place(-1, 0, Black); err != ErrOutOfBounds {
		t.Errorf("err=%v, want ErrOutOfBounds", err)
	}
	if _, err := b.Place(9, 0, Black); err != ErrOutOfBounds {
		t.Errorf("err=%v, want ErrOutOfBounds", err)
	}
}

func TestCornerCapture(t *testing.T) {
	b := New(9)
	mustPlace(t, b, 0, 0, White)
	mustPlace(t, b, 1, 0, Black)
	caps, err := b.Place(0, 1, Black)
	if err != nil {
		t.Fatal(err)
	}
	if len(caps) != 1 || caps[0] != (Point{0, 0}) {
		t.Errorf("captures=%v, want [{0 0}]", caps)
	}
	if b.Get(0, 0) != Empty {
		t.Error("captured stone not removed")
	}
}

func TestSuicide(t *testing.T) {
	b := New(9)
	mustPlace(t, b, 0, 1, Black)
	mustPlace(t, b, 1, 0, Black)
	_, err := b.Place(0, 0, White)
	if err != ErrSuicide {
		t.Errorf("err=%v, want ErrSuicide", err)
	}
}

func TestSnapbackIsLegal(t *testing.T) {
	// Capturing opponent stones with a "suicide-like" move is legal.
	b := New(5)
	// Set up:
	//   . W W .
	//   W . . W
	//   . W W .
	mustPlace(t, b, 1, 0, White)
	mustPlace(t, b, 2, 0, White)
	mustPlace(t, b, 0, 1, White)
	mustPlace(t, b, 3, 1, White)
	mustPlace(t, b, 1, 2, White)
	mustPlace(t, b, 2, 2, White)
	// Black places at (1,1) — has only 1 liberty (2,1).
	if _, err := b.Place(1, 1, Black); err != nil {
		t.Fatalf("placing at (1,1): %v", err)
	}
	// Black now at (2,1) — would be suicide, but it captures B(1,1)? No wait, B(1,1) is black, same color.
	// Let me redo: White plays (2,1). B group at (1,1) has 0 libs → captured. Then W(2,1)'s liberties? Neighbors:
	// (1,1)=now empty, (3,1)=W, (2,0)=W, (2,2)=W. So W(2,1) has 1 lib at (1,1). Not suicide.
	caps, err := b.Place(2, 1, White)
	if err != nil {
		t.Fatalf("snapback should be legal, got %v", err)
	}
	if len(caps) != 1 {
		t.Errorf("captures=%d, want 1", len(caps))
	}
}

func TestMultiStoneCapture(t *testing.T) {
	b := New(9)
	mustPlace(t, b, 0, 0, White)
	mustPlace(t, b, 1, 0, White)
	mustPlace(t, b, 2, 0, Black)
	mustPlace(t, b, 0, 1, Black)
	mustPlace(t, b, 1, 1, Black)
	caps, err := b.Place(2, 0, Black) // already there — wait need a different setup
	// reset and redo
	_ = caps
	_ = err

	b = New(9)
	// Two white stones at (0,0) and (1,0); surround them.
	mustPlace(t, b, 0, 0, White)
	mustPlace(t, b, 1, 0, White)
	mustPlace(t, b, 2, 0, Black)
	mustPlace(t, b, 0, 1, Black)
	caps, err = b.Place(1, 1, Black)
	if err != nil {
		t.Fatal(err)
	}
	if len(caps) != 2 {
		t.Errorf("captures=%d, want 2", len(caps))
	}
}

func TestScoreEmptyBoard(t *testing.T) {
	b := New(9)
	bl, wh := b.Score(6.5)
	if bl != 0 {
		t.Errorf("black=%v, want 0", bl)
	}
	if wh != 6.5 {
		t.Errorf("white=%v, want 6.5", wh)
	}
}

func TestScoreSimplePosition(t *testing.T) {
	// 5x5 with vertical walls:
	// . B . W .
	// . B . W .
	// . B . W .
	// . B . W .
	// . B . W .
	b := New(5)
	for y := 0; y < 5; y++ {
		mustPlace(t, b, 1, y, Black)
		mustPlace(t, b, 3, y, White)
	}
	bl, wh := b.Score(0)
	// Black: 5 stones + column 0 (5 territory) = 10
	// White: 5 stones + column 4 (5 territory) = 10
	// Column 2 is dame.
	if bl != 10 {
		t.Errorf("black=%v, want 10", bl)
	}
	if wh != 10 {
		t.Errorf("white=%v, want 10", wh)
	}
}

func TestHashEquality(t *testing.T) {
	b1 := New(9)
	b2 := New(9)
	if b1.Hash() != b2.Hash() {
		t.Error("empty 9x9 boards should hash equal")
	}
	mustPlace(t, b1, 4, 4, Black)
	mustPlace(t, b2, 4, 4, Black)
	if b1.Hash() != b2.Hash() {
		t.Error("identical positions should hash equal")
	}
	mustPlace(t, b2, 0, 0, White)
	if b1.Hash() == b2.Hash() {
		t.Error("different positions should hash differently")
	}
}

func TestCopyIndependence(t *testing.T) {
	b1 := New(9)
	mustPlace(t, b1, 4, 4, Black)
	b2 := b1.Copy()
	mustPlace(t, b2, 3, 3, White)
	if b1.Get(3, 3) != Empty {
		t.Error("copy was not independent")
	}
	if !b1.Equal(b1.Copy()) {
		t.Error("Equal failed on copy")
	}
}

func TestOpponent(t *testing.T) {
	if Black.Opponent() != White {
		t.Error("Black.Opponent() should be White")
	}
	if White.Opponent() != Black {
		t.Error("White.Opponent() should be Black")
	}
	if Empty.Opponent() != Empty {
		t.Error("Empty.Opponent() should be Empty")
	}
}

func mustPlace(t *testing.T, b *Board, x, y int, c Color) {
	t.Helper()
	if _, err := b.Place(x, y, c); err != nil {
		t.Fatalf("Place(%d,%d,%v): %v", x, y, c, err)
	}
}
