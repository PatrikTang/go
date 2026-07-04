package ui

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"golearn/board"
	"golearn/engine"
)

type GameMode int

const (
	ModeVsAI GameMode = iota
	ModeHotSeat
)

const (
	modeVsAILabel    = "Vs computer"
	modeHotSeatLabel = "Two players (hot seat)"
)

func (m GameMode) String() string {
	if m == ModeHotSeat {
		return modeHotSeatLabel
	}
	return modeVsAILabel
}

func parseMode(s string) GameMode {
	if s == modeHotSeatLabel {
		return ModeHotSeat
	}
	return ModeVsAI
}

type NewGameSettings struct {
	Size       int
	Komi       float64
	Mode       GameMode
	Difficulty engine.Difficulty
	HumanColor board.Color
}

func ShowNewGameDialog(parent fyne.Window, cur NewGameSettings, onStart func(NewGameSettings)) {
	modeSel := widget.NewSelect([]string{modeVsAILabel, modeHotSeatLabel}, nil)
	modeSel.SetSelected(cur.Mode.String())

	sizeSel := widget.NewSelect([]string{"9", "13", "19"}, nil)
	sizeSel.SetSelected(strconv.Itoa(cur.Size))

	komiEntry := widget.NewEntry()
	komiEntry.SetText(fmt.Sprintf("%g", cur.Komi))

	diffSel := widget.NewSelect([]string{"Beginner", "Easy", "Intermediate", "Strong"}, nil)
	diffSel.SetSelected(cur.Difficulty.String())

	colorSel := widget.NewSelect([]string{"Black (first)", "White (second)"}, nil)
	if cur.HumanColor == board.White {
		colorSel.SetSelected("White (second)")
	} else {
		colorSel.SetSelected("Black (first)")
	}

	diffItem := widget.NewFormItem("Difficulty", diffSel)
	colorItem := widget.NewFormItem("You play", colorSel)
	applyMode := func(label string) {
		if parseMode(label) == ModeHotSeat {
			diffSel.Disable()
			colorSel.Disable()
		} else {
			diffSel.Enable()
			colorSel.Enable()
		}
	}
	modeSel.OnChanged = applyMode
	applyMode(modeSel.Selected)

	form := widget.NewForm(
		widget.NewFormItem("Mode", modeSel),
		widget.NewFormItem("Board size", sizeSel),
		widget.NewFormItem("Komi", komiEntry),
		diffItem,
		colorItem,
	)

	d := dialog.NewCustomConfirm("New Game", "Start", "Cancel", container.NewVBox(form), func(ok bool) {
		if !ok {
			return
		}
		s := NewGameSettings{Size: 9, Komi: 6.5}
		s.Mode = parseMode(modeSel.Selected)
		if v, err := strconv.Atoi(sizeSel.Selected); err == nil {
			s.Size = v
		}
		if v, err := strconv.ParseFloat(komiEntry.Text, 64); err == nil {
			s.Komi = v
		}
		s.Difficulty = engine.ParseDifficulty(diffSel.Selected)
		if colorSel.Selected == "White (second)" {
			s.HumanColor = board.White
		} else {
			s.HumanColor = board.Black
		}
		onStart(s)
	}, parent)
	d.Resize(fyne.NewSize(380, 320))
	d.Show()
}

func ShowGameOverDialog(parent fyne.Window, message string) {
	dialog.NewInformation("Game Over", message, parent).Show()
}

func ShowRulesDialog(parent fyne.Window) {
	text := "Go (Baduk / Weiqi) — the basics\n\n" +
		"• Players take turns placing stones on the empty intersections. Black moves first.\n" +
		"• Stones connected horizontally or vertically form a single group and share liberties (adjacent empty points).\n" +
		"• A group with zero liberties is captured and removed from the board.\n" +
		"• You may not play a move that would leave your own group with zero liberties (suicide), unless it captures opponent stones first.\n" +
		"• You may not immediately recreate the previous board position (the ko rule).\n" +
		"• Passing twice in succession ends the game.\n\n" +
		"Scoring (Chinese / area rules):\n" +
		"• Your score = stones on the board + empty points surrounded only by your color.\n" +
		"• Komi is added to White's score to compensate for moving second."
	d := dialog.NewInformation("How to Play Go", text, parent)
	d.Resize(fyne.NewSize(520, 380))
	d.Show()
}
