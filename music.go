//

// Command music is a utility for monitoring a path for new music, and moving it elsewhere

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	notifier "github.com/deckarep/gosx-notifier"
    "github.com/codegangsta/cli"
    "github.com/op/go-logging"
    "github.com/landaire/pboextractor/Godeps/_workspace/src/github.com/codegangsta/cli"
)

const (
	DownloadsPath = "/Users/lander/Downloads"
	MusicPath     = "/Users/lander/ext_music"
	MaxTransfers  = 5
)

var log = logging.MustGetLogger("music")

func main() {
    app := cli.NewApp()
    app.Author = "Lander Brandt"
    app.Email = "@landaire"
    app.EnableBashCompletion = true

    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name: "source, src",
            Usage: "source directory to monitor",
        },
        cli.StringFlag{
            Name: "dest, dst",
            Usage: "destination directory to move files to",
        },
        cli.BoolFlag{
            Name: "verbose, v",
            Value: false,
            Usage: "enable verbose logging",
        },
    }

    app.Action(func (c *cli.Context) {

    })

    logging.DEBUG("Starting application")

    app.Run(os.Args)
}

func scan(source string, destination string, pattern regexp.Regexp) {
    // The channel which will send the found files to the doMove function
    pathChan := make(chan string, MaxTransfers)
    go handleFile(pathChan)
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
                pathChan <- entry.Name()
            }
        }
        time.Sleep(30 * time.Second)
    }
}

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
