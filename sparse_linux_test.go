// +build linux

package main

import (
	"io/ioutil"
	"os"
	"syscall"
	"testing"
)

func TestSparse(t *testing.T) {
	dir, src, dst, err := files(1*MB, 1*MB, 0.0)
	t.Log(dir, "sparse")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	os.Truncate(src, 0)
	os.Truncate(src, 1*MB)

	p := NewProgress(ioutil.Discard)
	err = Sync(src, dst, p)
	if err != nil {
		t.Fatal(err)
	}
	match, err := compareFiles(src, dst)
	if err != nil {
		t.Fatal(err)
	}
	if !match {
		t.Errorf("%s != %s", src, dst)
	}
	fi, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	stat := fi.Sys().(*syscall.Stat_t)
	blocks := stat.Blocks
	if blocks > 0 {
		t.Errorf("file is not sparse %s blocks %d", dst, blocks)
	}
}
