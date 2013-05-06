# music.go

music.go is a music file manager written in Go. It will monitor the desired directory (DownloadsPath) and move any found .mp3 file to your MusicPath. This program was created after I got tired of manually moving songs downloaded from various music blogs found off of [hypem](http://hypem.com).

### Add a music cronjob

1. First run `go build music.go`  
2. Open your terminal and run `crontab -e`  
3. Go into append mode and add `@reboot /path/to/compiled/binary` or `go run /path/to/source.go`

Enjoy!