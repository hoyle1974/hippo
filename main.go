package main

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Printf("Hippo\n")

	fmt.Printf("Loading config . . .\n")
	config := loadYaml()

	fmt.Printf("Creating scanners . . .\n")
	hippo, err := NewHippo(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Starting . . .\n")
	hippo.Start()

}
