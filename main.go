package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image"
	"sync"
)

type Feedback interface {
	SetProgress(value float64)
	ShowImage(image image.Image)
}

type FeedbackImpl struct {
	lock      sync.Mutex
	progress  *widget.ProgressBar
	container *fyne.Container
	image     *canvas.Image
}

func (f *FeedbackImpl) SetProgress(value float64) {
	f.progress.SetValue(value)
}

func (f *FeedbackImpl) ShowImage(image image.Image) {
	if f.lock.TryLock() {
		ni := canvas.NewImageFromImage(image)
		ni.FillMode = canvas.ImageFillContain

		f.container.Remove(f.image)
		f.container.Add(ni)
		f.image = ni
		f.lock.Unlock()
	}

}

type NilFeedback struct {
}

func (f *NilFeedback) SetProgress(value float64) {
}

func (f *NilFeedback) ShowImage(image image.Image) {
}


func createGui(config Config) (Feedback, fyne.Window) {
	if !config.Gui {
		return &NilFeedback{}, nil
	}
	a := app.New()
	w := a.NewWindow("Hungry, Hungry Hippo")
	w.SetFullScreen(true)

	image := canvas.NewImageFromResource(theme.FyneLogo())
	image.FillMode = canvas.ImageFillStretch

	progress := widget.NewProgressBar()

	//c := container.NewVBox(progress, image)

	c := container.NewBorder(progress, nil, nil, nil, image)
	c.Resize(fyne.NewSize(512, 512))

	w.SetContent(c)

	feedback := &FeedbackImpl{sync.Mutex{}, progress, c, image}

	return feedback, w
}

func main() {
	fmt.Printf("Hippo\n")


	fmt.Printf("Loading config . . .\n")
	config := loadYaml()

	feedback, w := createGui(config)

	fmt.Printf("Creating scanners . . .\n")
	hippo, err := NewHippo(config, feedback)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Starting . . .\n")

	if w!= nil {
		go hippo.Start()
		w.ShowAndRun()
	} else {
		hippo.Start()
	}

}
