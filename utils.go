package logger

import (
	"syscall"

	"github.com/pkg/errors"
)

//
// INode returns the inode for the given file(path).
//
func INode(file string) (uint64, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(file, &stat); err != nil {
		return 0, errors.Errorf("%s: %s", file, err.Error())
	}
	return stat.Ino, nil
}

//
// Size reutrns the filesize of the given file(path).
//
func Size(file string) (int64, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(file, &stat); err != nil {
		return 0, errors.Errorf("%s: %s", file, err.Error())
	}

	return stat.Size, nil
}
