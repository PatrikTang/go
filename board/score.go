package board

// Score returns black and white area scores using Chinese (area) rules.
// Komi is added to white's score.
func (b *Board) Score(komi float64) (black, white float64) {
	emptyVisited := make([]bool, len(b.grid))

	for y := 0; y < b.Size; y++ {
		for x := 0; x < b.Size; x++ {
			i := b.idx(x, y)
			switch b.grid[i] {
			case Black:
				black++
			case White:
				white++
			case Empty:
				if emptyVisited[i] {
					continue
				}
				region, borders := b.emptyRegion(Point{x, y}, emptyVisited)
				switch {
				case borders[Black] && !borders[White]:
					black += float64(len(region))
				case borders[White] && !borders[Black]:
					white += float64(len(region))
				}
			}
		}
	}
	white += komi
	return
}

func (b *Board) emptyRegion(start Point, visited []bool) (region []Point, borders map[Color]bool) {
	borders = make(map[Color]bool)
	stack := []Point{start}
	for len(stack) > 0 {
		p := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		i := b.idx(p.X, p.Y)
		c := b.grid[i]
		if c != Empty {
			borders[c] = true
			continue
		}
		if visited[i] {
			continue
		}
		visited[i] = true
		region = append(region, p)
		for _, n := range b.neighbors(p) {
			stack = append(stack, n)
		}
	}
	return
}
