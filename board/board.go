package board

import "hash/fnv"

type Color int8

const (
	Empty Color = iota
	Black
	White
)

func (c Color) Opponent() Color {
	switch c {
	case Black:
		return White
	case White:
		return Black
	}
	return Empty
}

func (c Color) String() string {
	switch c {
	case Black:
		return "B"
	case White:
		return "W"
	}
	return "."
}

type Point struct {
	X, Y int
}

type Board struct {
	Size int
	grid []Color
}

func New(size int) *Board {
	return &Board{Size: size, grid: make([]Color, size*size)}
}

func (b *Board) idx(x, y int) int { return y*b.Size + x }

func (b *Board) Get(x, y int) Color {
	if !b.InBounds(x, y) {
		return Empty
	}
	return b.grid[b.idx(x, y)]
}

func (b *Board) set(x, y int, c Color) {
	b.grid[b.idx(x, y)] = c
}

func (b *Board) InBounds(x, y int) bool {
	return x >= 0 && x < b.Size && y >= 0 && y < b.Size
}

func (b *Board) Copy() *Board {
	g := make([]Color, len(b.grid))
	copy(g, b.grid)
	return &Board{Size: b.Size, grid: g}
}

func (b *Board) Equal(o *Board) bool {
	if o == nil || b.Size != o.Size {
		return false
	}
	for i := range b.grid {
		if b.grid[i] != o.grid[i] {
			return false
		}
	}
	return true
}

func (b *Board) Hash() uint64 {
	h := fnv.New64a()
	buf := make([]byte, len(b.grid))
	for i, c := range b.grid {
		buf[i] = byte(c)
	}
	h.Write(buf)
	return h.Sum64()
}

func (b *Board) neighbors(p Point) []Point {
	out := make([]Point, 0, 4)
	for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		n := Point{p.X + d[0], p.Y + d[1]}
		if b.InBounds(n.X, n.Y) {
			out = append(out, n)
		}
	}
	return out
}
