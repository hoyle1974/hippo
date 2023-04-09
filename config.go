package main

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Mail    Mail    `yaml:"mail"`
	DB      DB      `yaml:"db"`
	Dev     Dev     `yaml:"dev"`
	Storage Storage `yaml:"storage"`
	Gui     bool    `yaml:"gui"`
}

type Mail struct {
	Enabled  bool   `yaml:"enable"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Smtp     string `yaml:"smtp"`
}

type DB struct {
	File string `yaml:"file"`
}

type Storage struct {
	Path string `yaml:"path"`
}

type Dev struct {
	Images string `yaml:"images"`
}

func loadYaml() Config {
	h, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	file := h + "/.hippo/hippo.yaml"

	if FileExists("./hippo.yaml") {
		file = "./hippo.yaml"
	}

	log.Printf("Loading Config: %s\n", file)
	yfile, err := ioutil.ReadFile(file)

	if err != nil {
		log.Fatal(err)
	}

	var config Config
	err2 := yaml.Unmarshal(yfile, &config)
	if err2 != nil {
		log.Fatal(err2)
	}

	return config
}
