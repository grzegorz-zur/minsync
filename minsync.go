package main

import (
	"bytes"
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"os"
)

func main() {

	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s source destination", os.Args[0])
		os.Exit(1)
	}
	src := os.Args[1]
	dst := os.Args[2]

	p := NewProgress(os.Stdout)
	err := Sync(src, dst, p)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(2)
	}

}

func Sync(source, destination string, progress *Progress) error {

	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer func() {
		err := src.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	dst, err := os.OpenFile(destination, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer func() {
		err := dst.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	stat, err := src.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()

	err = dst.Truncate(size)
	if err != nil {
		return err
	}

	err = ReadAhead(src, 0, size)
	if err != nil && err != unix.EOPNOTSUPP {
		return err
	}
	err = ReadAhead(dst, 0, size)
	if err != nil && err != unix.EOPNOTSUPP {
		return err
	}

	defer progress.End()
	progress.Start(size)

	s := make([]byte, BLOCK_SIZE)
	d := make([]byte, BLOCK_SIZE)
	z := make([]byte, BLOCK_SIZE)

	sparse := true

	for offset := int64(0); offset < size; {

		n, err := src.Read(s)
		if err != nil && err != io.EOF {
			return err
		}
		_, err = dst.Read(d)
		if err != nil && err != io.EOF {
			return err
		}
		progress.Read(n)

		zero := sparse && bytes.Equal(s[:n], z[:n])
		zerofailure := false

		if zero {
			err = PunchHole(dst, offset, int64(n))
			switch err {
			case nil:
				progress.Zeroed(n)
			case unix.EOPNOTSUPP:
				zerofailure = true
				sparse = false
			default:
				return err
			}
		}

		if !zero && !bytes.Equal(s[:n], d[:n]) || zerofailure {
			_, err = dst.WriteAt(s[:n], offset)
			if err != nil {
				return err
			}
			progress.Written(n)
		}

		offset += int64(n)

	}

	err = dst.Sync()
	if err != nil {
		return err
	}

	return nil

}
