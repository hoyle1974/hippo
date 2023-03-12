package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type fileDB struct {
	lock     sync.Mutex
	keys     map[string]bool
	filePath string
}

type FileDB interface {
	DoesKeyExist(key string) bool
	AddKey(key string, path string, timestamp time.Time)
}

func NewFileDB(filePath string) (FileDB, error) {
	f := fileDB{}
	f.filePath = filePath
	fmt.Printf("New File DB: %v\n", filePath)
	f.keys = make(map[string]bool)

	readFile, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &f, nil
		}
		return nil, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		line := strings.Split(fileScanner.Text(), ",")
		f.keys[line[0]] = true
	}

	return &f, nil
}

func (f *fileDB) DoesKeyExist(key string) bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	_, ok := f.keys[key]
	return ok
}

func (f *fileDB) AddKey(key string, path string, timestamp time.Time) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.keys[key] = true

	file, err := os.OpenFile(f.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
	if _, err := file.WriteString(key + "," + path + "," + timestamp.String() + "\n"); err != nil {
		log.Println(err)
	}
}
