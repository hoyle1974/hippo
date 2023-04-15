package main

import (
	"fmt"
	_ "net/http/pprof"
)

func main() {
	fmt.Printf("Hippo\n")

	fmt.Printf("Loading config . . .\n")
	config := loadYaml()

	var gui Gui

	if config.EbitenGui {
		gui = createEbitenGui()
	} else {
		gui = createNilGui()
	}

	fmt.Printf("Creating scanners . . .\n")
	hippo, err := NewHippo(config, gui)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Starting . . .\n")

	go hippo.Start()

	gui.Start()

}
