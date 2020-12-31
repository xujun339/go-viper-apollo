package file_helper

import (
	"io/ioutil"
	"os"
)

type WrapFile struct {
	fsize int64
	fp    *os.File
}

func (w *WrapFile) Size() int64 {
	return w.fsize
}

func (w *WrapFile) Fp() *os.File {
	return w.fp
}

func (w *WrapFile) Write(p []byte) (n int, err error) {
	n, err = w.fp.Write(p)
	w.fsize += int64(n)
	return
}

func NewWrapFile(fpath string) (*WrapFile, error) {
	fp, err := os.OpenFile(fpath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}
	fi, err := fp.Stat()
	if err != nil {
		return nil, err
	}
	return &WrapFile{fp: fp, fsize: fi.Size()}, nil
}

func (w *WrapFile) Close() error {
	return w.fp.Close()
}

func ReadFile(fpath string) ([]byte, error)  {
	return ioutil.ReadFile(fpath)
}