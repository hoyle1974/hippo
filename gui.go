package main

import (
	"fmt"
	"image"
	"time"
)

type Gui interface {
	SetProgress(value float64)
	ShowImage(image image.Image)
	Start()
}

func createNilGui() Gui {
	fmt.Println("Nil GUI")

	return &NilGui{}
}

type NilGui struct {
}

func (f *NilGui) SetProgress(value float64) {
}

func (f *NilGui) ShowImage(image image.Image) {
}

func (f *NilGui) Start() {
	for {
		time.Sleep(time.Minute)
	}
}
