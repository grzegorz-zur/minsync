// +build !linux

package main

import (
	"os"
	"syscall"
)

func ReadAhead(file *os.File, offset, length int64) error {
	return syscall.EOPNOTSUPP
}
