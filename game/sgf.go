package game

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golearn/board"
)

type SGFGame struct {
	Game  *Game
	Black string
	White string
}

func WriteSGF(w io.Writer, g *Game, blackName, whiteName string) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "(;FF[4]CA[UTF-8]GM[1]SZ[%d]KM[%g]", g.Board.Size, g.Komi)
	if blackName != "" {
		fmt.Fprintf(bw, "PB[%s]", escapeSGF(blackName))
	}
	if whiteName != "" {
		fmt.Fprintf(bw, "PW[%s]", escapeSGF(whiteName))
	}
	for _, m := range g.moves {
		if m.Resign {
			break
		}
		tag := "B"
		if m.Color == board.White {
			tag = "W"
		}
		coord := ""
		if !m.Pass {
			coord = sgfCoord(m.Point.X, m.Point.Y)
		}
		fmt.Fprintf(bw, ";%s[%s]", tag, coord)
	}
	if g.finished && g.resignedBy != board.Empty {
		winner := "W"
		if g.resignedBy == board.White {
			winner = "B"
		}
		fmt.Fprintf(bw, "RE[%s+Resign]", winner)
	}
	fmt.Fprint(bw, ")")
	return bw.Flush()
}

func ReadSGF(r io.Reader) (*SGFGame, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	props, moves, err := parseSGF(string(data))
	if err != nil {
		return nil, err
	}
	size := 19
	if v, ok := props["SZ"]; ok {
		size, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid SZ: %v", err)
		}
	}
	komi := 6.5
	if v, ok := props["KM"]; ok {
		komi, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid KM: %v", err)
		}
	}
	g := New(size, komi)
	for i, m := range moves {
		expected := board.Black
		if m.Color == "W" {
			expected = board.White
		}
		if g.turn != expected {
			g.turn = expected
		}
		if m.Coord == "" || (size <= 19 && m.Coord == "tt") {
			if err := g.Pass(); err != nil {
				return nil, fmt.Errorf("move %d: %v", i, err)
			}
			continue
		}
		x, y, err := parseSGFCoord(m.Coord)
		if err != nil {
			return nil, fmt.Errorf("move %d: %v", i, err)
		}
		if err := g.Play(x, y); err != nil {
			return nil, fmt.Errorf("move %d (%s%s): %v", i, m.Color, m.Coord, err)
		}
	}
	return &SGFGame{Game: g, Black: props["PB"], White: props["PW"]}, nil
}

func sgfCoord(x, y int) string {
	return string([]byte{byte('a' + x), byte('a' + y)})
}

func parseSGFCoord(s string) (int, int, error) {
	if len(s) != 2 {
		return 0, 0, fmt.Errorf("bad sgf coord %q", s)
	}
	return int(s[0] - 'a'), int(s[1] - 'a'), nil
}

func escapeSGF(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `]`, `\]`)
	return s
}

type sgfMove struct {
	Color string
	Coord string
}

func parseSGF(data string) (props map[string]string, moves []sgfMove, err error) {
	props = make(map[string]string)
	data = strings.TrimSpace(data)
	if !strings.HasPrefix(data, "(") || !strings.HasSuffix(data, ")") {
		return nil, nil, errors.New("not an sgf file (missing parens)")
	}
	data = data[1 : len(data)-1]
	nodes := splitNodes(data)
	for i, node := range nodes {
		nodeProps := parseNode(node)
		if i == 0 {
			for k, v := range nodeProps {
				props[k] = v
			}
		}
		if b, ok := nodeProps["B"]; ok {
			moves = append(moves, sgfMove{Color: "B", Coord: b})
		} else if w, ok := nodeProps["W"]; ok {
			moves = append(moves, sgfMove{Color: "W", Coord: w})
		}
	}
	return
}

func splitNodes(s string) []string {
	var nodes []string
	depth := 0
	inBracket := false
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\\':
			if inBracket && i+1 < len(s) {
				i++
			}
		case '[':
			inBracket = true
		case ']':
			inBracket = false
		case '(':
			if !inBracket {
				depth++
			}
		case ')':
			if !inBracket {
				depth--
			}
		case ';':
			if !inBracket && depth == 0 {
				if i > start {
					nodes = append(nodes, s[start:i])
				}
				start = i + 1
			}
		}
	}
	if start < len(s) {
		nodes = append(nodes, s[start:])
	}
	return nodes
}

func parseNode(node string) map[string]string {
	out := make(map[string]string)
	i := 0
	for i < len(node) {
		j := i
		for j < len(node) && node[j] >= 'A' && node[j] <= 'Z' {
			j++
		}
		if j == i {
			i++
			continue
		}
		ident := node[i:j]
		i = j
		var val string
		first := true
		for i < len(node) && node[i] == '[' {
			i++
			var sb strings.Builder
			for i < len(node) && node[i] != ']' {
				if node[i] == '\\' && i+1 < len(node) {
					sb.WriteByte(node[i+1])
					i += 2
					continue
				}
				sb.WriteByte(node[i])
				i++
			}
			if i < len(node) {
				i++
			}
			if first {
				val = sb.String()
				first = false
			}
		}
		out[ident] = val
	}
	return out
}
