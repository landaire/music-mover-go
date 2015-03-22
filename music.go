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
    "io/ioutil"
    "time"
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
    log.Info("Watching path %s", sourcePath)
    if _, err := os.Stat(destPath); err != nil && os.IsNotExist(err) {
        log.Error("Destination unavailable: %s", destPath)
    }

    files := make(chan string)

    go handleFiles(files)

    for {
        readDir(sourcePath, files)
        time.Sleep(15 * time.Second)
    }
}

func readDir(dir string, validFiles chan string) {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        log.Error("Error occurred reading source dir: %s", err)
        return
    }

    for _, fileinfo := range files {
        // Not doing recursive searches
        if fileinfo.IsDir() || !filenamePattern.MatchString(fileinfo.Name()) {
            continue
        }

        validFiles <- filepath.Join(sourcePath, fileinfo.Name())
    }
}

// Attempts to move a file to the destination path and perform cleanup afterwards
func handleFiles(files chan string) {
    for file := range files {
        newPath := destFilePath(file)
        if err := moveFile(file, newPath); err != nil {
            log.Error("Could not move file %s: %s", file, err)
            continue
        }

        if err := os.Remove(file); err != nil {
            log.Error("Could not remove file %s: %s", file, err)
            continue
        }

        fireNotification(destFilePath(file))
    }
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
	note := notifier.NewNotification(fmt.Sprintf("Moved %s", filepath.Base(newFilePath)))
	note.Title = "Music Mover"
	note.Sender = "com.apple.Finder"
	note.Link = fmt.Sprintf("file://%s", newFilePath)
	note.Push()
}

// Returns the destination file path from a given original file path
func destFilePath(originalPath string) string {
    return path.Join(destPath, filepath.Base(originalPath))
}
