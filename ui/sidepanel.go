package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"golearn/board"
	"golearn/game"
)

type SidePanel struct {
	turnLabel     *widget.Label
	statusLabel   *widget.Label
	capturesLabel *widget.Label
	moveLabel     *widget.Label

	PassButton   *widget.Button
	UndoButton   *widget.Button
	ResignButton *widget.Button
	HintButton   *widget.Button

	root *fyne.Container
}

func NewSidePanel(onPass, onUndo, onResign, onHint func()) *SidePanel {
	s := &SidePanel{
		turnLabel:     widget.NewLabel("Turn: Black"),
		statusLabel:   widget.NewLabel(""),
		capturesLabel: widget.NewLabel("Captures — B: 0  W: 0"),
		moveLabel:     widget.NewLabel("Move: 0"),
		HintButton:    widget.NewButton("Hint", onHint),
		PassButton:    widget.NewButton("Pass", onPass),
		UndoButton:    widget.NewButton("Undo", onUndo),
		ResignButton:  widget.NewButton("Resign", onResign),
	}
	s.turnLabel.TextStyle = fyne.TextStyle{Bold: true}
	s.root = container.NewVBox(
		s.turnLabel,
		s.statusLabel,
		widget.NewSeparator(),
		s.capturesLabel,
		s.moveLabel,
		widget.NewSeparator(),
		s.HintButton,
		s.PassButton,
		s.UndoButton,
		s.ResignButton,
	)
	return s
}

func (s *SidePanel) Container() *fyne.Container { return s.root }

func (s *SidePanel) Update(g *game.Game) {
	turn := "Black"
	if g.Turn() == board.White {
		turn = "White"
	}
	if g.IsFinished() {
		turn = "Game Over"
	}
	s.turnLabel.SetText("Turn: " + turn)
	s.capturesLabel.SetText(fmt.Sprintf("Captures — B: %d  W: %d", g.Captures(board.Black), g.Captures(board.White)))
	s.moveLabel.SetText(fmt.Sprintf("Move: %d", g.MoveCount()))
}

func (s *SidePanel) SetStatus(text string) {
	s.statusLabel.SetText(text)
}
