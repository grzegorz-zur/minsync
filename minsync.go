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
	KB    = 1024
	MB    = KB * 1024
	BLOCK = 4 * KB
)

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
	fs, err := os.Open(src)
	if err != nil {
		return
	}
	defer fs.Close()

	fd, err := os.OpenFile(dst, os.O_RDWR, 0)
	if err != nil {
		return
	}
	defer fd.Close()
	defer fd.Sync()

	bs := make([]byte, BLOCK)
	bd := make([]byte, BLOCK)
	offset := int64(0)

	for {
		ns, errs := fs.Read(bs)
		nd, errd := fd.Read(bd)
		reads++

		if !Compare(bs[:ns], bd[:nd]) {
			_, err = fd.WriteAt(bs[:ns], offset)
			writes++
			if err != nil {
				return
			}
		}
		offset += int64(ns)

		switch {
		case errs == io.EOF:
			err = fd.Truncate(offset)
			if err != nil {
				return
			}
			return
		case errd == io.EOF:
			continue
		case errs != nil:
			err = errs
			return
		case errd != nil:
			err = errs
			return
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
