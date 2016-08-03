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
	MB          = KB * 1024
	BLOCK_SIZE  = 4 * KB
	BUFFER_SIZE = 1024
)

type Reading struct {
	Data   []byte
	Offset int64
	Error  error
}

func main() {
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

func Sync(src, dst string) (reads, writes int, err error) {

	readings := make(chan Reading, BUFFER_SIZE)
	stop, clean := Reader(src, readings)
	defer func() { <-clean }()
	defer close(stop)

	fd, err := os.OpenFile(dst, os.O_RDWR, 0)
	if err != nil {
		return
	}
	defer fd.Close()
	defer fd.Sync()

	bd := make([]byte, BLOCK_SIZE)

	for {
		reading := <-readings
		nd, errd := fd.Read(bd)
		reads++

		if !Compare(reading.Data, bd[:nd]) {
			_, err = fd.WriteAt(reading.Data, reading.Offset)
			writes++
			if err != nil {
				return
			}
		}

		switch {
		case reading.Error == io.EOF:
			err = fd.Truncate(reading.Offset)
			if err != nil {
				return
			}
			return
		case errd == io.EOF:
			continue
		case reading.Error != nil:
			err = reading.Error
			return
		case errd != nil:
			err = errd
			return
		}
	}
}

func Reader(name string, readings chan Reading) (stop, clean chan struct{}) {

	stop = make(chan struct{})
	clean = make(chan struct{})

	go func() {
		defer close(clean)
		defer close(readings)

		file, err := os.Open(name)
		if err != nil {
			readings <- Reading{nil, 0, err}
			return
		}
		defer file.Close()

		offset := int64(0)
		for {
			select {
			case <-stop:
				return
			default:
				data := make([]byte, BLOCK_SIZE)
				n, err := file.Read(data)
				readings <- Reading{data[:n], offset, err}
				offset += int64(n)
				if err == io.EOF {
					return
				}
			}
		}

	}()

	return
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
