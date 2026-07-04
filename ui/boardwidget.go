package ui

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"golearn/board"
	"golearn/game"
)

var (
	woodColor  = color.RGBA{222, 184, 135, 255}
	gridColor  = color.RGBA{40, 30, 20, 255}
	blackStone = color.RGBA{20, 20, 20, 255}
	whiteStone = color.RGBA{240, 240, 240, 255}
	stoneEdge  = color.RGBA{60, 60, 60, 255}
)

type BoardWidget struct {
	widget.BaseWidget

	game   *game.Game
	OnPlay func(x, y int)

	hover  *board.Point
	raster *canvas.Raster
}

func NewBoardWidget(g *game.Game) *BoardWidget {
	bw := &BoardWidget{game: g}
	bw.ExtendBaseWidget(bw)
	bw.raster = canvas.NewRaster(bw.draw)
	return bw
}

func (b *BoardWidget) SetGame(g *game.Game) {
	b.game = g
	b.hover = nil
	if b.raster != nil {
		b.raster.Refresh()
	}
}

func (b *BoardWidget) CreateRenderer() fyne.WidgetRenderer {
	return &boardRenderer{widget: b, raster: b.raster}
}

func (b *BoardWidget) Tapped(ev *fyne.PointEvent) {
	if b.game == nil || b.OnPlay == nil {
		return
	}
	g := computeGeom(int(b.Size().Width), int(b.Size().Height), b.game.Board.Size)
	p, ok := g.cellAt(int(ev.Position.X), int(ev.Position.Y))
	if !ok {
		return
	}
	b.OnPlay(p.X, p.Y)
}

func (b *BoardWidget) MouseIn(*desktop.MouseEvent) {}

func (b *BoardWidget) MouseMoved(ev *desktop.MouseEvent) {
	if b.game == nil {
		return
	}
	g := computeGeom(int(b.Size().Width), int(b.Size().Height), b.game.Board.Size)
	p, ok := g.cellAt(int(ev.Position.X), int(ev.Position.Y))
	if !ok {
		if b.hover != nil {
			b.hover = nil
			b.raster.Refresh()
		}
		return
	}
	if b.game.Board.Get(p.X, p.Y) != board.Empty {
		if b.hover != nil {
			b.hover = nil
			b.raster.Refresh()
		}
		return
	}
	if b.hover == nil || *b.hover != p {
		pp := p
		b.hover = &pp
		b.raster.Refresh()
	}
}

func (b *BoardWidget) MouseOut() {
	if b.hover != nil {
		b.hover = nil
		b.raster.Refresh()
	}
}

func (b *BoardWidget) draw(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, woodColor)
		}
	}
	if b.game == nil {
		return img
	}
	size := b.game.Board.Size
	geom := computeGeom(w, h, size)
	if geom.Cell <= 0 {
		return img
	}

	x0, y0, cell := geom.X0, geom.Y0, geom.Cell
	for i := 0; i < size; i++ {
		px := x0 + i*cell
		py := y0 + i*cell
		drawLine(img, x0, py, x0+(size-1)*cell, py, gridColor)
		drawLine(img, px, y0, px, y0+(size-1)*cell, gridColor)
	}

	hoshiR := cell / 10
	if hoshiR < 2 {
		hoshiR = 2
	}
	for _, p := range hoshiPoints(size) {
		drawDisk(img, x0+p.X*cell, y0+p.Y*cell, hoshiR, gridColor)
	}

	stoneR := cell/2 - 1
	if stoneR < 3 {
		stoneR = 3
	}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			c := b.game.Board.Get(x, y)
			if c == board.Empty {
				continue
			}
			cx := x0 + x*cell
			cy := y0 + y*cell
			if c == board.Black {
				drawDisk(img, cx, cy, stoneR, blackStone)
			} else {
				drawDisk(img, cx, cy, stoneR, whiteStone)
				drawRing(img, cx, cy, stoneR, 1, stoneEdge)
			}
		}
	}

	if mc := b.game.MoveCount(); mc > 0 {
		moves := b.game.Moves()
		last := moves[mc-1]
		if !last.Pass && !last.Resign {
			markColor := whiteStone
			if last.Color == board.White {
				markColor = blackStone
			}
			drawRing(img, x0+last.Point.X*cell, y0+last.Point.Y*cell, stoneR/2, 2, markColor)
		}
	}

	if b.hover != nil && !b.game.IsFinished() {
		c := b.game.Turn()
		fill := color.RGBA{20, 20, 20, 110}
		if c == board.White {
			fill = color.RGBA{240, 240, 240, 130}
		}
		cx := x0 + b.hover.X*cell
		cy := y0 + b.hover.Y*cell
		drawDisk(img, cx, cy, stoneR, fill)
	}

	return img
}

type boardRenderer struct {
	widget *BoardWidget
	raster *canvas.Raster
}

func (r *boardRenderer) Layout(size fyne.Size) {
	r.raster.Resize(size)
}

func (r *boardRenderer) MinSize() fyne.Size { return fyne.NewSize(360, 360) }

func (r *boardRenderer) Refresh() {
	r.raster.Refresh()
}

func (r *boardRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func (r *boardRenderer) Destroy() {}
