package logger

import (
	"bytes"
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

// BufferSize for reading the files.
const BufferSize = 32

const (
	// ChangeEventCreate ...
	ChangeEventCreate = "CREATE"
	// ChangeEventWrite ...
	ChangeEventWrite = "WRITE"
	// ChangeEventRename ...
	ChangeEventRename = "RENAME"
	// ChangeEventRemove ...
	ChangeEventRemove = "REMOVE"
	// ChangeEventChmod ...
	ChangeEventChmod = "CHMOD"
	// ChangeEventUnknown ...
	ChangeEventUnknown = "UNKNOWN"
)

//
// File struct holds information about a logged file.
//
type File struct {
	Path         string
	INode        uint64
	Size         int64
	LastReadByte int64
}

//
// NewFile creates and returns a new File struct for the given path.
//
func NewFile(path string) (File, error) {
	inode, err := INode(path)
	if err != nil {
		return File{}, err
	}

	size, err := Size(path)
	if err != nil {
		return File{}, err
	}

	file := File{
		Path:         path,
		INode:        inode,
		Size:         size,
		LastReadByte: 0,
	}

	return file, err
}

//
// Parse will read and return the unparsed/new lines of the file.
//
func (f *File) Parse() (lines []*Line, err error) {

	if f.Size < f.LastReadByte {
		f.LastReadByte = 0
	}

	if f.Size == f.LastReadByte {
		return lines, nil
	}

	for {
		if f.LastReadByte >= f.Size {
			break
		}

		var lineBytes []byte
		lineBytes, err = f.NextLine()
		if err != nil || lineBytes == nil {
			err = errors.New("problem reading next line")
			break
		}

		line, err := NewLine(lineBytes, f)
		if err != nil {
			break
		}

		lines = append(lines, line)
	}
	return lines, err
}

//
// NextLine returns the next, unprocessed line in the file.
//
func (f *File) NextLine() (line []byte, err error) {
	lf, err := os.Open(f.Path)
	if err != nil {
		return line, err
	}
	defer lf.Close()

	dat := make([]byte, BufferSize)
	for {
		if f.LastReadByte >= f.Size {
			break
		}

		_, err := lf.ReadAt(dat, f.LastReadByte)
		if err != nil {
			if err == io.EOF {
				splitted := bytes.Split(dat, []byte{10})
				if len(splitted) > 1 {
					dat = splitted[0]
					f.LastReadByte++
				}
				line = append(line, dat...)
				f.LastReadByte += int64(len(line))
			}
			break
		}
		splitted := bytes.Split(dat, []byte{10})
		returnLine := false
		if len(splitted) > 1 {
			dat = splitted[0]
			f.LastReadByte++
			returnLine = true
		}
		line = append(line, dat...)
		f.LastReadByte += int64(len(dat))

		if returnLine {
			break
		}
	}

	return line, err
}

//
// HandleEvent ...
//
func (f *File) HandleEvent(op fsnotify.Op) (change *Change, err error) {

	switch op {
	case fsnotify.Create:
		return f.createEvent()
	case fsnotify.Write:
		return f.writeEvent()
	case fsnotify.Remove:
		return f.removeEvent()
	case fsnotify.Rename:
		return f.renameEvent()
	case fsnotify.Chmod:
		return f.chmodEvent()
	default:
		return f.unknownEvent()

	}

}

//
// createEvent ...
//
func (f *File) createEvent() (change *Change, err error) {
	change = &Change{
		Event: ChangeEventCreate,
		File:  f,
	}
	return change, nil
}

//
// writeEvent ...
//
func (f *File) writeEvent() (change *Change, err error) {
	change = &Change{
		Event: ChangeEventWrite,
		File:  f,
	}

	err = f.UpdateSize()
	if err != nil {
		return change, err
	}
	lines, err := f.Parse()
	if err != nil {
		return change, err
	}
	change.Lines = lines

	return change, nil
}

//
// removeEvent ...
//
func (f *File) removeEvent() (change *Change, err error) {
	change = &Change{
		Event: ChangeEventRemove,
		File:  f,
	}

	delete(watchedFiles, f.Path)
	watcher.Remove(f.Path)

	return change, nil
}

//
// renameEvent ...
//
func (f *File) renameEvent() (change *Change, err error) {
	change = &Change{
		Event: ChangeEventRename,
	}

	delete(watchedFiles, f.Path)
	watcher.Remove(f.Path)
	file, err := find(f.Path)
	if err != nil {
		log.Printf("%s was renamed and not recreated.\n", f.Path)
		return change, err
	}

	watchedFiles[f.Path] = &file
	watcher.Add(f.Path)
	change.File = &file

	lines, err := file.Parse()
	if err != nil {
		return change, err
	}
	change.Lines = lines

	return change, nil
}

//
// chmodEvent ...
//
func (f *File) chmodEvent() (change *Change, err error) {
	change = &Change{
		Event: ChangeEventChmod,
		File:  f,
	}
	return change, nil
}

//
// unknownEvent ...
//
func (f *File) unknownEvent() (change *Change, err error) {
	change = &Change{
		Event: ChangeEventUnknown,
		File:  f,
	}
	return change, nil
}

//
// find checks tries to find and return a File struct for the given path.
// If nothing is found, it will wait a second and try it again (5 times in total).
//
func find(path string) (file File, err error) {

	for i := 0; i <= 5; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}

	file, err = NewFile(path)
	return file, err
}

//
// UpdateSize will update the files size. If a change to the file is triggered, we should update the size.
//
func (f *File) UpdateSize() error {
	size, err := Size(f.Path)
	if err != nil {
		return err
	}
	f.Size = size
	return nil
}
