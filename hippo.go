package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
)

type hippo struct {
	lock      sync.Mutex
	config    Config
	db        FileDB
	start     time.Time
	skipped   int
	processed int
}

type Hippo interface {
	Start()
}

func NewHippo(config Config) (Hippo, error) {
	var hippo hippo
	hippo.config = config

	db, err := NewFileDB(config.DB.File)
	if err != nil {
		return nil, err
	}
	hippo.db = db

	return &hippo, nil
}

func (h *hippo) Start() {

	// initialize
	block, err := ghw.Block()
	if err != nil {
		fmt.Printf("Error getting block storage info: %v", err)
	}

	fmt.Printf("%v\n", block)

	lastRemoveable := false

	fmt.Println("----------------")
	for {
		block, err := ghw.Block()
		if err != nil {
			fmt.Printf("Error getting block storage info: %v", err)
		}

		// Check if removeable count changed
		removeableFound := false
		for _, disk := range block.Disks {
			if disk.IsRemovable {
				for _, part := range disk.Partitions {
					//fmt.Printf("  %v\n", part)
					if len(part.MountPoint) != 0 {
						removeableFound = true
					}
				}
			}
		}
		if removeableFound && !lastRemoveable {
			sendMail(h.config, "SD Card Discovered", "")
		}
		if !removeableFound && lastRemoveable {
			sendMail(h.config, "SD Card Removed", "")
		}
		lastRemoveable = removeableFound

		for _, disk := range block.Disks {
			if disk.IsRemovable {
				//fmt.Printf(" %v\n", disk)
				for _, part := range disk.Partitions {
					//fmt.Printf("  %v\n", part)
					if len(part.MountPoint) != 0 {
						h.processFolder(part.MountPoint, part.UUID)
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}

func (h *hippo) BeginArchive() {
	h.start = time.Now()
	h.skipped = 0
	h.processed = 0
}

func (h *hippo) statusTick() {
	diff := time.Now().Sub(h.start)

	if diff.Seconds() > 300 {
		h.start = time.Now()

		s := fmt.Sprintf("Processed %v files\n", h.processed)
		if h.skipped != 0 {
			s = s + fmt.Sprintf("Skipped %v files\n", h.skipped)
		}

		if h.skipped == 0 && h.processed != 0 {
			return
		}

		sendMail(h.config, "In Progress . . .", s)
	}
}

func (h *hippo) Archive(path string, info os.FileInfo) {
	h.lock.Lock()
	defer h.lock.Unlock()

	h.statusTick()

	// Have we seen this file before?
	key := MD5file(path)
	if h.db.DoesKeyExist(key) {
		h.skipped++
		return
	}

	mt := info.ModTime()
	year := mt.Format("2006")
	month := mt.Format("200601")
	day := mt.Format("20060102")

	newPath := fmt.Sprintf("/home/jstrohm/dshome/hippo/%s/%s/%s", year, month, day)
	os.MkdirAll(newPath, os.ModePerm)

	fmt.Printf("Copying %s to %s\n", info.Name(), newPath)

	CopyFile(path, newPath+"/"+info.Name())

	h.processed++
	h.db.AddKey(key, newPath+"/"+info.Name(), mt)
}

func (h *hippo) FinishArchive() {
	h.lock.Lock()
	defer h.lock.Unlock()

	s := fmt.Sprintf("Processed %v files\n", h.processed)
	if h.skipped != 0 {
		s = s + fmt.Sprintf("Skipped %v files\n", h.skipped)
	}

	if h.skipped == 0 && h.processed != 0 {
		return
	}

	sendMail(h.config, "Process Images", s)
}

func (h *hippo) processFolder(path string, uuid string) {
	fmt.Printf("	(%s) Process Folder: %s\n", uuid, path)

	h.BeginArchive()

	count := 0
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && IsImage(path) {
				count += 1
				if count < 5 {
				}
				h.Archive(path, info)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	h.FinishArchive()
}
