package main

import (
	"fmt"
	"time"
)

const (
	KiB        = 1024
	MiB        = 1024 * KiB
	GiB        = 1024 * MiB
	BLOCK_SIZE = 4 * KiB
)

func Size(n int64) string {
	switch {
	case n >= GiB:
		return fmt.Sprintf("%.2fGiB", float64(n)/float64(GiB))
	case n >= MiB:
		return fmt.Sprintf("%dMiB", n/MiB)
	case n >= KiB:
		return fmt.Sprintf("%dKiB", n/KiB)
	default:
		return fmt.Sprintf("%dB", n)
	}
}

func Speed(n int64) string {
	return Size(n) + "/s"
}

func Duration(d time.Duration) string {
	return time.Duration(int64(time.Second) * int64(d.Seconds())).String()
}
