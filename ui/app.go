package ui

import (
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"

	"golearn/board"
	"golearn/engine"
	"golearn/game"
)

type App struct {
	fyneApp fyne.App
	window  fyne.Window

	game   *game.Game
	engine engine.Engine

	humanColor board.Color
	settings   NewGameSettings

	boardWidget *BoardWidget
	sidePanel   *SidePanel

	mu         sync.Mutex
	aiThinking bool
}

func NewApp() *App {
	return &App{
		settings: NewGameSettings{
			Size:       9,
			Komi:       6.5,
			Difficulty: engine.Beginner,
			HumanColor: board.Black,
		},
	}
}

func (a *App) Run() {
	a.fyneApp = fyneapp.NewWithID("dev.golearn.app")
	a.window = a.fyneApp.NewWindow("Go Learn")
	a.window.Resize(fyne.NewSize(900, 640))
	a.buildMenu()
	a.startNewGame(a.settings)
	a.window.SetOnClosed(func() {
		if a.engine != nil {
			_ = a.engine.Close()
		}
	})
	a.window.ShowAndRun()
}

func (a *App) buildMenu() {
	a.window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("New Game…", func() {
				ShowNewGameDialog(a.window, a.settings, a.startNewGame)
			}),
			fyne.NewMenuItem("Save SGF…", a.saveSGF),
			fyne.NewMenuItem("Load SGF…", a.loadSGF),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { a.fyneApp.Quit() }),
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("How to Play", func() { ShowRulesDialog(a.window) }),
		),
	))
}

func (a *App) startNewGame(s NewGameSettings) {
	a.settings = s
	a.humanColor = s.HumanColor
	a.game = game.New(s.Size, s.Komi)

	if a.engine != nil {
		_ = a.engine.Close()
		a.engine = nil
	}

	if s.Mode == ModeVsAI {
		e, err := engine.Launch(context.Background(), engine.LaunchOptions{Difficulty: s.Difficulty})
		if err != nil {
			dialog.ShowInformation("Engine fallback",
				fmt.Sprintf("Could not launch %s engine (%v).\nFalling back to a built-in mock opponent that plays random legal moves.",
					s.Difficulty, err),
				a.window)
			e = engine.NewMockEngine()
		}
		a.engine = e
		if err := a.engine.BoardSize(s.Size); err != nil {
			dialog.ShowError(fmt.Errorf("engine boardsize: %w", err), a.window)
		}
		_ = a.engine.ClearBoard()
		_ = a.engine.Komi(s.Komi)
	}

	if a.boardWidget == nil {
		a.boardWidget = NewBoardWidget(a.game)
		a.boardWidget.OnPlay = a.handlePlay
		a.sidePanel = NewSidePanel(a.handlePass, a.handleUndo, a.handleResign, a.handleHint)
		content := container.NewBorder(nil, nil, nil, a.sidePanel.Container(), a.boardWidget)
		a.window.SetContent(content)
	} else {
		a.boardWidget.SetGame(a.game)
	}
	a.sidePanel.Update(a.game)
	if s.Mode == ModeHotSeat {
		a.sidePanel.SetStatus(fmt.Sprintf("%dx%d hot seat, komi %g", s.Size, s.Size, s.Komi))
		a.sidePanel.HintButton.Disable()
	} else {
		a.sidePanel.SetStatus(fmt.Sprintf("%dx%d, %s, komi %g", s.Size, s.Size, s.Difficulty, s.Komi))
		a.sidePanel.HintButton.Enable()
	}

	if a.engine != nil && a.humanColor == board.White {
		go a.aiMove()
	}
}

func (a *App) handlePlay(x, y int) {
	a.mu.Lock()
	if a.aiThinking || a.game.IsFinished() {
		a.mu.Unlock()
		return
	}
	if a.engine != nil && a.game.Turn() != a.humanColor {
		a.mu.Unlock()
		return
	}
	moveColor := a.game.Turn()
	if err := a.game.Play(x, y); err != nil {
		a.mu.Unlock()
		a.sidePanel.SetStatus("Illegal: " + err.Error())
		return
	}
	a.mu.Unlock()
	if a.engine != nil {
		_ = a.engine.Play(moveColor, engine.Move{X: x, Y: y})
	}
	a.refresh()
	if a.game.IsFinished() {
		a.showEndOfGame()
		return
	}
	if a.engine != nil {
		go a.aiMove()
	}
}

func (a *App) handlePass() {
	a.mu.Lock()
	if a.aiThinking || a.game.IsFinished() {
		a.mu.Unlock()
		return
	}
	if a.engine != nil && a.game.Turn() != a.humanColor {
		a.mu.Unlock()
		return
	}
	moveColor := a.game.Turn()
	_ = a.game.Pass()
	a.mu.Unlock()
	if a.engine != nil {
		_ = a.engine.Play(moveColor, engine.Move{Pass: true})
	}
	a.refresh()
	if a.game.IsFinished() {
		a.showEndOfGame()
		return
	}
	if a.engine != nil {
		go a.aiMove()
	}
}

func (a *App) handleResign() {
	a.mu.Lock()
	if a.game.IsFinished() {
		a.mu.Unlock()
		return
	}
	_ = a.game.Resign()
	a.mu.Unlock()
	a.refresh()
	a.showEndOfGame()
}

func (a *App) handleUndo() {
	a.mu.Lock()
	if a.aiThinking || len(a.game.Moves()) == 0 {
		a.mu.Unlock()
		return
	}
	if err := a.game.Undo(); err != nil {
		a.mu.Unlock()
		return
	}
	if a.engine != nil {
		_ = a.engine.Undo()
		// In vs-AI mode, if we just undid the AI's reply, also undo the human's
		// preceding move so the human is on the move again.
		if a.game.Turn() != a.humanColor && len(a.game.Moves()) > 0 {
			_ = a.game.Undo()
			_ = a.engine.Undo()
		}
	}
	a.mu.Unlock()
	a.refresh()
}

func (a *App) handleHint() {
	if a.engine == nil {
		return
	}
	a.mu.Lock()
	if a.aiThinking || a.game.IsFinished() || a.game.Turn() != a.humanColor {
		a.mu.Unlock()
		return
	}
	a.aiThinking = true
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			a.aiThinking = false
			a.mu.Unlock()
		}()
		mv, err := a.engine.GenMove(a.humanColor)
		if err != nil {
			fyne.Do(func() {
				dialog.ShowError(err, a.window)
			})
			return
		}
		_ = a.engine.Undo()
		fyne.Do(func() {
			msg := "Engine suggests: pass"
			if mv.Resign {
				msg = "Engine would resign here."
			} else if !mv.Pass {
				msg = fmt.Sprintf("Engine suggests: %s (column %d, row %d from top)",
					labelCoord(mv.X, mv.Y, a.game.Board.Size), mv.X+1, mv.Y+1)
			}
			dialog.ShowInformation("Hint", msg, a.window)
		})
	}()
}

func (a *App) aiMove() {
	if a.engine == nil {
		return
	}
	a.mu.Lock()
	if a.game.IsFinished() || a.aiThinking || a.game.Turn() == a.humanColor {
		a.mu.Unlock()
		return
	}
	a.aiThinking = true
	aiColor := a.game.Turn()
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		a.aiThinking = false
		a.mu.Unlock()
		a.refresh()
	}()

	mv, err := a.engine.GenMove(aiColor)
	if err != nil {
		fyne.Do(func() {
			dialog.ShowError(fmt.Errorf("engine genmove: %w", err), a.window)
		})
		return
	}

	a.mu.Lock()
	if mv.Resign {
		_ = a.game.Resign()
		a.mu.Unlock()
		fyne.Do(a.showEndOfGame)
		return
	}
	if mv.Pass {
		_ = a.game.Pass()
	} else {
		if err := a.game.Play(mv.X, mv.Y); err != nil {
			a.mu.Unlock()
			fyne.Do(func() {
				dialog.ShowError(fmt.Errorf("engine illegal move: %w", err), a.window)
			})
			return
		}
	}
	a.mu.Unlock()
	if a.game.IsFinished() {
		fyne.Do(a.showEndOfGame)
	}
}

func (a *App) refresh() {
	fyne.Do(func() {
		if a.boardWidget != nil {
			a.boardWidget.Refresh()
		}
		if a.sidePanel != nil {
			a.sidePanel.Update(a.game)
		}
	})
}

func (a *App) showEndOfGame() {
	bl, wh := a.game.Score()
	var msg string
	if a.game.ResignedBy() != board.Empty {
		winner := "Black"
		if a.game.ResignedBy() == board.Black {
			winner = "White"
		}
		msg = fmt.Sprintf("%s wins by resignation.", winner)
	} else {
		diff := bl - wh
		switch {
		case diff > 0:
			msg = fmt.Sprintf("Black wins by %g.\nBlack: %g  White: %g (incl. komi %g)", diff, bl, wh, a.game.Komi)
		case diff < 0:
			msg = fmt.Sprintf("White wins by %g.\nBlack: %g  White: %g (incl. komi %g)", -diff, bl, wh, a.game.Komi)
		default:
			msg = fmt.Sprintf("Tie game.\nBlack: %g  White: %g", bl, wh)
		}
	}
	ShowGameOverDialog(a.window, msg)
}

func (a *App) saveSGF() {
	d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		if wc == nil {
			return
		}
		defer wc.Close()
		blackName, whiteName := a.sgfPlayerNames()
		if err := game.WriteSGF(wc, a.game, blackName, whiteName); err != nil {
			dialog.ShowError(err, a.window)
		}
	}, a.window)
	d.SetFileName("game.sgf")
	d.Show()
}

func (a *App) sgfPlayerNames() (black, white string) {
	if a.settings.Mode == ModeHotSeat {
		return "Black", "White"
	}
	humanName := "Human"
	computerName := "Computer (" + a.settings.Difficulty.String() + ")"
	if a.humanColor == board.White {
		return computerName, humanName
	}
	return humanName, computerName
}

func (a *App) loadSGF() {
	d := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, a.window)
			return
		}
		if rc == nil {
			return
		}
		defer rc.Close()
		sgfGame, err := game.ReadSGF(rc)
		if err != nil {
			dialog.ShowError(fmt.Errorf("load sgf: %w", err), a.window)
			return
		}
		a.game = sgfGame.Game
		if a.engine != nil {
			_ = a.engine.BoardSize(a.game.Board.Size)
			_ = a.engine.ClearBoard()
			_ = a.engine.Komi(a.game.Komi)
			for _, m := range a.game.Moves() {
				if m.Resign {
					continue
				}
				_ = a.engine.Play(m.Color, engine.Move{X: m.Point.X, Y: m.Point.Y, Pass: m.Pass})
			}
		}
		a.settings.Size = a.game.Board.Size
		a.settings.Komi = a.game.Komi
		if a.settings.Mode == ModeVsAI {
			a.humanColor = board.Black
		}
		a.boardWidget.SetGame(a.game)
		a.sidePanel.Update(a.game)
		a.sidePanel.SetStatus(fmt.Sprintf("Loaded SGF (%s vs %s)", sgfGame.Black, sgfGame.White))
	}, a.window)
	d.SetFilter(storage.NewExtensionFileFilter([]string{".sgf"}))
	d.Show()
}

func labelCoord(x, y, size int) string {
	letters := "ABCDEFGHJKLMNOPQRST"
	if x < 0 || x >= len(letters) {
		return "?"
	}
	return fmt.Sprintf("%c%d", letters[x], size-y)
}
