package main

import (
	"fmt"
	"github.com/landr0id/go-taglib"
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
			fmt.Println("Moved", oldPath, "to", newPath)
		}

		// Check to see if the file has its ID3v2 tags set
		file, err := taglib.Read(newPath)
		if err != nil {
			fmt.Println("Error reading ID3v2 tags: returned nil file struct")
			continue
		}

		// Check the title and artist
		title := file.Title()
		artist := file.Artist()
		if title == "" && artist == "" {
			fmt.Println("Song has no artist or title set in the ID3v2 tags. Setting them accordingly.")
			// The title and artist weren't present, so let's write them
			strings := regex.FindStringSubmatch(name)
			artist := strings[1]
			title := strings[2]
			strings = nil
			fmt.Printf("Title: %s\nArtist: %s\n", title, artist)

			file.SetTitle(title)
			file.SetArtist(artist)
			file.Save()
		}
		file.Close()
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
