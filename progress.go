package main

import (
	"fmt"
	"time"
)

type Progress struct {
	start   time.Time
	size    int64
	read    int64
	written int64
	sr      chan Op
	dr      chan Op
	dw      chan Op
	done    chan struct{}
}

func Start(size int64, sr, dr, dw chan Op) *Progress {
	p := &Progress{
		start: time.Now(),
		size:  size,
		sr:    sr,
		dr:    dr,
		dw:    dw,
		done:  make(chan struct{}),
	}
	go p.Run()
	return p
}

func (p *Progress) Step(read, written int64) {
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
	fmt.Printf("\x1B[2J")
	fmt.Printf("\x1B[H")
	fmt.Printf("file size\t%s\n", Size(p.size))
	fmt.Println()
	fmt.Printf("bytes read\t%s\n", Size(p.read))
	fmt.Printf("bytes written\t%s\n", Size(p.written))
	fmt.Println()
	fmt.Printf("read speed\t%s\n", Speed(p.Speed()))
	fmt.Println()
	fmt.Printf("changed blocks\t%s\n", Percentage(p.Changes()))
	fmt.Println()
	fmt.Printf("input read buffer\t%s\n", Percentage(len(p.sr)/cap(p.sr)))
	fmt.Printf("output read buffer\t%s\n", Percentage(len(p.dr)/cap(p.dr)))
	fmt.Printf("output write buffer\t%s\n", Percentage(len(p.dw)/cap(p.dw)))
	fmt.Println()
	fmt.Printf("time estimated\t%s\n", Duration(p.Estimated()))
	fmt.Printf("time elapsed\t%s\n", Duration(time.Since(p.start)))
}
