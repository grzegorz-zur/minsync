// +build linux

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

func PunchHole(file *os.File, offset, length int64) error {
	return unix.Fallocate(int(file.Fd()), unix.FALLOC_FL_KEEP_SIZE|unix.FALLOC_FL_PUNCH_HOLE, offset, length)
}
