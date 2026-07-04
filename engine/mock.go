package engine

import (
	"fmt"
	"math/rand"
	"sync"

	"golearn/board"
)

// MockEngine is an in-process Engine that plays uniformly random legal
// moves. It's used as a fallback when no real engine binary is available
// and is intentionally weak — adequate for development and demos.
type MockEngine struct {
	mu         sync.Mutex
	b          *board.Board
	size       int
	komi       float64
	rng        *rand.Rand
	history    []*board.Board
	moveColors []board.Color
}

func NewMockEngine() *MockEngine {
	return &MockEngine{size: 19, rng: rand.New(rand.NewSource(1))}
}

func (m *MockEngine) ensure() {
	if m.b == nil {
		if m.size == 0 {
			m.size = 19
		}
		m.b = board.New(m.size)
		m.history = []*board.Board{m.b.Copy()}
	}
}

func (m *MockEngine) Name() (string, error) { return "MockEngine", nil }

func (m *MockEngine) BoardSize(size int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.size = size
	m.b = board.New(size)
	m.history = []*board.Board{m.b.Copy()}
	m.moveColors = nil
	return nil
}

func (m *MockEngine) ClearBoard() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure()
	m.b = board.New(m.size)
	m.history = []*board.Board{m.b.Copy()}
	m.moveColors = nil
	return nil
}

func (m *MockEngine) Komi(k float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.komi = k
	return nil
}

func (m *MockEngine) Play(c board.Color, mv Move) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure()
	if mv.Pass || mv.Resign {
		m.history = append(m.history, m.b.Copy())
		m.moveColors = append(m.moveColors, c)
		return nil
	}
	test := m.b.Copy()
	if _, err := test.Place(mv.X, mv.Y, c); err != nil {
		return err
	}
	m.b = test
	m.history = append(m.history, test.Copy())
	m.moveColors = append(m.moveColors, c)
	return nil
}

func (m *MockEngine) GenMove(c board.Color) (Move, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure()
	for tries := 0; tries < 200; tries++ {
		x := m.rng.Intn(m.size)
		y := m.rng.Intn(m.size)
		if m.b.Get(x, y) != board.Empty {
			continue
		}
		test := m.b.Copy()
		if _, err := test.Place(x, y, c); err != nil {
			continue
		}
		fillsOwnEye := true
		for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			nx, ny := x+d[0], y+d[1]
			if !test.InBounds(nx, ny) {
				continue
			}
			if test.Get(nx, ny) != c {
				fillsOwnEye = false
				break
			}
		}
		if fillsOwnEye {
			continue
		}
		ko := false
		for _, p := range m.history {
			if p.Equal(test) {
				ko = true
				break
			}
		}
		if ko {
			continue
		}
		m.b = test
		m.history = append(m.history, test.Copy())
		m.moveColors = append(m.moveColors, c)
		return Move{X: x, Y: y}, nil
	}
	m.history = append(m.history, m.b.Copy())
	m.moveColors = append(m.moveColors, c)
	return Move{Pass: true}, nil
}

func (m *MockEngine) Undo() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.history) <= 1 {
		return fmt.Errorf("nothing to undo")
	}
	m.history = m.history[:len(m.history)-1]
	m.moveColors = m.moveColors[:len(m.moveColors)-1]
	m.b = m.history[len(m.history)-1].Copy()
	return nil
}

func (m *MockEngine) FinalScore() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure()
	bl, wh := m.b.Score(m.komi)
	diff := bl - wh
	switch {
	case diff > 0:
		return fmt.Sprintf("B+%g", diff), nil
	case diff < 0:
		return fmt.Sprintf("W+%g", -diff), nil
	default:
		return "0", nil
	}
}

func (m *MockEngine) Close() error { return nil }
