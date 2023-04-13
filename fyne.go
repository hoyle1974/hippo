package main

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type FyneGui struct {
	progress  *widget.ProgressBar
	container *fyne.Container
	image     *canvas.Image
	window    fyne.Window
}

func (f *FyneGui) SetProgress(value float64) {
	f.progress.SetValue(value)
}

func (f *FyneGui) ShowImage(image image.Image) {
	newImage := canvas.NewImageFromImage(image)
	newImage.FillMode = canvas.ImageFillContain

	f.container.Remove(f.image)
	f.container.Add(newImage)

	f.image = newImage
}

func (f *FyneGui) Start() {
	f.window.ShowAndRun()
}

func createFyneGui() Gui {
	fmt.Println("Fyne GUI")

	a := app.New()
	w := a.NewWindow("Hungry, Hungry Hippo")
	//w.SetFullScreen(true)

	image := canvas.NewImageFromResource(theme.FyneLogo())
	image.FillMode = canvas.ImageFillStretch

	progress := widget.NewProgressBar()

	c := container.NewBorder(progress, nil, nil, nil, image)
	c.Resize(fyne.NewSize(512, 512))
	w.Resize(fyne.NewSize(512, 512))

	w.SetContent(c)

	return &FyneGui{progress, c, image, w}
}
