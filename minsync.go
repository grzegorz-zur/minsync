package main

import (
	"fmt"
	"io"
	"os"
)

const (
	KB    = 1024
	MB    = KB * 1024
	BLOCK = 4 * KB
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s source destination", os.Args[0])
		os.Exit(1)
	}
	src := os.Args[1]
	dst := os.Args[2]
	err := Sync(src, dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(2)
	}
}

func Sync(src, dst string) error {
	fs, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fs.Close()

	fd, err := os.OpenFile(dst, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer fd.Close()
	defer fd.Sync()

	bs := make([]byte, BLOCK)
	bd := make([]byte, BLOCK)
	offset := int64(0)

	for {
		ns, errs := fs.Read(bs)
		nd, errd := fd.Read(bd)
		if !Compare(bs[:ns], bd[:nd]) {
			_, err := fd.WriteAt(bs[:ns], offset)
			if err != nil {
				return err
			}
		}
		offset += int64(ns)
		switch {
		case errs == io.EOF:
			err = fd.Truncate(offset)
			if err != nil {
				return err
			}
			return nil
		case errd == io.EOF:
			continue
		case errs != nil:
			return errs
		case errd != nil:
			return errd
		}
	}
}

func Compare(b1, b2 []byte) bool {
	if len(b1) != len(b2) {
		return false
	}
	for i := 0; i < len(b1); i++ {
		if b1[i] != b2[i] {
			return false
		}
	}
	return true
}
