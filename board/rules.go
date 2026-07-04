package board

import "errors"

var (
	ErrOutOfBounds  = errors.New("move out of bounds")
	ErrOccupied     = errors.New("point already occupied")
	ErrSuicide      = errors.New("suicide move")
	ErrInvalidColor = errors.New("invalid color")
)

// Group returns the stones connected to (x, y) by color and the set of
// empty adjacent points (liberties).
func (b *Board) Group(x, y int) (stones, liberties []Point) {
	color := b.Get(x, y)
	if color == Empty {
		return nil, nil
	}
	visited := make(map[Point]bool)
	libSet := make(map[Point]bool)
	stack := []Point{{x, y}}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if visited[p] {
			continue
		}
		visited[p] = true
		stones = append(stones, p)
		for _, n := range b.neighbors(p) {
			switch b.Get(n.X, n.Y) {
			case Empty:
				libSet[n] = true
			case color:
				if !visited[n] {
					stack = append(stack, n)
				}
			}
		}
	}
	liberties = make([]Point, 0, len(libSet))
	for p := range libSet {
		liberties = append(liberties, p)
	}
	return
}

// Liberties returns the liberty count of the group at (x, y).
func (b *Board) Liberties(x, y int) int {
	_, libs := b.Group(x, y)
	return len(libs)
}

// Place places a stone of color c at (x, y), captures opponent groups
// with zero liberties, and rejects suicide. Returns captured points.
// Place does NOT check ko — callers are responsible for ko detection.
func (b *Board) Place(x, y int, c Color) ([]Point, error) {
	if !b.InBounds(x, y) {
		return nil, ErrOutOfBounds
	}
	if b.Get(x, y) != Empty {
		return nil, ErrOccupied
	}
	if c != Black && c != White {
		return nil, ErrInvalidColor
	}

	b.set(x, y, c)

	var captures []Point
	processed := make(map[Point]bool)
	opp := c.Opponent()
	for _, n := range b.neighbors(Point{x, y}) {
		if b.Get(n.X, n.Y) != opp || processed[n] {
			continue
		}
		stones, libs := b.Group(n.X, n.Y)
		for _, s := range stones {
			processed[s] = true
		}
		if len(libs) == 0 {
			for _, s := range stones {
				b.set(s.X, s.Y, Empty)
				captures = append(captures, s)
			}
		}
	}

	if b.Liberties(x, y) == 0 {
		b.set(x, y, Empty)
		for _, p := range captures {
			b.set(p.X, p.Y, opp)
		}
		return nil, ErrSuicide
	}

	return captures, nil
}
