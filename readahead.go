// +build !linux

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

func ReadAhead(file *os.File, offset, length int64) error {
	return unix.EOPNOTSUPP
}
