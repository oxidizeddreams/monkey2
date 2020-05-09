package op

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Worktree struct {
	baseDir string
}

func (w Worktree) NewFile(name, text string) {
	path := w.completePath(name)
	if err := ioutil.WriteFile(path, []byte(text), 0755); err != nil {
		w.panic(path, err)
	}
}

func (w Worktree) OverrideFile(name, text string, off, size int64) {
	path := w.completePath(name)
	f, err := os.OpenFile(path, os.O_RDWR, 0755)
	if err != nil {
		w.panic(path, err)
	}

	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		w.panic(path, err)
	}

	overriddenLen := fi.Size() - off
	if overriddenLen < 0 {
		w.panic(path, fmt.Errorf("size: %d, offset: %d", fi.Size(), off))
	}

	var overriddenBuf []byte
	if overriddenLen > 0 {
		overriddenBuf = make([]byte, 0, overriddenLen)
		offset := off
		var n int64
		var err error
		for n < overriddenLen && err != io.EOF {
			var m int
			m, err = f.ReadAt(overriddenBuf[n:], offset)
			if m == 0 {
				w.panic(path, err)
			}
			n += int64(m)
			offset += int64(m)
		}
	}

	if len(text) > 0 {
		buf := []byte(text)
		offset := off
		var n int64
		var err error
		for n < int64(len(text)) {
			var m int
			m, err = f.WriteAt(buf[n:], offset)
			if m == 0 {
				w.panic(path, err)
			}

			n += int64(m)
			offset += int64(m)
		}
	}

	if int64(len(overriddenBuf)) > size {
		buf := overriddenBuf[size:]
		var n int64
		var err error
		for n < int64(len(buf)) {
			var m int
			m, err = f.Write(buf[n:])
			if m == 0 {
				w.panic(path, err)
			}

			n += int64(m)
		}
	}
}

func (w Worktree) NewDir(name string) {
	path := w.completePath(name)
	if err := os.MkdirAll(path, 0755); err != nil {
		w.panic(path, err)
	}
}

func (w Worktree) Delete(name string) {
	path := w.completePath(name)
	if err := os.RemoveAll(path); err != nil {
		w.panic(path, err)
	}
}

func (w Worktree) Rename(origin, target string) {
	originPath := w.completePath(origin)
	targetPath := w.completePath(target)
	if err := os.Rename(originPath, targetPath); err != nil {
		w.panic(originPath, err)
	}
}

func (w Worktree) completePath(name string) (path string) {
	path = filepath.Join(w.baseDir, name)
	if _, err := os.Lstat(path); !os.IsNotExist(err) {
		panic(path)
	}

	return
}

func (w Worktree) panic(path string, err error) {
	panic(fmt.Sprintf("%s:%s", path, err))
}