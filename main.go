package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jaypipes/ghw"
	"gopkg.in/yaml.v2"
)

var lock sync.Mutex
var seen map[string]bool
var processed []string
var config Config

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

func sendMail(subject string, body string) {
	log.Println("Mail: " + subject + "   >>> " + body)
	if !config.Mail.Enabled {
		return
	}

	//from := "hoyle.hoyle@gmail.com"
	from := config.Mail.From

	//pass := "hfuekbfvfsohveqa"
	password := config.Mail.Password

	//to := "6012099198@txt.att.net"
	to := config.Mail.To

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: " + subject + "\n\n" +
		body

	err := smtp.SendMail(fmt.Sprintf("%s:%d", config.Mail.Smtp, config.Mail.Port),
		smtp.PlainAuth("", from, password, config.Mail.Smtp),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
		return
	}
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

func isImage(path string) bool {
	path = strings.ToLower(path)
	if strings.HasSuffix(path, ".jpg") {
		return true
	}
	if strings.HasSuffix(path, ".gif") {
		return true
	}
	if strings.HasSuffix(path, ".png") {
		return true
	}
	if strings.HasSuffix(path, ".raf") {
		return true
	}
	if strings.HasSuffix(path, ".arw") {
		return true
	}
	if strings.HasSuffix(path, ".nef") {
		return true
	}

	return false
}

func md5file(path string) string {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return string(h.Sum(nil))
}

func archive(path string, info os.FileInfo) {
	lock.Lock()
	defer lock.Unlock()

	// Have we seen this file before?
	key := md5file(path)
	_, found := seen[key]
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

	processed = append(processed, path)
	seen[key] = true
}

func archiveDone() {
	lock.Lock()
	defer lock.Unlock()

	s := fmt.Sprintf("Processed %v files.\n", len(processed))

	processed = []string{}

	sendMail("Process Images", s)
}

func processFolder(path string, uuid string) {
	fmt.Printf("	(%s) Process Folder: %s\n", uuid, path)

	count := 0
	err := filepath.Walk(path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && isImage(path) {
				count += 1
				if count < 5 {
				}
				archive(path, info)
			}
			return nil
		})
	if err != nil {
		log.Println(err)
	}

	archiveDone()
}

func loadYaml() {
	yfile, err := ioutil.ReadFile("hippo.yaml")

	if err != nil {
		log.Fatal(err)
	}

	err2 := yaml.Unmarshal(yfile, &config)
	if err2 != nil {
		log.Fatal(err2)
	}
}

func main() {
	loadYaml()
	fmt.Println(config)

	// initialize
	seen = make(map[string]bool)

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
			sendMail("SD Card Discovered", "")
		}
		if !removeableFound && lastRemoveable {
			sendMail("SD Card Removed", "")
		}
		lastRemoveable = removeableFound

		for _, disk := range block.Disks {
			if disk.IsRemovable {
				//fmt.Printf(" %v\n", disk)
				for _, part := range disk.Partitions {
					//fmt.Printf("  %v\n", part)
					if len(part.MountPoint) != 0 {
						processFolder(part.MountPoint, part.UUID)
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
}
