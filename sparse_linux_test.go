// +build linux

package main

import (
	"io/ioutil"
	"os"
	"syscall"
	"testing"
)

func TestSparse(t *testing.T) {

	sparse, err := ioutil.TempFile("", "sparse_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(sparse.Name())
	err = sparse.Truncate(KB)
	if err != nil {
		t.Fatal(err)
	}
	fi, err := sparse.Stat()
	if err != nil {
		t.Fatal(err)
	}
	blocks := fi.Sys().(*syscall.Stat_t).Blocks
	if blocks > 0 {
		t.Skip("sparse files not supported")
	}

	dir, src, dst, err := randomFiles(MB, MB, 1, 1)
	t.Log(dir)
	if err != nil {
		t.Fatal(err)
	}

	p := NewProgress(ioutil.Discard)
	defer t.Logf("%+v", p)
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
	fi, err = os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	blocks = fi.Sys().(*syscall.Stat_t).Blocks
	if blocks > 0 {
		t.Errorf("file is not sparse %s blocks %d", dst, blocks)
	}
	if !t.Failed() {
		os.RemoveAll(dir)
	}

}
