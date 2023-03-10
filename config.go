package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Mail Mail `yaml:"mail"`
}

type Mail struct {
	Enabled  bool   `yaml:"enable"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	Smtp     string `yaml:"smtp"`
}

func loadYaml() Config {
	yfile, err := ioutil.ReadFile("hippo.yaml")

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
