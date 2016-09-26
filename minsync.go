package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
)

func main() {
	defer func() {
		err := recover()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v", err)
			os.Exit(3)
		}
	}()

	cpuprofile := flag.String("cpuprofile", "", "write cpu profiling data")
	flag.Parse()

	if *cpuprofile != "" {
		cpu, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Fprintf(os.Stderr, err.Error())
			os.Exit(1)
		}
		pprof.StartCPUProfile(cpu)
		defer pprof.StopCPUProfile()
	}

	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "%s source destination", os.Args[0])
		os.Exit(1)
	}
	src := flag.Arg(0)
	dst := flag.Arg(1)

	err := Sync(src, dst)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(2)
	}

}

type Op struct {
	Data   []byte
	Offset int64
}

func Sync(src, dst string) (err error) {

	sf, err := os.Open(src)
	if err != nil {
		return
	}
	defer sf.Close()

	df, err := os.OpenFile(dst, os.O_RDWR, 0)
	if err != nil {
		return
	}
	defer df.Close()

	si, err := sf.Stat()
	if err != nil {
		return
	}
	size := si.Size()
	blocks := size / BLOCK_SIZE
	if size%BLOCK_SIZE > 0 {
		blocks += 1
	}

	err = df.Truncate(size)
	if err != nil {
		return
	}

	sr := make(chan Op, BUFFER_SIZE)
	dr := make(chan Op, BUFFER_SIZE)

	sw := make(chan Op, BUFFER_SIZE)
	dw := make(chan Op, BUFFER_SIZE)

	go ReadWrite(sf, sr, sw)
	go ReadWrite(df, dr, dw)

	progress := Start(size, sr, dr, dw)
	defer progress.End()
	writes := int64(0)

	for reads := int64(1); reads <= blocks; reads++ {
		s, d := <-sr, <-dr
		if !Compare(s.Data, d.Data) {
			dw <- Op{s.Data, s.Offset}
			writes++
		}
		progress.Step(reads*BLOCK_SIZE, writes*BLOCK_SIZE)
	}

	close(sw)
	close(dw)

	<-dr
	<-sr

	err = df.Sync()
	if err != nil {
		panic(err)
	}

	return

}

func ReadWrite(file *os.File, read, write chan Op) {

	defer close(read)

	for offset := int64(0); ; {
		select {
		case w, ok := <-write:
			if !ok {
				return
			}
			_, err := file.WriteAt(w.Data, w.Offset)
			if err != nil {
				panic(err)
			}
		default:
			data := make([]byte, BLOCK_SIZE)
			n, err := file.Read(data)
			if err != nil && err != io.EOF {
				panic(err)
			}
			if n != 0 {
				read <- Op{data[:n], offset}
				offset += int64(n)
			}
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
