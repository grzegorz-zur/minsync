package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"runtime/pprof"
	"time"
)

const (
	KB          = 1024
	MB          = 1024 * KB
	GB          = 1024 * MB
	BLOCK_SIZE  = 4 * KB
	BUFFER_SIZE = GB / BLOCK_SIZE
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

	start := time.Now()
	reads, writes, err := Sync(src, dst)
	duration := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err.Error())
		os.Exit(2)
	}

	ratio := 0.0
	if reads > 0 {
		ratio = float64(writes) / float64(reads) * 100
	}
	fmt.Printf("reads\t%d\nwrites\t%d\nratio\t%3.2f%%\ntime\t%v\n", reads, writes, ratio, duration)
}

type Op struct {
	Data   []byte
	Offset int64
}

func Sync(src, dst string) (reads, writes int64, err error) {

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

	for reads = int64(0); reads < blocks; reads++ {
		s, d := <-sr, <-dr
		if !Compare(s.Data, d.Data) {
			dw <- Op{s.Data, s.Offset}
			writes++
		}
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
