package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jaypipes/ghw"
)

type hippo struct {
	config Config
	db     FileDB
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
	var mountPoint string
	var session Session

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
						mountPoint = part.MountPoint
						goto done
					}
				}
			}
		}
	done:

		if removeableFound && !lastRemoveable {
			if session == nil {
				sendMail(h.config, "SD Card Discovered", "Creating new archive session")
				session = NewSession(h, mountPoint)
				session.Start()
			} else {
				log.Printf("Found new media, but old sesssion never cleared")
			}
		}
		if !removeableFound && lastRemoveable {
			sendMail(h.config, "SD Card Removed", "Finishing session")
			session.Complete()
			session = nil
		}
		lastRemoveable = removeableFound

		time.Sleep(1 * time.Second)
	}
}
