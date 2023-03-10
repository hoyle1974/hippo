package main

func main() {
	config := loadYaml()
	hippo := NewHippo(config)

	hippo.Start()
}
