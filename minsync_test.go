package main

import (
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
)

func TestSync(t *testing.T) {

	cases := []struct {
		size1   int
		size2   int
		changes float32
	}{
		{0, 0, 0},

		{1 * KB, 1 * KB, 0},
		{1 * KB, 1 * KB, 0.5},
		{1 * KB, 1 * KB, 1},
		{0 * KB, 1 * KB, 0},
		{0 * KB, 1 * KB, 0.5},
		{0 * KB, 1 * KB, 1},
		{1 * KB, 0 * KB, 0},
		{0 * KB, 1 * KB, 0.5},
		{1 * KB, 0 * KB, 1},

		{4 * KB, 4 * KB, 0},
		{4 * KB, 4 * KB, 0.5},
		{4 * KB, 4 * KB, 1},
		{0 * KB, 4 * KB, 0},
		{0 * KB, 4 * KB, 0.5},
		{0 * KB, 4 * KB, 1},
		{4 * KB, 0 * KB, 0},
		{4 * KB, 0 * KB, 0.5},
		{4 * KB, 0 * KB, 1},

		{1 * MB, 1 * MB, 0.1},
		{1 * MB, 2 * MB, 0.1},
		{2 * MB, 1 * MB, 0.1},
		{1 * MB, 1 * MB, 0.5},
		{1 * MB, 2 * MB, 0.5},
		{2 * MB, 1 * MB, 0.5},
		{1 * MB, 1 * MB, 1.0},
		{1 * MB, 2 * MB, 1.0},
		{2 * MB, 1 * MB, 1.0},

		{MB, MB + KB, 0.5},
		{MB + KB, MB, 0.5},
		{MB, MB + KB, 1.0},
		{MB + KB, MB, 1.0},
	}

	for _, c := range cases {
		dir, src, dst, err := files(c.size1, c.size2, c.changes)
		t.Log(dir, c)
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dir)
		err = Sync(src, dst)
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
	}
}

func BenchmarkSync(b *testing.B) {
	for i := 0; i < b.N; i++ {
		dir, src, dst, err := files(128*MB, 128*MB, 0.1)
		if err != nil {
			b.Fatal(err)
		}
		defer os.RemoveAll(dir)
		b.StartTimer()
		err = Sync(src, dst)
		b.StopTimer()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func files(size1, size2 int, changes float32) (string, string, string, error) {
	temp, err := ioutil.TempDir("", "minsync_")
	if err != nil {
		return "", "", "", err
	}

	data := randomBytes(size1)

	src := path.Join(temp, "src")
	err = ioutil.WriteFile(src, data, os.ModePerm)
	if err != nil {
		return "", "", "", err
	}

	if size1 <= size2 {
		rest := randomBytes(size2 - size1)
		data = append(data, rest...)
	} else {
		data = data[:size2]
	}

	count := int(changes * float32(len(data)) / float32(BLOCK_SIZE))
	for i := 0; i < count; i++ {
		index := rand.Intn(len(data))
		data[index] = byte(-data[index])
	}

	dst := path.Join(temp, "dst")
	err = ioutil.WriteFile(dst, data, os.ModePerm)
	if err != nil {
		return "", "", "", err
	}

	return temp, src, dst, err
}

func randomBytes(size int) []byte {
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = byte(rand.Intn(256) - 128)
	}
	return b
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
		if !Compare(b1[:n1], b2[:n2]) {
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

func TestCompare(t *testing.T) {

	cases := []struct {
		a, b string
	}{
		{"", ""},
		{"a", "a"},
		{"ab", "a"},
		{"abc", "aBc"},
	}

	for _, c := range cases {
		r1 := Compare([]byte(c.a), []byte(c.b))
		r2 := Compare([]byte(c.b), []byte(c.a))
		if r1 != r2 {
			t.Errorf("%t != %t", r1, r2)
		}
		e := c.a == c.b
		if r1 != e {
			t.Errorf("%t != %t", r1, e)
		}
	}

}
