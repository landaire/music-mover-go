package main

import (
	"fmt"
	notifier "github.com/deckarep/gosx-notifier"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const (
	DownloadsPath = "/Users/lander/Downloads"
	MusicPath     = "/Users/lander/ext_music"
	MaxTransfers  = 5
)

var regex = regexp.MustCompile(`^(?P<Artist>.*) - (?P<Title>.*)\.mp3`)

func moveFile(source, destination string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer destFile.Close()

	io.Copy(destFile, sourceFile)
	return nil
}

func handleFile(c chan string) {
	for name := range c {
		oldPath := filepath.Join(DownloadsPath, name)
		newPath := filepath.Join(MusicPath, name)
		err := moveFile(oldPath, newPath)
		if err != nil {
			fmt.Println("Error occured when trying to move the file:", err)
			continue
		}
		err = os.Remove(oldPath)
		if err != nil {
			fmt.Println("Error occured when trying to remove the file:", err)
		} else {
			// OS X notification
			note := notifier.NewNotification(fmt.Sprintf("Found and moved %s", filepath.Base(newPath)))
			note.Title = "Music Mover"
			note.Sender = "com.apple.Finder"
			note.Link = fmt.Sprintf("file://%s", newPath)
			note.Push()

			fmt.Println("Moved", oldPath, "to", newPath)
		}
	}
}

func main() {
	// The channel which will send the found files to the doMove function
	c := make(chan string, MaxTransfers)
	go handleFile(c)
	fmt.Println("Looking for new files with the extension .mp3 in", DownloadsPath)
	for {
		entries, err := ioutil.ReadDir(DownloadsPath)
		if err != nil {
			fmt.Println("Error")
			continue
		}
		for _, entry := range entries {
			match, _ := regexp.MatchString(`(?i).*\.mp3$`, entry.Name())
			if match {
				fmt.Println("\nFound file:", entry.Name())
				c <- entry.Name()
			}
		}
		time.Sleep(30 * time.Second)
	}
}
