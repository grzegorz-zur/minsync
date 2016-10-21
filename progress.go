package main

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Progress struct {
	mutex   sync.Mutex
	output  io.Writer
	start   time.Time
	size    int64
	read    int64
	written int64
	sr      chan Op
	dr      chan Op
	dw      chan Op
	done    chan struct{}
	printed bool
}

func NewProgress(w io.Writer) *Progress {
	return &Progress{
		output: w,
		done:   make(chan struct{}),
	}
}

func (p *Progress) Start(size int64, sr, dr, dw chan Op) {
	p.start = time.Now()
	p.size = size
	p.sr = sr
	p.dr = dr
	p.dw = dw
	go p.Run()
}

func (p *Progress) Step(read, written int64) {
	defer p.mutex.Unlock()
	p.mutex.Lock()
	p.read = read
	p.written = written
}

func (p *Progress) End() {
	p.done <- struct{}{}
	<-p.done
}

func (p *Progress) Run() {
	p.Print()
	for {
		select {
		case <-time.After(time.Second):
			p.Print()
		case <-p.done:
			p.Print()
			p.done <- struct{}{}
			return
		}
	}
}

func (p *Progress) Changes() int {
	if p.read == 0 {
		return 0
	}
	return int(100 * p.written / p.read)
}

func (p *Progress) Speed() int64 {
	s := int64(time.Since(p.start).Seconds())
	if s == 0 {
		return int64(0)
	}
	return p.read / s
}

func (p *Progress) Estimated() time.Duration {
	s := p.Speed()
	if s == 0 {
		return time.Duration(0)
	}
	return time.Duration((p.size - p.read) / s * int64(time.Second))
}

func (p *Progress) Print() {
	defer p.mutex.Unlock()
	p.mutex.Lock()
	if p.printed {
		fmt.Fprintf(p.output, "\x1B[15A")
	}
	fmt.Fprintf(p.output, "\x1B[J")
	fmt.Fprintf(p.output, "file size              %s\n", Size(p.size))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "bytes read             %s\n", Size(p.read))
	fmt.Fprintf(p.output, "bytes written          %s\n", Size(p.written))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "read speed             %s\n", Speed(p.Speed()))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "changed blocks            %s\n", Percentage(p.Changes()))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "input read buffer         %s\n", Percentage(100*len(p.sr)/cap(p.sr)))
	fmt.Fprintf(p.output, "output read buffer        %s\n", Percentage(100*len(p.dr)/cap(p.dr)))
	fmt.Fprintf(p.output, "output write buffer       %s\n", Percentage(100*len(p.dw)/cap(p.dw)))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "time estimated              %s\n", Duration(p.Estimated()))
	fmt.Fprintf(p.output, "time elapsed                %s\n", Duration(time.Since(p.start)))
	p.printed = true
}
