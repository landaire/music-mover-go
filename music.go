package main

import (
	"fmt"
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
)

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

func doMove(c chan string) {
	name := <-c
	oldPath := filepath.Join(DownloadsPath, name)
	newPath := filepath.Join(MusicPath, name)
	err := moveFile(oldPath, newPath)
	if err != nil {
		fmt.Println("Error occured when trying to move the file:", err)
		return
	}
	err = os.Remove(oldPath)
	if err != nil {
		fmt.Println("Error occured when trying to remove the file:", err)
	} else {
		fmt.Println("Moved", oldPath, "to", newPath)
	}
}

func main() {
	// The channel which will send the found files to the doMove function
	c := make(chan string)
	go doMove(c)
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
				fmt.Println("Found file:", entry.Name())
				c <- entry.Name()
			}
		}
		time.Sleep(30 * time.Second)
	}
	// Wait until the channel is clear
	<-c
}
