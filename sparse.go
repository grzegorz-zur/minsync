// +build !linux

package main

import (
	"os"
	"syscall"
)

func PunchHole(file *os.File, offset, size int64) error {
	return syscall.EOPNOTSUPP
}
