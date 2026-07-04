package ui

import (
	"image"
	"image/color"

	"golearn/board"
)

type boardGeom struct {
	X0, Y0, Cell, Size int
}

func computeGeom(w, h, size int) boardGeom {
	if w <= 0 || h <= 0 || size <= 1 {
		return boardGeom{Size: size}
	}
	dim := w
	if h < w {
		dim = h
	}
	margin := dim / 20
	if margin < 8 {
		margin = 8
	}
	usable := dim - 2*margin
	cell := usable / (size - 1)
	if cell < 4 {
		cell = 4
	}
	stride := (size - 1) * cell
	x0 := (w - stride) / 2
	y0 := (h - stride) / 2
	return boardGeom{X0: x0, Y0: y0, Cell: cell, Size: size}
}

func (g boardGeom) cellAt(px, py int) (board.Point, bool) {
	if g.Cell <= 0 {
		return board.Point{}, false
	}
	gx := int(float64(px-g.X0)/float64(g.Cell) + 0.5)
	gy := int(float64(py-g.Y0)/float64(g.Cell) + 0.5)
	if gx < 0 || gx >= g.Size || gy < 0 || gy >= g.Size {
		return board.Point{}, false
	}
	dx := px - (g.X0 + gx*g.Cell)
	dy := py - (g.Y0 + gy*g.Cell)
	r := g.Cell / 2
	if dx*dx+dy*dy > r*r {
		return board.Point{}, false
	}
	return board.Point{X: gx, Y: gy}, true
}

func hoshiPoints(size int) []board.Point {
	switch size {
	case 9:
		return []board.Point{{X: 2, Y: 2}, {X: 6, Y: 2}, {X: 2, Y: 6}, {X: 6, Y: 6}, {X: 4, Y: 4}}
	case 13:
		return []board.Point{{X: 3, Y: 3}, {X: 9, Y: 3}, {X: 3, Y: 9}, {X: 9, Y: 9}, {X: 6, Y: 6}}
	case 19:
		return []board.Point{
			{X: 3, Y: 3}, {X: 9, Y: 3}, {X: 15, Y: 3},
			{X: 3, Y: 9}, {X: 9, Y: 9}, {X: 15, Y: 9},
			{X: 3, Y: 15}, {X: 9, Y: 15}, {X: 15, Y: 15},
		}
	}
	return nil
}

func drawLine(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	dx := abs(x1 - x0)
	dy := -abs(y1 - y0)
	sx := 1
	if x0 > x1 {
		sx = -1
	}
	sy := 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		img.SetRGBA(x0, y0, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func drawDisk(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	r2 := r * r
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r2 {
				img.SetRGBA(cx+x, cy+y, c)
			}
		}
	}
}

func drawRing(img *image.RGBA, cx, cy, r, thickness int, c color.RGBA) {
	rOuter := r * r
	inner := r - thickness
	if inner < 0 {
		inner = 0
	}
	rInner := inner * inner
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			d := x*x + y*y
			if d <= rOuter && d >= rInner {
				img.SetRGBA(cx+x, cy+y, c)
			}
		}
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
