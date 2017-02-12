// +build linux

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

func ReadAhead(file *os.File, offset, length int64) error {
	return unix.Fadvise(int(file.Fd()), offset, length, unix.FADV_SEQUENTIAL)
}
