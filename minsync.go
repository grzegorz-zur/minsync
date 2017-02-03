package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"syscall"
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

type Op struct {
	Data   []byte
	Offset int64
}

func Sync(src, dst string, p *Progress) error {

	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	df, err := os.OpenFile(dst, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer df.Close()

	si, err := sf.Stat()
	if err != nil {
		return err
	}
	size := si.Size()
	blocks := size / BLOCK_SIZE
	if size%BLOCK_SIZE > 0 {
		blocks += 1
	}

	err = df.Truncate(size)
	if err != nil {
		return err
	}

	sr := make(chan Op, BUFFER_SIZE)
	sw := make(chan Op, BUFFER_SIZE)
	se := make(chan error)

	dr := make(chan Op, BUFFER_SIZE)
	dw := make(chan Op, BUFFER_SIZE)
	de := make(chan error)

	go ReadWrite(sf, sr, sw, se)
	go ReadWrite(df, dr, dw, de)

	defer p.End()
	p.Start(size, sr, dr, dw)
	writes := int64(0)

loop:
	for reads := int64(1); reads <= blocks; reads++ {
		var s, d Op

		select {
		case s = <-sr:
		case err = <-se:
			break loop
		}

		select {
		case d = <-dr:
		case err = <-de:
			break loop
		}

		if bytes.Compare(s.Data, d.Data) != 0 {
			dw <- Op{s.Data, s.Offset}
			writes++
		}

		p.Step(reads*BLOCK_SIZE, writes*BLOCK_SIZE)
	}

	close(sw)
	close(dw)

	<-sr
	<-dr

	if err != nil {
		return err
	}

	select {
	case err = <-se:
	case err = <-de:
	default:
	}
	if err != nil {
		return err
	}

	err = df.Sync()
	if err != nil {
		return err
	}

	return nil

}

func ReadWrite(file *os.File, read, write chan Op, errs chan error) {

	defer close(read)

	sparse := true

	for offset := int64(0); ; {
		select {
		case w, ok := <-write:
			if !ok {
				return
			}
			if sparse && Zeros(w.Data) {
				err := PunchHole(file, w.Offset, int64(len(w.Data)))
				if err == syscall.EOPNOTSUPP {
					sparse = false
					_, err = file.WriteAt(w.Data, w.Offset)
				}
				if err != nil {
					errs <- err
					return
				}
			} else {
				_, err := file.WriteAt(w.Data, w.Offset)
				if err != nil {
					errs <- err
					return
				}
			}
		default:
			data := make([]byte, BLOCK_SIZE)
			n, err := file.Read(data)
			if err != nil && err != io.EOF {
				errs <- err
				return
			}
			if n != 0 {
				read <- Op{data[:n], offset}
				offset += int64(n)
			}
		}
	}

}

func Zeros(data []byte) bool {

	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true

}
