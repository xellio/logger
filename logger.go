package logger

import (
	"log"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

var watcher *fsnotify.Watcher
var watchedFiles map[string]*File

//
// Change struct holds information about *NEW* changes in a logfile.
//
type Change struct {
	Event string
	File  *File
	Lines []*Line
}

//
// Start the logger. New changes in the given files will return on the given channel.
//
func Start(result chan *Change, files ...string) error {
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer func() {
		derr := watcher.Close()
		if derr != nil {
			log.Println(derr)
		}
	}()

	errCh := make(chan error)
	watchedFiles = make(map[string]*File)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				change, err := (watchedFiles[event.Name]).HandleEvent(event.Op)
				if err != nil {
					errCh <- err
				}
				result <- change
			case err := <-watcher.Errors:
				errCh <- err
			}
		}
	}()

	for _, path := range files {
		path = filepath.Clean(path)
		lf, err := NewFile(path)
		if err != nil {
			return err
		}

		err = watcher.Add(path)
		if err != nil {
			return err
		}

		lines, err := lf.Parse()
		if err != nil {
			return err
		}
		result <- &Change{
			File:  &lf,
			Lines: lines,
		}
		watchedFiles[path] = &lf
	}

	return <-errCh
}
