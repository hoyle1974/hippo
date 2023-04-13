package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"sync"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type EbitenGui struct {
	lock     sync.Mutex
	progress float64
	image    *ebiten.Image
}

func (e *EbitenGui) SetProgress(value float64) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.progress = value
}

func (e *EbitenGui) ShowImage(image image.Image) {
	e.lock.Lock()
	defer e.lock.Unlock()

	var err error
	e.image, err = ebiten.NewImageFromImage(image, ebiten.FilterLinear)
	if err != nil {
		panic(err)
	}
}

func (e *EbitenGui) Start() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hippo")
	ebiten.SetWindowResizable(true)
	if err := ebiten.RunGame(e); err != nil {
		log.Fatal(err)
	}

}

func createEbitenGui() Gui {
	fmt.Println("Ebiten GUI")

	return &EbitenGui{}
}

func (g *EbitenGui) Update(image *ebiten.Image) error {
	return nil
}

func (e *EbitenGui) Draw(screen *ebiten.Image) {
	e.lock.Lock()
	defer e.lock.Unlock()

	screenWidth := float64(screen.Bounds().Size().X)
	if e.image != nil {
		imageWidth := float64(e.image.Bounds().Size().X)

		op := &ebiten.DrawImageOptions{}

		ws := float64(screen.Bounds().Size().X)
		hs := float64(screen.Bounds().Size().Y)

		wi := float64(e.image.Bounds().Size().X)
		hi := float64(e.image.Bounds().Size().Y)

		scale := 0.0
		if ws/hs > wi/hi {
			scale = hs / hi
		} else {
			scale = ws / wi
		}
		// op.GeoM.Translate(-float64(imageWidth*scale)/2, 0)
		op.GeoM.Translate(-imageWidth/2, 0)
		op.GeoM.Scale(scale, scale)
		op.GeoM.Translate(screenWidth/2, 0)

		screen.DrawImage(e.image, op)

		ebitenutil.DrawRect(screen, 0, 0, screenWidth*e.progress, 18, color.RGBA64{65535, 65535, 65535, 32767})

	}

	ebitenutil.DrawRect(screen, 0, 0, screenWidth, 18, color.RGBA64{16384, 16384, 65535, 32767})
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Progress %v%%", int(e.progress*100)))

}

func (g *EbitenGui) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}
