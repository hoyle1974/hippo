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
	seen      map[string]bool
	processed []string
}

type Hippo interface {
	Start()
}

func NewHippo(config Config) Hippo {
	var hippo hippo
	hippo.config = config
	hippo.seen = make(map[string]bool)

	return &hippo
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
}

func (h *hippo) Archive(path string, info os.FileInfo) {
	h.lock.Lock()
	defer h.lock.Unlock()

	// Have we seen this file before?
	key := MD5file(path)
	_, found := h.seen[key]
	if found {
		// Already processed
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

	h.processed = append(h.processed, path)
	h.seen[key] = true
}

func (h *hippo) FinishArchive() {
	h.lock.Lock()
	defer h.lock.Unlock()

	s := fmt.Sprintf("Processed %v files.\n", len(h.processed))

	h.processed = []string{}

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
