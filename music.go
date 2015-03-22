//

// Command music is a utility for monitoring a path for new music, and moving it elsewhere

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"

	"path"

	"github.com/codegangsta/cli"
	notifier "github.com/deckarep/gosx-notifier"
	"github.com/op/go-logging"
	fsnotify "gopkg.in/fsnotify.v1"
)

const (
	DefaultPattern = `(?i).*\.mp3$`
	sourceFlag     = "source"
	destFlag       = "dest"
	patternFlag    = "file-pattern"
	verboseFlag    = "verbose"
)

var (
	log                  = logging.MustGetLogger("music")
	sourcePath, destPath string
	filenamePattern      *regexp.Regexp
)

func main() {
	app := cli.NewApp()
	app.Author = "Lander Brandt"
	app.Email = "@landaire"
	app.EnableBashCompletion = true

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  sourceFlag + ", src",
			Usage: "source directory to monitor",
		},
		cli.StringFlag{
			Name:  destFlag + ", dst",
			Usage: "destination directory to move files to",
		},
		cli.StringFlag{
			Name:  patternFlag + ", fp",
			Value: DefaultPattern,
			Usage: "file pattern to detect",
		},
		cli.BoolFlag{
			Name:  verboseFlag,
			Usage: "enable verbose logging",
		},
	}

	app.Action = func(c *cli.Context) {
		outfile, _ := os.Open(os.DevNull)
		defer outfile.Close()

		backend := logging.NewLogBackend(outfile, "", 0)
		if c.Bool(verboseFlag) {
			backend = logging.NewLogBackend(os.Stdout, "", 0)
		}

		logging.SetFormatter(logging.GlogFormatter)
		logging.SetBackend(backend)
		log.Info("Starting application")

		sourcePath = c.String(sourceFlag)
		destPath = c.String(destFlag)

		filenamePattern = regexp.MustCompile(c.String(patternFlag))

		scan()
	}

	app.Run(os.Args)
}

// Scans the directory to watch for new files matching the pattern
func scan() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Could not open fsnotify watcher: %s", err)
	}

    log.Info("Watching path %s", sourcePath)
    if _, err := os.Stat(destPath); err != nil {
        if os.IsNotExist(err) {
            log.Error("Destination %s is unavailable at this time", destPath)
        }
    }

	for event := range watcher.Events {
		if event.Op != fsnotify.Create || !filenamePattern.MatchString(event.Name) {
			continue
		}

        // Ensure the destination path is available. This is here in the case such as mine where my external hard drive
        // is not always plugged in. This keeps the queue intact until the drive is connected
        for {
            _, err = os.Stat(destPath)
            if err == nil {
                break;
            } else if !os.IsNotExist(err) {
                log.Fatal(err)
            }
        }

		if err = handleFile(event.Name); err != nil {
			log.Error("Error occured when trying to move the file %s: %s", event.Name, err)
			continue
		}

		fireNotification(destFilePath(event.Name))
	}
}

// Attempts to move a file to the destination path and perform cleanup afterwards
func handleFile(file string) error {
	if stat, _ := os.Stat(file); stat.IsDir() {
		return nil
	}

	newPath := destFilePath(file)
	if err := moveFile(file, newPath); err != nil {
		return err
	}

	if err := os.Remove(file); err != nil {
		return err
	}

	return nil
}

// Moves source file to destination
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

// Fires a push notification (OS X only)
func fireNotification(newFilePath string) {
	if runtime.GOOS != "darwin" {
		return
	}

	// OS X notification
	note := notifier.NewNotification(fmt.Sprintf("Found and moved %s", filepath.Base(newFilePath)))
	note.Title = "Music Mover"
	note.Sender = "com.apple.Finder"
	note.Link = fmt.Sprintf("file://%s", newFilePath)
	note.Push()
}

// Returns the destination file path from a given original file path
func destFilePath(originalPath string) string {
	return path.Join(destPath, filepath.Base(originalPath))
}
