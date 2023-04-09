package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

type node struct {
	filepath     string
	info         os.FileInfo
	destdir      string
	thumbnaildir string
	thumbnail    string
}

func newNode(config Config, filepath string, info os.FileInfo) node {
	mt := info.ModTime()
	year := mt.Format("2006")
	month := mt.Format("200601")
	day := mt.Format("20060102")

	return node{
		filepath:     filepath,
		info:         info,
		destdir:      fmt.Sprintf("%s/hippo/%s/%s/%s", config.Storage.Path, year, month, day),
		thumbnaildir: fmt.Sprintf("%s/hippo/thumbnail/%s/%s/%s", config.Storage.Path, year, month, day),
	}
}

func (n node) destFile() string {
	return n.destdir + "/" + n.info.Name()
}

func (n node) getThumbnailDir() string {
	return n.thumbnaildir
}
func (n node) getThumbnail() string {
	return n.thumbnaildir + "/" + n.info.Name()
}

type session struct {
	hippo     *hippo
	path      string
	start     time.Time
	skipped   int
	processed int
	errors    int
	toProcess []node
}

type Session interface {
	Start()
	Complete()
}

func NewSession(hippo *hippo, path string) Session {
	s := session{}
	s.hippo = hippo
	s.path = path

	return &s
}

func (s *session) Start() {
	// Get the file list for this session
	err := filepath.Walk(s.path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && IsImage(path) {
				s.toProcess = append(s.toProcess, newNode(s.hippo.config, path, info))
			}
			return nil
		})
	if err != nil {
		log.Println(err)
		return
	}

	msg := fmt.Sprintf("Files found: %v\n", len(s.toProcess))

	// Quickly see how many files we should process, before calculating md5
	temp := []node{}
	for _, node := range s.toProcess {
		dst := node.destFile()

		// Does the file exist in the destination, are they the same size?
		if !FileExists(dst) {
			temp = append(temp, node)
		} else {
			fi, err := os.Stat(dst)
			if err != nil || fi.Size() != node.info.Size() {
				temp = append(temp, node)
			}
		}
	}
	s.toProcess = temp

	msg = msg + fmt.Sprintf("Files need to be archived: %v\n", len(s.toProcess))

	if len(s.toProcess) > 0 {
		msg = msg + fmt.Sprintf("Starting archive, will message with status every few minutes\n")
	} else {
		msg = msg + fmt.Sprintf("There are no files that need to be archived\n")
	}

	sendMail(s.hippo.config, "Starting Archive", msg)

	s.archiveFolder()
}

func (s *session) Complete() {
}

func (s *session) archiveFolder() {
	log.Printf("	Archive Folder: %s\n", s.path)

	//s.toProcess = s.toProcess[0:50]

	s.beginArchive()

	processed := []node{}
	for idx, node := range s.toProcess {
		s.hippo.feedback.SetProgress(float64(idx) / float64(len(s.toProcess)))
		if s.Archive(idx, len(s.toProcess), node, s.hippo.feedback) {
			processed = append(processed, node)
		}
	}

	s.toProcess = processed

	s.finishArchive()
}

func (s *session) beginArchive() {
	fmt.Printf("Start archive of %d images\n", len(s.toProcess))
	s.start = time.Now()
	s.skipped = 0
	s.processed = 0
	s.errors = 0
}

func (s *session) statusTick() {
	diff := time.Now().Sub(s.start)

	if diff.Seconds() > 300 {
		s.start = time.Now()

		msg := fmt.Sprintf("Archived %v files<br>", s.processed)
		if s.skipped != 0 {
			msg = msg + fmt.Sprintf("Skipped %v files<br>", s.skipped)
		}

		if s.skipped == 0 && s.processed != 0 {
			return
		}

		sendMail(s.hippo.config, "In Progress . . .", msg)
		fmt.Println(msg)
	}
}

func (s *session) Archive(idx int, max int, node node, feedback Feedback) bool {
	s.statusTick()

	// Load file into memory
	data, err := os.ReadFile(node.filepath)
	if err != nil {
		log.Printf("ReadFile: %v\n", err)
		s.errors++
		return false
	}

	// Load image and resize if possible
	if strings.HasSuffix(strings.ToLower(node.info.Name()), ".jpg") ||
		strings.HasSuffix(strings.ToLower(node.info.Name()), ".png") ||
		strings.HasSuffix(strings.ToLower(node.info.Name()), ".gif") {
		image, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			log.Printf("failed to encode (%v): %v", node.info.Name(), err)
		} else {
			go feedback.ShowImage(image)
			newImage := resize.Resize(80, 0, image, resize.Bilinear)

			os.MkdirAll(node.getThumbnailDir(), os.ModePerm)
			f, err := os.Create(node.getThumbnail())
			if err == nil {
				if err = jpeg.Encode(f, newImage, nil); err != nil {
					log.Printf("failed to encode: %v", err)
				}
			}
			f.Close()
		}
	}

	m := md5.Sum(data)
	key := base64.StdEncoding.EncodeToString(m[:])
	if s.hippo.db.DoesKeyExist(key) && FileExists(node.destFile()) {
		s.skipped++
		return false
	}

	os.MkdirAll(node.destdir, os.ModePerm)

	err = os.WriteFile(node.destFile(), data, 0666)
	if err != nil {
		log.Printf("WriteFile: %v\n", err)
		s.errors++
		return false
	}

	fmt.Printf("Archiving (%d/%d) %s\n", idx+1, max, node.destFile())
	s.processed++
	s.hippo.db.AddKey(key, node.destFile(), node.info.ModTime())

	return true
}

func (s *session) finishArchive() {
	defer fmt.Println("Finished archiving")
	fmt.Println("Writing status message")
	msg := ""

	msg = fmt.Sprintf("Processed %v files.<br>", s.processed)
	if s.skipped != 0 {
		msg = msg + fmt.Sprintf("Skipped %v files.<br>", s.skipped)
	}

	//if s.skipped == 0 && s.processed != 0 {
	//sendMail(s.hippo.config, "Processed images", msg)
	//return
	//}

	images := []string{}
	for _, node := range s.toProcess {
		if FileExists(node.getThumbnail()) {
			images = append(images, node.getThumbnail())
		}
	}
	sendStatusMail(s.hippo.config, "Processed Images", msg, images)

}
