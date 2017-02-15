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
	zeroed  int64
	done    chan struct{}
	printed bool
}

func NewProgress(w io.Writer) *Progress {
	return &Progress{
		output: w,
		done:   make(chan struct{}),
	}
}

func (p *Progress) Start(size int64) {
	p.start = time.Now()
	p.size = size
	go p.Run()
}

func (p *Progress) Read(n int) {
	defer p.mutex.Unlock()
	p.mutex.Lock()
	p.read += int64(n)
}

func (p *Progress) Written(n int) {
	defer p.mutex.Unlock()
	p.mutex.Lock()
	p.written += int64(n)
}

func (p *Progress) Zeroed(n int) {
	defer p.mutex.Unlock()
	p.mutex.Lock()
	p.zeroed += int64(n)
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
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.printed {
		fmt.Fprintf(p.output, "\x1B[10A")
	}
	fmt.Fprintf(p.output, "\x1B[J")
	fmt.Fprintf(p.output, "size            %8s\n", Size(p.size))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "read            %8s\n", Size(p.read))
	fmt.Fprintf(p.output, "written         %8s\n", Size(p.written))
	fmt.Fprintf(p.output, "zeroed          %8s\n", Size(p.zeroed))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "speed           %8s\n", Speed(p.Speed()))
	fmt.Fprintf(p.output, "\n")
	fmt.Fprintf(p.output, "time elapsed    %8s\n", Duration(time.Since(p.start)))
	fmt.Fprintf(p.output, "time left       %8s\n", Duration(p.Estimated()))
	p.printed = true
}
