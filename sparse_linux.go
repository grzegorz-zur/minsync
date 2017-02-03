// +build linux

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

func PunchHole(file *os.File, offset, length int64) error {
	return unix.Fallocate(int(file.Fd()), 1|2, offset, length)
}
