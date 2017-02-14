package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"
)

func TestSync(t *testing.T) {

	type testCase struct {
		src     int64
		dst     int64
		zeros   float32
		changes float32
	}
	var cases []testCase

	sizes := []int64{0, KB, KB + BLOCK_SIZE, MB, MB + KB, MB + BLOCK_SIZE, BLOCK_SIZE, BLOCK_SIZE + 1}
	probs := []float32{0, 0.33, 0.5, 0.66, 0.9, 1}

	for _, s1 := range sizes {
		for _, s2 := range sizes {
			for _, z := range probs {
				for _, c := range probs {
					cases = append(cases, testCase{s1, s2, z, c})
				}
			}
		}
	}

	for _, c := range cases {
		t.Run(
			fmt.Sprintf("case %d %d %.2f %.2f", c.src, c.dst, c.zeros, c.changes),
			func(t *testing.T) {
				dir, src, dst, err := randomFiles(c.src, c.dst, c.zeros, c.changes)
				t.Log("directory", dir)
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
					t.Errorf("files differ")
				}
				if !t.Failed() {
					os.RemoveAll(dir)
				}
			})
	}

}

func BenchmarkSync(b *testing.B) {

	for i := 0; i < b.N; i++ {
		dir, src, dst, err := randomFiles(MB, MB, 0.3, 0.1)
		if err != nil {
			b.Fatal(err)
		}
		defer os.RemoveAll(dir)
		b.StartTimer()
		p := NewProgress(ioutil.Discard)
		err = Sync(src, dst, p)
		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
	}

}

func randomFiles(ssize, dsize int64, zeros, changes float32) (string, string, string, error) {

	rand.Seed(time.Now().UnixNano())

	dir, err := ioutil.TempDir("", "minsync_")
	if err != nil {
		return "", "", "", err
	}

	sname := path.Join(dir, "src")
	src, err := os.Create(sname)
	if err != nil {
		return "", "", "", err
	}
	err = src.Truncate(ssize)
	if err != nil {
		return "", "", "", err
	}

	dname := path.Join(dir, "dst")
	dst, err := os.Create(dname)
	if err != nil {
		return "", "", "", err
	}
	err = dst.Truncate(dsize)
	if err != nil {
		return "", "", "", err
	}

	b := make([]byte, BLOCK_SIZE)

	for offset := int64(0); offset < ssize || offset < dsize; offset += BLOCK_SIZE {

		if rand.Float32() < zeros {
			for i := range b {
				b[i] = 0
			}
		} else {
			rand.Read(b)
		}

		if offset < ssize {
			n := ssize - offset
			if n > int64(len(b)) {
				n = int64(len(b))
			}
			_, err = src.WriteAt(b[:n], offset)
			if err != nil {
				return "", "", "", err
			}
		}

		if offset < dsize {
			n := dsize - offset
			if n > int64(len(b)) {
				n = int64(len(b))
			}
			if rand.Float32() < changes {
				i := rand.Intn(int(n))
				b[i] += 1
			}
			_, err = dst.WriteAt(b[:n], offset)
			if err != nil {
				return "", "", "", err
			}
		}

	}

	return dir, sname, dname, err

}

func compareFiles(n1, n2 string) (bool, error) {

	f1, err := os.Open(n1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(n2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	return compareContents(f1, f2)

}

func compareContents(r1, r2 io.Reader) (bool, error) {

	b1 := make([]byte, BLOCK_SIZE)
	b2 := make([]byte, BLOCK_SIZE)

	for {
		n1, err1 := r1.Read(b1)
		n2, err2 := r2.Read(b2)
		if n1 != n2 {
			return false, nil
		}
		if !bytes.Equal(b1[:n1], b2[:n2]) {
			return false, nil
		}
		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}
		if err1 != nil {
			return false, err1
		}
		if err2 != nil {
			return false, err2
		}
	}

}
