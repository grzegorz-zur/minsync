package main

import (
	"fmt"
	"time"
)

const (
	KB          = 1024
	MB          = 1024 * KB
	GB          = 1024 * MB
	BLOCK_SIZE  = 4 * KB
	BUFFER_SIZE = 128 * MB / BLOCK_SIZE
)

func Size(n int64) string {
	switch {
	case n >= GB:
		return fmt.Sprintf("%6.2fGB", float64(n)/float64(GB))
	case n >= MB:
		return fmt.Sprintf("%6dMB", n/MB)
	case n >= KB:
		return fmt.Sprintf("%6dKB", n/KB)
	default:
		return fmt.Sprintf("%6dB", n)
	}
}

func Speed(n int64) string {
	return Size(n) + "/s"
}

func Percentage(n int) string {
	return fmt.Sprintf("%3d%%", n)
}

func Duration(d time.Duration) string {
	return time.Duration(int64(time.Second) * int64(d.Seconds())).String()
}
